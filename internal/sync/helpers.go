package sync

import (
	"database/sql"
	"fmt"
	"time"

	"timesheet/internal/db"
)

// Internal record types with timestamps for sync
type clientRecord struct {
	Id        int
	Name      string
	CreatedAt string
	UpdatedAt string
	IsActive  int
}

type clientRateRecord struct {
	Id            int
	ClientId      int
	HourlyRate    float64
	EffectiveDate string
	Notes         string
	CreatedAt     string
	UpdatedAt     string
}

type timesheetRecord struct {
	Id            int
	Date          string
	ClientName    string
	ClientHours   sql.NullInt64
	VacationHours sql.NullInt64
	IdleHours     sql.NullInt64
	TrainingHours sql.NullInt64
	SickHours     sql.NullInt64
	HolidayHours  sql.NullInt64
	ClientId      sql.NullInt64
	CreatedAt     string
	UpdatedAt     string
}

type trainingBudgetRecord struct {
	Id             int
	Date           string
	TrainingName   string
	Hours          int
	CostWithoutVat float64
	CreatedAt      string
	UpdatedAt      string
}

// ============== Clients ==============

func (s *SyncService) getClientsFromDB(dbConn *sql.DB, dbType string) ([]clientRecord, error) {
	query := `SELECT id, name, COALESCE(created_at, ''), COALESCE(updated_at, ''), COALESCE(is_active, 1) FROM clients`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []clientRecord
	for rows.Next() {
		var c clientRecord
		if err := rows.Scan(&c.Id, &c.Name, &c.CreatedAt, &c.UpdatedAt, &c.IsActive); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func (s *SyncService) getClientIdMap(dbConn *sql.DB, dbType string) (map[string]int, error) {
	query := `SELECT id, name FROM clients`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[name] = id
	}
	return result, rows.Err()
}

func (s *SyncService) insertClientToRemote(c clientRecord) error {
	query := `INSERT INTO clients (name, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4)`
	_, err := s.remoteDB.Exec(query, c.Name, c.CreatedAt, c.UpdatedAt, c.IsActive)
	return err
}

func (s *SyncService) updateClientInRemote(c clientRecord, remoteId int) error {
	query := `UPDATE clients SET name = $1, updated_at = $2, is_active = $3 WHERE id = $4`
	_, err := s.remoteDB.Exec(query, c.Name, c.UpdatedAt, c.IsActive, remoteId)
	return err
}

func (s *SyncService) insertClientToLocal(c clientRecord) error {
	query := `INSERT INTO clients (name, created_at, updated_at, is_active) VALUES (?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, c.Name, c.CreatedAt, c.UpdatedAt, c.IsActive)
	return err
}

func (s *SyncService) updateClientInLocal(c clientRecord, localId int) error {
	query := `UPDATE clients SET name = ?, updated_at = ?, is_active = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, c.Name, c.UpdatedAt, c.IsActive, localId)
	return err
}

// ============== Client Rates ==============

func (s *SyncService) getClientRatesFromDB(dbConn *sql.DB, dbType string) ([]clientRateRecord, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, COALESCE(notes, ''), COALESCE(created_at, ''), COALESCE(updated_at, '') FROM client_rates`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []clientRateRecord
	for rows.Next() {
		var r clientRateRecord
		if err := rows.Scan(&r.Id, &r.ClientId, &r.HourlyRate, &r.EffectiveDate, &r.Notes, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}
	return rates, rows.Err()
}

func (s *SyncService) insertClientRateToRemote(r clientRateRecord, remoteClientId int) error {
	query := `INSERT INTO client_rates (client_id, hourly_rate, effective_date, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.remoteDB.Exec(query, remoteClientId, r.HourlyRate, r.EffectiveDate, r.Notes, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *SyncService) updateClientRateInRemote(r clientRateRecord, remoteId int, remoteClientId int) error {
	query := `UPDATE client_rates SET client_id = $1, hourly_rate = $2, effective_date = $3, notes = $4, updated_at = $5 WHERE id = $6`
	_, err := s.remoteDB.Exec(query, remoteClientId, r.HourlyRate, r.EffectiveDate, r.Notes, r.UpdatedAt, remoteId)
	return err
}

func (s *SyncService) insertClientRateToLocal(r clientRateRecord, localClientId int) error {
	query := `INSERT INTO client_rates (client_id, hourly_rate, effective_date, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, localClientId, r.HourlyRate, r.EffectiveDate, r.Notes, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *SyncService) updateClientRateInLocal(r clientRateRecord, localId int, localClientId int) error {
	query := `UPDATE client_rates SET client_id = ?, hourly_rate = ?, effective_date = ?, notes = ?, updated_at = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, localClientId, r.HourlyRate, r.EffectiveDate, r.Notes, r.UpdatedAt, localId)
	return err
}

// ============== Timesheet ==============

func (s *SyncService) getTimesheetFromDB(dbConn *sql.DB, dbType string) ([]timesheetRecord, error) {
	query := `SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, client_id, COALESCE(created_at, ''), COALESCE(updated_at, '') FROM timesheet`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []timesheetRecord
	for rows.Next() {
		var e timesheetRecord
		if err := rows.Scan(&e.Id, &e.Date, &e.ClientName, &e.ClientHours, &e.VacationHours, &e.IdleHours, &e.TrainingHours, &e.SickHours, &e.HolidayHours, &e.ClientId, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SyncService) insertTimesheetToRemote(e timesheetRecord) error {
	query := `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, client_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := s.remoteDB.Exec(query, e.Date, e.ClientName, e.ClientHours, e.VacationHours, e.IdleHours, e.TrainingHours, e.SickHours, e.HolidayHours, e.ClientId, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateTimesheetInRemote(e timesheetRecord, remoteId int) error {
	query := `UPDATE timesheet SET date = $1, client_name = $2, client_hours = $3, vacation_hours = $4, idle_hours = $5, training_hours = $6, sick_hours = $7, holiday_hours = $8, client_id = $9, updated_at = $10 WHERE id = $11`
	_, err := s.remoteDB.Exec(query, e.Date, e.ClientName, e.ClientHours, e.VacationHours, e.IdleHours, e.TrainingHours, e.SickHours, e.HolidayHours, e.ClientId, e.UpdatedAt, remoteId)
	return err
}

func (s *SyncService) insertTimesheetToLocal(e timesheetRecord) error {
	query := `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, client_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, e.Date, e.ClientName, e.ClientHours, e.VacationHours, e.IdleHours, e.TrainingHours, e.SickHours, e.HolidayHours, e.ClientId, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateTimesheetInLocal(e timesheetRecord, localId int) error {
	query := `UPDATE timesheet SET date = ?, client_name = ?, client_hours = ?, vacation_hours = ?, idle_hours = ?, training_hours = ?, sick_hours = ?, holiday_hours = ?, client_id = ?, updated_at = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, e.Date, e.ClientName, e.ClientHours, e.VacationHours, e.IdleHours, e.TrainingHours, e.SickHours, e.HolidayHours, e.ClientId, e.UpdatedAt, localId)
	return err
}

// ============== Training Budget ==============

func (s *SyncService) getTrainingBudgetFromDB(dbConn *sql.DB, dbType string) ([]trainingBudgetRecord, error) {
	query := `SELECT id, date, training_name, hours, cost_without_vat, COALESCE(created_at, ''), COALESCE(updated_at, '') FROM training_budget`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []trainingBudgetRecord
	for rows.Next() {
		var e trainingBudgetRecord
		if err := rows.Scan(&e.Id, &e.Date, &e.TrainingName, &e.Hours, &e.CostWithoutVat, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SyncService) insertTrainingBudgetToRemote(e trainingBudgetRecord) error {
	query := `INSERT INTO training_budget (date, training_name, hours, cost_without_vat, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.remoteDB.Exec(query, e.Date, e.TrainingName, e.Hours, e.CostWithoutVat, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateTrainingBudgetInRemote(e trainingBudgetRecord, remoteId int) error {
	query := `UPDATE training_budget SET date = $1, training_name = $2, hours = $3, cost_without_vat = $4, updated_at = $5 WHERE id = $6`
	_, err := s.remoteDB.Exec(query, e.Date, e.TrainingName, e.Hours, e.CostWithoutVat, e.UpdatedAt, remoteId)
	return err
}

func (s *SyncService) insertTrainingBudgetToLocal(e trainingBudgetRecord) error {
	query := `INSERT INTO training_budget (date, training_name, hours, cost_without_vat, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, e.Date, e.TrainingName, e.Hours, e.CostWithoutVat, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateTrainingBudgetInLocal(e trainingBudgetRecord, localId int) error {
	query := `UPDATE training_budget SET date = ?, training_name = ?, hours = ?, cost_without_vat = ?, updated_at = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, e.Date, e.TrainingName, e.Hours, e.CostWithoutVat, e.UpdatedAt, localId)
	return err
}

// ============== Vacation Carryover ==============

func (s *SyncService) getVacationCarryoverFromDB(dbConn *sql.DB, dbType string) ([]db.VacationCarryover, error) {
	query := `SELECT id, year, carryover_hours, source_year, COALESCE(created_at, ''), COALESCE(updated_at, ''), COALESCE(notes, '') FROM vacation_carryover`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []db.VacationCarryover
	for rows.Next() {
		var e db.VacationCarryover
		if err := rows.Scan(&e.Id, &e.Year, &e.CarryoverHours, &e.SourceYear, &e.CreatedAt, &e.UpdatedAt, &e.Notes); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SyncService) insertVacationCarryoverToRemote(e db.VacationCarryover) error {
	query := `INSERT INTO vacation_carryover (year, carryover_hours, source_year, created_at, updated_at, notes) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.remoteDB.Exec(query, e.Year, e.CarryoverHours, e.SourceYear, e.CreatedAt, e.UpdatedAt, e.Notes)
	return err
}

func (s *SyncService) updateVacationCarryoverInRemote(e db.VacationCarryover, remoteId int) error {
	query := `UPDATE vacation_carryover SET year = $1, carryover_hours = $2, source_year = $3, updated_at = $4, notes = $5 WHERE id = $6`
	_, err := s.remoteDB.Exec(query, e.Year, e.CarryoverHours, e.SourceYear, e.UpdatedAt, e.Notes, remoteId)
	return err
}

func (s *SyncService) insertVacationCarryoverToLocal(e db.VacationCarryover) error {
	query := `INSERT INTO vacation_carryover (year, carryover_hours, source_year, created_at, updated_at, notes) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, e.Year, e.CarryoverHours, e.SourceYear, e.CreatedAt, e.UpdatedAt, e.Notes)
	return err
}

func (s *SyncService) updateVacationCarryoverInLocal(e db.VacationCarryover, localId int) error {
	query := `UPDATE vacation_carryover SET year = ?, carryover_hours = ?, source_year = ?, updated_at = ?, notes = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, e.Year, e.CarryoverHours, e.SourceYear, e.UpdatedAt, e.Notes, localId)
	return err
}

// ============== Buffer Hours ==============

func (s *SyncService) getBufferHoursFromDB(dbConn *sql.DB, dbType string) ([]db.BufferEntry, error) {
	query := `SELECT id, year, month, hours, COALESCE(notes, ''), COALESCE(created_at, ''), COALESCE(updated_at, '') FROM buffer_hours`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []db.BufferEntry
	for rows.Next() {
		var e db.BufferEntry
		if err := rows.Scan(&e.Id, &e.Year, &e.Month, &e.Hours, &e.Notes, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *SyncService) insertBufferHoursToRemote(e db.BufferEntry) error {
	query := `INSERT INTO buffer_hours (year, month, hours, notes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.remoteDB.Exec(query, e.Year, e.Month, e.Hours, e.Notes, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateBufferHoursInRemote(e db.BufferEntry, remoteId int) error {
	query := `UPDATE buffer_hours SET year = $1, month = $2, hours = $3, notes = $4, updated_at = $5 WHERE id = $6`
	_, err := s.remoteDB.Exec(query, e.Year, e.Month, e.Hours, e.Notes, e.UpdatedAt, remoteId)
	return err
}

func (s *SyncService) insertBufferHoursToLocal(e db.BufferEntry) error {
	query := `INSERT INTO buffer_hours (year, month, hours, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.localDB.Exec(query, e.Year, e.Month, e.Hours, e.Notes, e.CreatedAt, e.UpdatedAt)
	return err
}

func (s *SyncService) updateBufferHoursInLocal(e db.BufferEntry, localId int) error {
	query := `UPDATE buffer_hours SET year = ?, month = ?, hours = ?, notes = ?, updated_at = ? WHERE id = ?`
	_, err := s.localDB.Exec(query, e.Year, e.Month, e.Hours, e.Notes, e.UpdatedAt, localId)
	return err
}

// ============== Tombstones ==============

// getTombstonesFromDB returns a map of record_key -> deleted_at timestamp
// for the given logical table name. Both SQLite and Postgres use the same
// schema for this table.
func (s *SyncService) getTombstonesFromDB(dbConn *sql.DB, dbType, tableName string) (map[string]string, error) {
	// Use the dialect's positional placeholder.
	var query string
	if dbType == "postgres" {
		query = `SELECT record_key, deleted_at FROM tombstones WHERE table_name = $1`
	} else {
		query = `SELECT record_key, deleted_at FROM tombstones WHERE table_name = ?`
	}
	rows, err := dbConn.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var key, deletedAt string
		if err := rows.Scan(&key, &deletedAt); err != nil {
			return nil, err
		}
		out[key] = deletedAt
	}
	return out, rows.Err()
}

func (s *SyncService) insertTombstoneToRemote(table, key, deletedAt string) error {
	_, err := s.remoteDB.Exec(
		`INSERT INTO tombstones (table_name, record_key, deleted_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (table_name, record_key) DO UPDATE SET deleted_at = EXCLUDED.deleted_at`,
		table, key, deletedAt,
	)
	return err
}

func (s *SyncService) insertTombstoneToLocal(table, key, deletedAt string) error {
	_, err := s.localDB.Exec(
		`INSERT OR REPLACE INTO tombstones (table_name, record_key, deleted_at) VALUES (?, ?, ?)`,
		table, key, deletedAt,
	)
	return err
}

func (s *SyncService) deleteTombstoneFromRemote(table, key string) error {
	_, err := s.remoteDB.Exec(`DELETE FROM tombstones WHERE table_name = $1 AND record_key = $2`, table, key)
	return err
}

func (s *SyncService) deleteTombstoneFromLocal(table, key string) error {
	_, err := s.localDB.Exec(`DELETE FROM tombstones WHERE table_name = ? AND record_key = ?`, table, key)
	return err
}

// tombstoneReconcileResult captures, after the tombstone pass, the set of
// keys that should be skipped by the subsequent upsert pass — these are
// rows that have lost a delete-vs-edit race or that have already been
// reconciled to "deleted on both sides".
type tombstoneReconcileResult struct {
	// killedKeys lists record keys the upsert pass MUST NOT re-insert.
	killedKeys map[string]struct{}
}

func newTombstoneReconcileResult() tombstoneReconcileResult {
	return tombstoneReconcileResult{killedKeys: make(map[string]struct{})}
}

func (r *tombstoneReconcileResult) kill(key string) {
	r.killedKeys[key] = struct{}{}
}

func (r tombstoneReconcileResult) isKilled(key string) bool {
	_, ok := r.killedKeys[key]
	return ok
}

// reconcileTombstones runs the delete-vs-edit reconciliation for a single
// logical table, given the current row state on both sides.
//
// rowUpdatedAt returns the updated_at string of a row identified by key on
// one side, plus whether it exists. deleteRow performs a hard delete of the
// row identified by key on the specified side (without writing a new
// tombstone — the propagated tombstone covers that).
//
// Semantics:
//   - If a tombstone on side A has deleted_at >= side B's row.updated_at,
//     the delete wins: row is hard-deleted from B, tombstone is propagated
//     to B, key is added to killedKeys so the upsert pass skips it.
//   - If side B's row.updated_at > tombstone.deleted_at, the edit wins:
//     the tombstone is dropped from A. The upsert pass will then push the
//     newer row from B back to A.
//   - If neither side has the row, the tombstone is simply propagated to
//     the side that doesn't have it yet.
func (s *SyncService) reconcileTombstones(
	table string,
	localTombstones, remoteTombstones map[string]string,
	localRowUpdatedAt, remoteRowUpdatedAt func(key string) (string, bool),
	deleteLocalRow, deleteRemoteRow func(key string) error,
) (tombstoneReconcileResult, error) {
	result := newTombstoneReconcileResult()

	// Union of keys with a tombstone on either side.
	keys := make(map[string]struct{}, len(localTombstones)+len(remoteTombstones))
	for k := range localTombstones {
		keys[k] = struct{}{}
	}
	for k := range remoteTombstones {
		keys[k] = struct{}{}
	}

	for key := range keys {
		localTs, hasLocalTs := localTombstones[key]
		remoteTs, hasRemoteTs := remoteTombstones[key]

		// Pick the most recent tombstone — that's the canonical delete time.
		ts := localTs
		if remoteTs > ts {
			ts = remoteTs
		}

		remoteUpdated, remoteHas := remoteRowUpdatedAt(key)
		localUpdated, localHas := localRowUpdatedAt(key)

		// Edit-beats-delete: if either side has a row with updated_at >
		// canonical tombstone deleted_at, the edit wins. Drop the
		// tombstone(s) and let the upsert pass propagate the live row.
		editBeatsDelete := (remoteHas && remoteUpdated > ts) || (localHas && localUpdated > ts)
		if editBeatsDelete {
			if hasLocalTs {
				if err := s.deleteTombstoneFromLocal(table, key); err != nil {
					return result, fmt.Errorf("drop losing local tombstone %s/%s: %w", table, key, err)
				}
			}
			if hasRemoteTs {
				if err := s.deleteTombstoneFromRemote(table, key); err != nil {
					return result, fmt.Errorf("drop losing remote tombstone %s/%s: %w", table, key, err)
				}
			}
			continue
		}

		// Delete wins. Hard-delete the row on whichever side still has it,
		// then make sure both sides carry the tombstone with the canonical
		// timestamp so future syncs don't reopen the race.
		if remoteHas {
			if err := deleteRemoteRow(key); err != nil {
				return result, fmt.Errorf("apply tombstone to remote %s/%s: %w", table, key, err)
			}
		}
		if localHas {
			if err := deleteLocalRow(key); err != nil {
				return result, fmt.Errorf("apply tombstone to local %s/%s: %w", table, key, err)
			}
		}
		if !hasRemoteTs || remoteTs != ts {
			if err := s.insertTombstoneToRemote(table, key, ts); err != nil {
				return result, fmt.Errorf("propagate tombstone to remote %s/%s: %w", table, key, err)
			}
		}
		if !hasLocalTs || localTs != ts {
			if err := s.insertTombstoneToLocal(table, key, ts); err != nil {
				return result, fmt.Errorf("propagate tombstone to local %s/%s: %w", table, key, err)
			}
		}
		result.kill(key)
	}

	return result, nil
}

// InitialMigration performs a one-time migration from local to remote
// This is used when setting up sync for the first time
func (s *SyncService) InitialMigration() error {
	stats := SyncStats{StartTime: time.Now()}

	// Push all local data to remote (one direction only)
	if err := s.Sync(SyncPushOnly); err != nil {
		return fmt.Errorf("initial migration failed: %w", err)
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	return nil
}
