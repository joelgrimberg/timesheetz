package sync

import (
	"database/sql"
	"testing"
	"time"

	"timesheet/internal/db"

	_ "modernc.org/sqlite"
)

// newSyncPair returns a SyncService backed by two independent in-memory
// SQLite databases. The "remote" side stands in for PostgreSQL; the
// modernc.org/sqlite driver accepts both `?` and `$N` placeholders, so the
// sync code's Postgres-style remote INSERTs run unchanged.
func newSyncPair(t *testing.T) (*SyncService, *sql.DB, *sql.DB) {
	t.Helper()

	localDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open local: %v", err)
	}
	t.Cleanup(func() { localDB.Close() })
	if err := db.ApplySQLiteSchema(localDB); err != nil {
		t.Fatalf("init local schema: %v", err)
	}

	remoteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open remote: %v", err)
	}
	t.Cleanup(func() { remoteDB.Close() })
	if err := db.ApplySQLiteSchema(remoteDB); err != nil {
		t.Fatalf("init remote schema: %v", err)
	}

	return NewSyncService(localDB, remoteDB, time.Minute), localDB, remoteDB
}

// seedTimesheetRow inserts a timesheet row with explicit timestamps so tests
// can control the "newer wins" comparisons deterministically.
func seedTimesheetRow(t *testing.T, conn *sql.DB, dialect, date, updatedAt string) {
	t.Helper()
	var q string
	if dialect == "postgres" {
		q = `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, created_at, updated_at) VALUES ($1, $2, 8, 0, 0, 0, 0, 0, $3, $3)`
	} else {
		q = `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, created_at, updated_at) VALUES (?, ?, 8, 0, 0, 0, 0, 0, ?, ?)`
	}
	if dialect == "postgres" {
		_, err := conn.Exec(q, date, "Acme", updatedAt)
		if err != nil {
			t.Fatalf("seed remote timesheet: %v", err)
		}
	} else {
		_, err := conn.Exec(q, date, "Acme", updatedAt, updatedAt)
		if err != nil {
			t.Fatalf("seed local timesheet: %v", err)
		}
	}
}

// writeTombstone inserts a tombstone directly, bypassing the data layer's
// per-Delete path. Tests use this to simulate "a delete already happened on
// this side at time T" without needing the rest of the data layer wired up.
func writeTombstone(t *testing.T, conn *sql.DB, dialect, table, key, deletedAt string) {
	t.Helper()
	var q string
	if dialect == "postgres" {
		q = `INSERT INTO tombstones (table_name, record_key, deleted_at) VALUES ($1, $2, $3) ON CONFLICT (table_name, record_key) DO UPDATE SET deleted_at = EXCLUDED.deleted_at`
	} else {
		q = `INSERT OR REPLACE INTO tombstones (table_name, record_key, deleted_at) VALUES (?, ?, ?)`
	}
	if _, err := conn.Exec(q, table, key, deletedAt); err != nil {
		t.Fatalf("write tombstone: %v", err)
	}
}

func countTimesheetRows(t *testing.T, conn *sql.DB, date string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM timesheet WHERE date = ?`, date).Scan(&n); err != nil {
		t.Fatalf("count timesheet: %v", err)
	}
	return n
}

func countTombstones(t *testing.T, conn *sql.DB, table, key string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tombstones WHERE table_name = ? AND record_key = ?`, table, key).Scan(&n); err != nil {
		t.Fatalf("count tombstones: %v", err)
	}
	return n
}

// TestSync_DeletePropagatesFromRemoteToLocal is the bug scenario the user
// originally reported: the row was deleted from the side the UI wrote to
// (Postgres / remote), the other side still has it, and without tombstones
// the next sync re-inserts it.
func TestSync_DeletePropagatesFromRemoteToLocal(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)

	// Simulate "remote deleted, tombstone written on remote at t1".
	writeTombstone(t, remoteDB, "postgres", db.TombstoneTableTimesheet, date, t1)
	if _, err := remoteDB.Exec(`DELETE FROM timesheet WHERE date = $1`, date); err != nil {
		t.Fatalf("delete remote row: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("local row should be deleted after sync, found %d", got)
	}
	if got := countTimesheetRows(t, remoteDB, date); got != 0 {
		t.Errorf("remote row should stay deleted, found %d", got)
	}
	if got := countTombstones(t, localDB, db.TombstoneTableTimesheet, date); got != 1 {
		t.Errorf("expected tombstone propagated to local, found %d", got)
	}
	if got := countTombstones(t, remoteDB, db.TombstoneTableTimesheet, date); got != 1 {
		t.Errorf("expected tombstone still on remote, found %d", got)
	}
}

func TestSync_DeletePropagatesFromLocalToRemote(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)

	writeTombstone(t, localDB, "sqlite", db.TombstoneTableTimesheet, date, t1)
	if _, err := localDB.Exec(`DELETE FROM timesheet WHERE date = ?`, date); err != nil {
		t.Fatalf("delete local row: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("local row should stay deleted, found %d", got)
	}
	if got := countTimesheetRows(t, remoteDB, date); got != 0 {
		t.Errorf("remote row should be deleted after sync, found %d", got)
	}
	if got := countTombstones(t, remoteDB, db.TombstoneTableTimesheet, date); got != 1 {
		t.Errorf("expected tombstone propagated to remote, found %d", got)
	}
}

// TestSync_EditBeatsDelete: when one side has a tombstone but the other
// side's row has been updated AFTER the tombstone, the edit wins. The
// tombstone is dropped and the row is restored on the deleted side.
func TestSync_EditBeatsDelete(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05" // local delete time
	const t2 = "2026-06-14 10:00:10" // remote edit time, AFTER local delete

	// Local already deleted at t1 and only has the tombstone now.
	writeTombstone(t, localDB, "sqlite", db.TombstoneTableTimesheet, date, t1)
	// Remote edited the row at t2 — a write more recent than the delete.
	seedTimesheetRow(t, remoteDB, "postgres", date, t2)

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 1 {
		t.Errorf("expected edit to restore local row, found %d", got)
	}
	if got := countTimesheetRows(t, remoteDB, date); got != 1 {
		t.Errorf("expected remote row preserved, found %d", got)
	}
	if got := countTombstones(t, localDB, db.TombstoneTableTimesheet, date); got != 0 {
		t.Errorf("losing local tombstone should be dropped, found %d", got)
	}
	if got := countTombstones(t, remoteDB, db.TombstoneTableTimesheet, date); got != 0 {
		t.Errorf("remote should have no tombstone, found %d", got)
	}
}

// TestSync_RepeatedSyncConverges: after a delete propagates, running the
// sync again should be a no-op — no re-inserts, no stat counts.
func TestSync_RepeatedSyncConverges(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	seedTimesheetRow(t, remoteDB, "postgres", date, t0)
	writeTombstone(t, remoteDB, "postgres", db.TombstoneTableTimesheet, date, t1)
	if _, err := remoteDB.Exec(`DELETE FROM timesheet WHERE date = $1`, date); err != nil {
		t.Fatalf("delete remote row: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("first sync: %v", err)
	}
	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	stats := svc.GetLastSyncStats()
	if stats.RecordsPushed != 0 || stats.RecordsPulled != 0 {
		t.Errorf("second sync should be a no-op; got pushed=%d pulled=%d", stats.RecordsPushed, stats.RecordsPulled)
	}
	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("row should still be deleted on local after second sync, found %d", got)
	}
	if got := countTimesheetRows(t, remoteDB, date); got != 0 {
		t.Errorf("row should still be deleted on remote after second sync, found %d", got)
	}
}

// TestSync_DeleteOnlyOnOneSide_RowOnlyOnOther: simplest possible case —
// remote has a tombstone, local has a fresh row at the same key with an
// older updated_at. The delete should win and remove the local row.
func TestSync_DeleteOnlyOnOneSide_RowOnlyOnOther(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const date = "2026-06-14"
	const t0 = "2026-06-14 10:00:00"
	const t1 = "2026-06-14 10:00:05"

	// Local has the row at t0; remote has only a tombstone at t1 (later).
	seedTimesheetRow(t, localDB, "sqlite", date, t0)
	writeTombstone(t, remoteDB, "postgres", db.TombstoneTableTimesheet, date, t1)

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if got := countTimesheetRows(t, localDB, date); got != 0 {
		t.Errorf("local row should be deleted, found %d", got)
	}
	if got := countTombstones(t, localDB, db.TombstoneTableTimesheet, date); got != 1 {
		t.Errorf("tombstone should be on local, found %d", got)
	}
}

// TestSync_BufferDeletePropagates exercises a different table (buffer_hours)
// to make sure the wiring is consistent across the six syncs, not just
// timesheet-specific.
func TestSync_BufferDeletePropagates(t *testing.T) {
	svc, localDB, remoteDB := newSyncPair(t)

	const (
		year     = 2026
		month    = 6
		hours    = 4
		t0       = "2026-06-14 10:00:00"
		t1       = "2026-06-14 10:00:05"
		notesVal = ""
	)
	key := db.TombstoneKeyBufferHours(year, month)

	insertBuffer := func(conn *sql.DB, dialect string) {
		t.Helper()
		var q string
		if dialect == "postgres" {
			q = `INSERT INTO buffer_hours (year, month, hours, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $5)`
		} else {
			q = `INSERT INTO buffer_hours (year, month, hours, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
		}
		if dialect == "postgres" {
			if _, err := conn.Exec(q, year, month, hours, notesVal, t0); err != nil {
				t.Fatalf("seed remote buffer: %v", err)
			}
		} else {
			if _, err := conn.Exec(q, year, month, hours, notesVal, t0, t0); err != nil {
				t.Fatalf("seed local buffer: %v", err)
			}
		}
	}
	insertBuffer(localDB, "sqlite")
	insertBuffer(remoteDB, "postgres")

	// Delete on remote with a tombstone at t1.
	writeTombstone(t, remoteDB, "postgres", db.TombstoneTableBufferHours, key, t1)
	if _, err := remoteDB.Exec(`DELETE FROM buffer_hours WHERE year = $1 AND month = $2`, year, month); err != nil {
		t.Fatalf("delete remote buffer: %v", err)
	}

	if err := svc.Sync(SyncBidirectional); err != nil {
		t.Fatalf("sync: %v", err)
	}

	var localCount, remoteCount int
	if err := localDB.QueryRow(`SELECT COUNT(*) FROM buffer_hours WHERE year = ? AND month = ?`, year, month).Scan(&localCount); err != nil {
		t.Fatalf("count local buffer: %v", err)
	}
	if err := remoteDB.QueryRow(`SELECT COUNT(*) FROM buffer_hours WHERE year = ? AND month = ?`, year, month).Scan(&remoteCount); err != nil {
		t.Fatalf("count remote buffer: %v", err)
	}
	if localCount != 0 || remoteCount != 0 {
		t.Errorf("buffer row should be gone on both sides; local=%d remote=%d", localCount, remoteCount)
	}
	if got := countTombstones(t, localDB, db.TombstoneTableBufferHours, key); got != 1 {
		t.Errorf("expected tombstone propagated to local, found %d", got)
	}
}

