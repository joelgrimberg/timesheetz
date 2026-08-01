package sync

import (
	"database/sql"
	"fmt"
	"os"

	"timesheet/internal/db"
	"timesheet/internal/logging"
)

// envSyncMode returns the raw TIMESHEETZ_SYNC_MODE env value.
// A tiny wrapper so tests can override it if we ever need to.
func envSyncMode() string { return os.Getenv("TIMESHEETZ_SYNC_MODE") }

// syncMode selects between the full-scan reconcile (Full) and the
// cursor-guarded skip fast-path (Incremental). Both modes produce
// identical converged state; Incremental just avoids full-table reads
// when max(updated_at) hasn't advanced since the previous sync.
//
// Default is Incremental. Set TIMESHEETZ_SYNC_MODE=full to fall back
// per-release-cycle while we build confidence in the incremental path.
type syncMode int

const (
	syncModeIncremental syncMode = iota
	syncModeFull
)

// getMaxUpdatedAt returns the greatest updated_at value in the given
// table, or the empty string if the table is empty. Used by the
// incremental fast-path to decide whether the full sync can be skipped.
func getMaxUpdatedAt(conn *sql.DB, dialect, table string) (string, error) {
	// Same query in both dialects; no placeholders needed.
	q := fmt.Sprintf(`SELECT COALESCE(MAX(updated_at), '') FROM %s`, table)
	var maxAt string
	if err := conn.QueryRow(q).Scan(&maxAt); err != nil {
		return "", fmt.Errorf("max(updated_at) for %s: %w", table, err)
	}
	return maxAt, nil
}

// getMaxTombstoneDeletedAt returns the greatest deleted_at for the given
// table's tombstones. Used together with getMaxUpdatedAt to decide if a
// sync cycle can be skipped: if neither rows nor tombstones have moved
// on either side, the reconcile has nothing to do.
func getMaxTombstoneDeletedAt(conn *sql.DB, dialect, table string) (string, error) {
	var q string
	switch dialect {
	case "postgres":
		q = `SELECT COALESCE(MAX(deleted_at), '') FROM tombstones WHERE table_name = $1`
	default:
		q = `SELECT COALESCE(MAX(deleted_at), '') FROM tombstones WHERE table_name = ?`
	}
	var maxAt string
	if err := conn.QueryRow(q, table).Scan(&maxAt); err != nil {
		return "", fmt.Errorf("max(deleted_at) for %s: %w", table, err)
	}
	return maxAt, nil
}

// shouldSkipTableSync returns (true, nil) when the sync loop can safely
// skip the full reconcile for the given table: neither side's data has
// changed since the last recorded cursor (with a clock-skew overlap).
//
// This is the incremental fast-path. It replaces a full-table scan on
// both sides with two tiny COUNT/MAX queries — a big win for the idle
// case, which is by far the common case (users don't edit continuously).
func (s *SyncService) shouldSkipTableSync(table string) (bool, error) {
	cursor, err := db.GetSyncCursor(s.localDB, "sqlite", table)
	if err != nil {
		return false, err
	}
	// A zero cursor means we've never successfully synced this table —
	// always do a full scan the first time so both sides converge.
	if cursor.Local == "" || cursor.Remote == "" {
		return false, nil
	}

	// Rows.
	localMax, err := getMaxUpdatedAt(s.localDB, "sqlite", table)
	if err != nil {
		return false, err
	}
	remoteMax, err := getMaxUpdatedAt(s.remoteDB, "postgres", table)
	if err != nil {
		return false, err
	}

	// Tombstones.
	localTombMax, err := getMaxTombstoneDeletedAt(s.localDB, "sqlite", table)
	if err != nil {
		return false, err
	}
	remoteTombMax, err := getMaxTombstoneDeletedAt(s.remoteDB, "postgres", table)
	if err != nil {
		return false, err
	}

	// The skip probe compares within one DB (localMax vs cursor.Local,
	// remoteMax vs cursor.Remote), so no cross-DB clock skew is
	// involved and no overlap window is needed. If anything on either
	// side advanced strictly beyond its cursor, there's work to do.
	if localMax > cursor.Local || remoteMax > cursor.Remote ||
		localTombMax > cursor.Local || remoteTombMax > cursor.Remote {
		return false, nil
	}

	logging.Log("Skipping %s sync — no changes since last cursor", table)
	return true, nil
}

// advanceCursorAfterSync records the new max updated_at values as the
// cursor for the next cycle. Called only after a table sync completes
// without error; on failure, the cursor is left in place so the next
// cycle retries the same range.
//
// The recorded cursor is the max of the current DB values *and* the
// prior cursor, so a table with no rows keeps the cursor from moving
// backwards to "".
func (s *SyncService) advanceCursorAfterSync(table string) error {
	prior, err := db.GetSyncCursor(s.localDB, "sqlite", table)
	if err != nil {
		return err
	}

	localMax, err := getMaxUpdatedAt(s.localDB, "sqlite", table)
	if err != nil {
		return err
	}
	remoteMax, err := getMaxUpdatedAt(s.remoteDB, "postgres", table)
	if err != nil {
		return err
	}

	// Include tombstone deleted_at in the cursor so a delete on side A
	// without any row change still bumps the cursor.
	localTombMax, _ := getMaxTombstoneDeletedAt(s.localDB, "sqlite", table)
	remoteTombMax, _ := getMaxTombstoneDeletedAt(s.remoteDB, "postgres", table)

	next := db.SyncCursor{
		Local:  db.MaxCursor(db.MaxCursor(prior.Local, localMax), localTombMax),
		Remote: db.MaxCursor(db.MaxCursor(prior.Remote, remoteMax), remoteTombMax),
	}
	return db.SetSyncCursor(s.localDB, "sqlite", table, next)
}
