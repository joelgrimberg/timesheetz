package sync

import (
	"testing"
	"time"

	"timesheet/internal/db"
)

// TestIncrementalSync_SkipsIdleCycle covers the core payoff of the
// incremental fast-path: on a second sync where nothing has changed,
// the sync must not touch the timesheet table at all.
//
// We assert this by draining the row from the local side AFTER the
// first sync but BEFORE the second one, without bumping updated_at. If
// the second sync had re-scanned the full table, it would notice the
// missing row and push it back down. With the fast-path skipping the
// scan (because both cursors say "no changes"), the local row stays
// gone — proving the reconcile did not run.
func TestIncrementalSync_SkipsIdleCycle(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)

	// First sync: cursor advances, both sides converge.
	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Simulate a sneaky local delete that bypasses updated_at (like a
	// manual DB edit outside timesheetz). A full sync would push the
	// remote row back to local; an incremental fast-path would skip
	// because cursors say nothing changed.
	if _, err := localDB.Exec(`DELETE FROM timesheet WHERE date = ?`, date); err != nil {
		t.Fatalf("stealth delete: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	// Local row should still be missing — proof the second sync skipped
	// the reconcile for the timesheet table.
	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("incremental sync unexpectedly re-scanned and restored local row (found %d)", got)
	}
}

// TestIncrementalSync_ProcessesChangedRow: when a row IS updated on one
// side, the incremental probe must NOT skip, and the change must
// propagate normally.
func TestIncrementalSync_ProcessesChangedRow(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Update the remote row (new updated_at). The cursor probe should
	// see remoteMax > cursor.Remote and fall through to the full sync.
	if _, err := remoteDB.Exec(
		`UPDATE timesheet SET client_hours = 12, updated_at = $1 WHERE date = $2`,
		t1, date,
	); err != nil {
		t.Fatalf("bump remote: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	// The updated hours must have propagated to local.
	var hours int
	if err := localDB.QueryRow(
		`SELECT client_hours FROM timesheet WHERE date = ?`, date,
	).Scan(&hours); err != nil {
		t.Fatalf("read local: %v", err)
	}
	if hours != 12 {
		t.Errorf("expected local hours to be updated to 12 (from remote), got %d", hours)
	}
}

// TestIncrementalSync_FirstSyncIsFullScan: with no cursor persisted
// yet, the probe must return "don't skip" so the initial reconcile runs
// unconditionally. Otherwise an initial sync would produce no
// convergence.
func TestIncrementalSync_FirstSyncIsFullScan(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"

	// Only remote has the row. First-ever sync must pull it into local.
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 1 {
		t.Errorf("first sync failed to pull remote row (found %d)", got)
	}
}

// TestIncrementalSync_TombstoneWakesUpSkip: even if row updated_at
// hasn't advanced, a fresh tombstone on either side means there IS work
// to do (the tombstone must propagate). The probe must include the
// max(deleted_at) check.
func TestIncrementalSync_TombstoneWakesUpSkip(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)
	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Delete on remote AND write a fresh tombstone. Row updated_at
	// didn't move (row is gone), but the tombstone did.
	writeTombstone(t, remoteDB, "postgres", db.TombstoneTableTimesheet, date, t1)
	if _, err := remoteDB.Exec(`DELETE FROM timesheet WHERE date = $1`, date); err != nil {
		t.Fatalf("delete remote: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	// Local row must be gone (tombstone propagated).
	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("tombstone did not propagate to local (found %d rows)", got)
	}
}

// TestPruneTombstones_RespectsRetention: tombstones older than the
// retention window are removed; younger ones stay.
func TestPruneTombstones_RespectsRetention(t *testing.T) {
	svc, localDB, _ := newSyncPair(t)

	// Fresh tombstone (today) — must survive.
	fresh := db.FormatTimestamp(time.Now().UTC())
	writeTombstone(t, localDB, "sqlite", db.TombstoneTableTimesheet, "fresh", fresh)

	// Ancient tombstone (100 days ago) — must be pruned with a 30-day
	// retention.
	ancient := db.FormatTimestamp(time.Now().UTC().Add(-100 * 24 * time.Hour))
	writeTombstone(t, localDB, "sqlite", db.TombstoneTableTimesheet, "ancient", ancient)

	pruned, _, err := svc.PruneTombstones(30 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 local pruned, got %d", pruned)
	}

	if got := countTombstones(t, localDB, db.TombstoneTableTimesheet, "fresh"); got != 1 {
		t.Errorf("fresh tombstone should be kept, found %d", got)
	}
	if got := countTombstones(t, localDB, db.TombstoneTableTimesheet, "ancient"); got != 0 {
		t.Errorf("ancient tombstone should be pruned, found %d", got)
	}
}

// TestSyncMode_EnvFullDisablesIncremental: setting TIMESHEETZ_SYNC_MODE
// to "full" opts out of the fast-path. In that mode, the "stealth
// delete" test case above should FAIL to preserve the row (proving the
// full scan happened).
func TestSyncMode_EnvFullDisablesIncremental(t *testing.T) {
	t.Setenv("TIMESHEETZ_SYNC_MODE", "full")
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)
	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Stealth delete on local — a full sync WILL notice and repush.
	if _, err := localDB.Exec(`DELETE FROM timesheet WHERE date = ?`, date); err != nil {
		t.Fatalf("stealth: %v", err)
	}
	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 1 {
		t.Errorf("full-mode sync should have restored local row from remote, found %d", got)
	}
}
