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
