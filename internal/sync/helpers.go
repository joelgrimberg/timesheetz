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
