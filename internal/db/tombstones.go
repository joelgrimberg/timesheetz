package db

import (
	"database/sql"
	"fmt"
)

// Tombstone table names — kept in sync with the keys used by the sync
// package when reconciling rows across SQLite and PostgreSQL.
const (
	TombstoneTableTimesheet         = "timesheet"
	TombstoneTableClients           = "clients"
	TombstoneTableClientRates       = "client_rates"
	TombstoneTableTrainingBudget    = "training_budget"
	TombstoneTableVacationCarryover = "vacation_carryover"
	TombstoneTableBufferHours       = "buffer_hours"
)

// TombstoneKeyClientRate, TombstoneKeyTrainingBudget,
// TombstoneKeyVacationCarryover, and TombstoneKeyBufferHours encode a row's
// natural sync key as a string. These MUST match the keys the sync package
// builds when mapping rows side-to-side, otherwise tombstones won't line up
// with the rows they're supposed to bury.

func TombstoneKeyClientRate(clientName, effectiveDate string) string {
	return clientName + "|" + effectiveDate
}

func TombstoneKeyTrainingBudget(date, trainingName string) string {
	return date + "|" + trainingName
}

func TombstoneKeyVacationCarryover(year int) string {
	return fmt.Sprintf("%d", year)
}

func TombstoneKeyBufferHours(year, month int) string {
	return fmt.Sprintf("%d-%02d", year, month)
}

// sqlExecer matches the subset of *sql.DB and *sql.Tx we need so tombstone
// writers can run either standalone or inside a caller-owned transaction.
type sqlExecer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// WriteSqliteTombstone upserts a tombstone row using SQLite syntax.
func WriteSqliteTombstone(ex sqlExecer, table, key string) error {
	_, err := ex.Exec(
		`INSERT OR REPLACE INTO tombstones (table_name, record_key, deleted_at) VALUES (?, ?, ?)`,
		table, key, NowTimestamp(),
	)
	if err != nil {
		return fmt.Errorf("failed to write tombstone for %s/%s: %w", table, key, err)
	}
	return nil
}

// WritePostgresTombstone upserts a tombstone row using PostgreSQL syntax.
func WritePostgresTombstone(ex sqlExecer, table, key string) error {
	_, err := ex.Exec(
		`INSERT INTO tombstones (table_name, record_key, deleted_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (table_name, record_key) DO UPDATE SET deleted_at = EXCLUDED.deleted_at`,
		table, key, NowTimestamp(),
	)
	if err != nil {
		return fmt.Errorf("failed to write tombstone for %s/%s: %w", table, key, err)
	}
	return nil
}

// PruneTombstones deletes tombstones whose deleted_at is older than
// (now - retention). Called at the end of a successful sync cycle so
// the tombstones table doesn't grow unbounded.
//
// The retention window must be larger than the longest realistic offline
// gap for any client — if a laptop syncs after being offline for longer
// than retention, it may re-insert a row we already forgot to keep buried.
// 30 days is the default (see sync.tombstoneRetention).
//
// olderThan should be the cutoff timestamp string
// (`FormatTimestamp(time.Now().UTC().Add(-retention))`).
func PruneTombstones(conn *sql.DB, dialect, olderThan string) (int64, error) {
	var q string
	switch dialect {
	case "postgres":
		q = `DELETE FROM tombstones WHERE deleted_at < $1`
	default:
		q = `DELETE FROM tombstones WHERE deleted_at < ?`
	}
	res, err := conn.Exec(q, olderThan)
	if err != nil {
		return 0, fmt.Errorf("prune tombstones (%s): %w", dialect, err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
