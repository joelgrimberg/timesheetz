package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SyncCursor holds the two per-table cursors that the sync loop advances
// on every successful cycle: how far it has processed rows on the local
// and remote sides respectively. Cursors are lexically-sortable
// timestamp strings (the same format written into updated_at columns by
// NowTimestamp).
type SyncCursor struct {
	Local  string
	Remote string
}

// GetSyncCursor returns the persisted cursor for the given table, or a
// zero-value cursor (empty strings) if none has been stored yet. The
// zero value naturally makes the first sync a full-table scan because
// `updated_at > ''` matches every row.
//
// dialect must be either "sqlite" or "postgres" so the correct
// placeholder style is used.
func GetSyncCursor(conn *sql.DB, dialect, table string) (SyncCursor, error) {
	var q string
	switch dialect {
	case "postgres":
		q = `SELECT local_cursor, remote_cursor FROM sync_state WHERE table_name = $1`
	default:
		q = `SELECT local_cursor, remote_cursor FROM sync_state WHERE table_name = ?`
	}

	var c SyncCursor
	err := conn.QueryRow(q, table).Scan(&c.Local, &c.Remote)
	if err == sql.ErrNoRows {
		return SyncCursor{}, nil
	}
	if err != nil {
		return SyncCursor{}, fmt.Errorf("get sync cursor for %s: %w", table, err)
	}
	return c, nil
}

// SetSyncCursor persists the cursor for the given table. Called at the
// end of a successful table sync; if the sync fails partway through, the
// cursor is not advanced and the next cycle retries the same rows.
func SetSyncCursor(conn *sql.DB, dialect, table string, c SyncCursor) error {
	now := NowTimestamp()
	var q string
	switch dialect {
	case "postgres":
		q = `INSERT INTO sync_state (table_name, local_cursor, remote_cursor, updated_at)
		     VALUES ($1, $2, $3, $4)
		     ON CONFLICT (table_name) DO UPDATE SET
		       local_cursor = EXCLUDED.local_cursor,
		       remote_cursor = EXCLUDED.remote_cursor,
		       updated_at = EXCLUDED.updated_at`
	default:
		q = `INSERT INTO sync_state (table_name, local_cursor, remote_cursor, updated_at)
		     VALUES (?, ?, ?, ?)
		     ON CONFLICT(table_name) DO UPDATE SET
		       local_cursor = excluded.local_cursor,
		       remote_cursor = excluded.remote_cursor,
		       updated_at = excluded.updated_at`
	}

	if _, err := conn.Exec(q, table, c.Local, c.Remote, now); err != nil {
		return fmt.Errorf("set sync cursor for %s: %w", table, err)
	}
	return nil
}

// MaxCursor returns the lexically greater of two cursor strings. Cursors
// are timestamp strings in "YYYY-MM-DD HH:MM:SS" form, so string
// comparison matches chronological order.
func MaxCursor(a, b string) string {
	if a > b {
		return a
	}
	return b
}

// ApplyCursorOverlap subtracts the given overlap window from a cursor to
// build the lower bound of an incremental read. The overlap absorbs
// clock skew between the laptop and a remote Postgres server, and
// catches rows whose commit visibility trailed their updated_at value.
//
// Returns an empty string when the cursor is empty (which the readers
// naturally interpret as "no lower bound"). Rows written before the
// overlap window are captured again on the next cycle; because the
// reconcile logic is idempotent (compares updated_at both sides), this
// is safe.
func ApplyCursorOverlap(cursor string, overlap time.Duration) string {
	if cursor == "" {
		return ""
	}
	t, err := ParseTimestamp(cursor)
	if err != nil {
		return cursor // fall back to the exact cursor on unexpected format
	}
	return FormatTimestamp(t.Add(-overlap))
}
