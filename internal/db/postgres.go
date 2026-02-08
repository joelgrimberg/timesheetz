package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/logging"

	_ "github.com/lib/pq"
)

var pgDB *sql.DB

// PostgresDBLayer implements DataLayer for PostgreSQL
type PostgresDBLayer struct{}

// ConnectPostgres establishes connection to PostgreSQL
func ConnectPostgres(connStr string) error {
	// Close any existing connection
	if pgDB != nil {
		pgDB.Close()
	}

	var err error
	pgDB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open postgres: %w", err)
	}

	// Test the connection
	if err = pgDB.Ping(); err != nil {
		pgDB.Close()
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Set connection pool settings
	pgDB.SetMaxOpenConns(10)
	pgDB.SetMaxIdleConns(5)
	pgDB.SetConnMaxLifetime(time.Hour)

	logging.Log("Connected to PostgreSQL database")
	return nil
}

// ClosePostgres closes the PostgreSQL connection
func ClosePostgres() {
	if pgDB != nil {
		pgDB.Close()
	}
	logging.Log("Disconnected from PostgreSQL database")
}

// GetPostgresDB returns the raw PostgreSQL database connection for sync operations
func GetPostgresDB() *sql.DB {
	return pgDB
}

// PingPostgres tests the PostgreSQL connection
func PingPostgres() error {
	if pgDB == nil {
		return fmt.Errorf("postgres connection not established")
	}
	return pgDB.Ping()
}

// Timesheet operations

func (p *PostgresDBLayer) GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error) {
	var query string
	var args []any
	argNum := 1

	baseQuery := `SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours,
		(client_hours + vacation_hours + idle_hours + training_hours + sick_hours + holiday_hours) AS total_hours
		FROM timesheet`

	if year != 0 && month != 0 {
		startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.UTC).Format("2006-01-02")
		query = baseQuery + fmt.Sprintf(" WHERE date BETWEEN $%d AND $%d", argNum, argNum+1)
		args = []any{startDate, endDate}
	} else if year != 0 {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC).Format("2006-01-02")
		query = baseQuery + fmt.Sprintf(" WHERE date BETWEEN $%d AND $%d", argNum, argNum+1)
		args = []any{startDate, endDate}
	} else {
		query = baseQuery
	}

	rows, err := pgDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var capacity int
	if year != 0 && month != 0 {
		capacity = 31
	} else if year != 0 {
		capacity = 365
	} else {
		capacity = 500
	}
	entries := make([]TimesheetEntry, 0, capacity)

	for rows.Next() {
		var entry TimesheetEntry
		if err := rows.Scan(&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours,
			&entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours, &entry.Sick_hours,
			&entry.Holiday_hours, &entry.Total_hours); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (p *PostgresDBLayer) GetTimesheetEntryByDate(date string) (TimesheetEntry, error) {
	query := `SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours,
		(client_hours + vacation_hours + idle_hours + training_hours + holiday_hours + sick_hours) AS total_hours
		FROM timesheet WHERE date = $1`

	var entry TimesheetEntry
	err := pgDB.QueryRow(query, date).Scan(
		&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours,
		&entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours,
		&entry.Sick_hours, &entry.Holiday_hours, &entry.Total_hours,
	)
	if err != nil {
		return TimesheetEntry{}, err
	}
	return entry, nil
}

func (p *PostgresDBLayer) AddTimesheetEntry(entry TimesheetEntry) error {
	query := `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := pgDB.Exec(query,
		entry.Date, entry.Client_name, entry.Client_hours, entry.Vacation_hours,
		entry.Idle_hours, entry.Training_hours, entry.Sick_hours, entry.Holiday_hours)
	return err
}

func (p *PostgresDBLayer) UpdateTimesheetEntry(entry TimesheetEntry) error {
	query := `UPDATE timesheet
		SET client_name = $1, client_hours = $2, vacation_hours = $3, idle_hours = $4,
		    training_hours = $5, holiday_hours = $6, sick_hours = $7
		WHERE date = $8`

	result, err := pgDB.Exec(query,
		entry.Client_name, entry.Client_hours, entry.Vacation_hours,
		entry.Idle_hours, entry.Training_hours, entry.Holiday_hours,
		entry.Sick_hours, entry.Date)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no entry found with date %s", entry.Date)
	}
	return nil
}

func (p *PostgresDBLayer) UpdateTimesheetEntryById(id string, data map[string]any) error {
	return UpdateTimesheetEntryByIdPostgres(id, data)
}

func (p *PostgresDBLayer) DeleteTimesheetEntryByDate(date string) error {
	_, err := pgDB.Exec("DELETE FROM timesheet WHERE date = $1", date)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

func (p *PostgresDBLayer) DeleteTimesheetEntry(id string) error {
	_, err := pgDB.Exec("DELETE FROM timesheet WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

func (p *PostgresDBLayer) GetLastClientName() (string, error) {
	query := `SELECT client_name FROM timesheet ORDER BY date DESC LIMIT 1`
	var clientName string
	err := pgDB.QueryRow(query).Scan(&clientName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get last client name: %w", err)
	}
	return clientName, nil
}

// Training/Vacation operations

func (p *PostgresDBLayer) GetTrainingEntriesForYear(year int) ([]TimesheetEntry, error) {
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	rows, err := pgDB.Query(`
		SELECT id, date, client_name, client_hours, training_hours, vacation_hours,
		       idle_hours, holiday_hours, sick_hours,
		       (client_hours + training_hours + vacation_hours + idle_hours + holiday_hours + sick_hours) as total_hours
		FROM timesheet
		WHERE date BETWEEN $1 AND $2
		AND training_hours > 0
		ORDER BY date DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]TimesheetEntry, 0, 50)
	for rows.Next() {
		var entry TimesheetEntry
		err := rows.Scan(
			&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours,
			&entry.Training_hours, &entry.Vacation_hours, &entry.Idle_hours,
			&entry.Holiday_hours, &entry.Sick_hours, &entry.Total_hours,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (p *PostgresDBLayer) GetVacationEntriesForYear(year int) ([]TimesheetEntry, error) {
	rows, err := pgDB.Query(`
		SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours,
		       (client_hours + vacation_hours + idle_hours + training_hours + sick_hours + holiday_hours) AS total_hours
		FROM timesheet
		WHERE EXTRACT(YEAR FROM date::date) = $1 AND vacation_hours > 0
		ORDER BY date DESC
	`, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query timesheet vacation entries: %w", err)
	}
	defer rows.Close()

	entries := make([]TimesheetEntry, 0, 30)
	for rows.Next() {
		var entry TimesheetEntry
		if err := rows.Scan(&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours,
			&entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours,
			&entry.Sick_hours, &entry.Holiday_hours, &entry.Total_hours); err != nil {
			return nil, fmt.Errorf("failed to scan timesheet vacation entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (p *PostgresDBLayer) GetVacationHoursForYear(year int) (int, error) {
	var total int
	err := pgDB.QueryRow(`
		SELECT COALESCE(SUM(vacation_hours), 0)
		FROM timesheet
		WHERE EXTRACT(YEAR FROM date::date) = $1 AND vacation_hours > 0
	`, year).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get vacation hours from timesheet table: %w", err)
	}
	return total, nil
}

// Vacation carryover operations

func (p *PostgresDBLayer) GetVacationCarryoverForYear(year int) (VacationCarryover, error) {
	var carryover VacationCarryover
	err := pgDB.QueryRow(`
		SELECT id, year, carryover_hours, source_year, created_at, updated_at, COALESCE(notes, '') as notes
		FROM vacation_carryover
		WHERE year = $1
	`, year).Scan(
		&carryover.Id, &carryover.Year, &carryover.CarryoverHours,
		&carryover.SourceYear, &carryover.CreatedAt, &carryover.UpdatedAt, &carryover.Notes,
	)

	if err == sql.ErrNoRows {
		return VacationCarryover{
			Year:           year,
			CarryoverHours: 0,
			SourceYear:     year - 1,
		}, nil
	}

	if err != nil {
		return VacationCarryover{}, fmt.Errorf("failed to get vacation carryover: %w", err)
	}
	return carryover, nil
}

func (p *PostgresDBLayer) SetVacationCarryover(carryover VacationCarryover) error {
	// PostgreSQL upsert using ON CONFLICT
	_, err := pgDB.Exec(`
		INSERT INTO vacation_carryover (year, carryover_hours, source_year, updated_at, notes)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4)
		ON CONFLICT (year) DO UPDATE SET
			carryover_hours = EXCLUDED.carryover_hours,
			source_year = EXCLUDED.source_year,
			updated_at = CURRENT_TIMESTAMP,
			notes = EXCLUDED.notes
	`, carryover.Year, carryover.CarryoverHours, carryover.SourceYear, carryover.Notes)

	if err != nil {
		return fmt.Errorf("failed to set vacation carryover: %w", err)
	}
	return nil
}

func (p *PostgresDBLayer) DeleteVacationCarryover(year int) error {
	_, err := pgDB.Exec(`DELETE FROM vacation_carryover WHERE year = $1`, year)
	if err != nil {
		return fmt.Errorf("failed to delete vacation carryover: %w", err)
	}
	return nil
}

func (p *PostgresDBLayer) GetVacationSummaryForYear(year int) (VacationSummary, error) {
	summary := VacationSummary{Year: year}

	cfg, err := config.GetConfig()
	if err != nil {
		return summary, fmt.Errorf("failed to get config: %w", err)
	}
	summary.YearlyTarget = cfg.VacationHours.YearlyTarget

	carryover, err := p.GetVacationCarryoverForYear(year)
	if err != nil {
		return summary, fmt.Errorf("failed to get carryover: %w", err)
	}
	summary.CarryoverHours = carryover.CarryoverHours

	usedHours, err := p.GetVacationHoursForYear(year)
	if err != nil {
		return summary, fmt.Errorf("failed to get used hours: %w", err)
	}
	summary.UsedHours = usedHours

	summary.TotalAvailable = summary.YearlyTarget + summary.CarryoverHours

	if usedHours <= summary.CarryoverHours {
		summary.UsedFromCarryover = usedHours
		summary.UsedFromCurrent = 0
	} else {
		summary.UsedFromCarryover = summary.CarryoverHours
		summary.UsedFromCurrent = usedHours - summary.CarryoverHours
	}

	summary.RemainingTotal = summary.TotalAvailable - usedHours
	return summary, nil
}

// Training budget operations

func (p *PostgresDBLayer) GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error) {
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	rows, err := pgDB.Query(`
		SELECT id, date, training_name, hours, cost_without_vat
		FROM training_budget
		WHERE date BETWEEN $1 AND $2
		ORDER BY date DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]TrainingBudgetEntry, 0, 50)
	for rows.Next() {
		var entry TrainingBudgetEntry
		err := rows.Scan(&entry.Id, &entry.Date, &entry.Training_name, &entry.Hours, &entry.Cost_without_vat)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (p *PostgresDBLayer) AddTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	query := `INSERT INTO training_budget (date, training_name, hours, cost_without_vat)
		VALUES ($1, $2, $3, $4)`
	_, err := pgDB.Exec(query, entry.Date, entry.Training_name, entry.Hours, entry.Cost_without_vat)
	return err
}

func (p *PostgresDBLayer) UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	query := `UPDATE training_budget
		SET date = $1, training_name = $2, hours = $3, cost_without_vat = $4
		WHERE id = $5`
	_, err := pgDB.Exec(query, entry.Date, entry.Training_name, entry.Hours, entry.Cost_without_vat, entry.Id)
	return err
}

func (p *PostgresDBLayer) DeleteTrainingBudgetEntry(id int) error {
	_, err := pgDB.Exec(`DELETE FROM training_budget WHERE id = $1`, id)
	return err
}

func (p *PostgresDBLayer) GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error) {
	query := `SELECT id, date, training_name, hours, cost_without_vat FROM training_budget WHERE id = $1`
	var entry TrainingBudgetEntry
	err := pgDB.QueryRow(query, id).Scan(&entry.Id, &entry.Date, &entry.Training_name, &entry.Hours, &entry.Cost_without_vat)
	if err != nil {
		return TrainingBudgetEntry{}, err
	}
	return entry, nil
}

func (p *PostgresDBLayer) GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error) {
	query := `SELECT id, date, training_name, hours, cost_without_vat FROM training_budget WHERE date = $1`
	var entry TrainingBudgetEntry
	err := pgDB.QueryRow(query, date).Scan(&entry.Id, &entry.Date, &entry.Training_name, &entry.Hours, &entry.Cost_without_vat)
	if err != nil {
		return TrainingBudgetEntry{}, err
	}
	return entry, nil
}

// Client operations

func (p *PostgresDBLayer) GetAllClients() ([]Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients ORDER BY name ASC`
	rows, err := pgDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	clients := make([]Client, 0, 10)
	for rows.Next() {
		var client Client
		var isActive int
		if err := rows.Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		client.IsActive = isActive == 1
		clients = append(clients, client)
	}
	return clients, rows.Err()
}

func (p *PostgresDBLayer) GetActiveClients() ([]Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE is_active = 1 ORDER BY name ASC`
	rows, err := pgDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active clients: %w", err)
	}
	defer rows.Close()

	clients := make([]Client, 0, 10)
	for rows.Next() {
		var client Client
		var isActive int
		if err := rows.Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		client.IsActive = isActive == 1
		clients = append(clients, client)
	}
	return clients, rows.Err()
}

func (p *PostgresDBLayer) GetClientById(id int) (Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE id = $1`
	var client Client
	var isActive int
	err := pgDB.QueryRow(query, id).Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, fmt.Errorf("client not found")
		}
		return Client{}, fmt.Errorf("failed to query client: %w", err)
	}
	client.IsActive = isActive == 1
	return client, nil
}

func (p *PostgresDBLayer) GetClientByName(name string) (Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE name = $1`
	var client Client
	var isActive int
	err := pgDB.QueryRow(query, name).Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, fmt.Errorf("client not found")
		}
		return Client{}, fmt.Errorf("failed to query client: %w", err)
	}
	client.IsActive = isActive == 1
	return client, nil
}

func (p *PostgresDBLayer) AddClient(client Client) (int, error) {
	query := `INSERT INTO clients (name, created_at, is_active) VALUES ($1, $2, $3) RETURNING id`
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	isActive := 0
	if client.IsActive {
		isActive = 1
	}

	var id int
	err := pgDB.QueryRow(query, client.Name, createdAt, isActive).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add client: %w", err)
	}
	return id, nil
}

func (p *PostgresDBLayer) UpdateClient(client Client) error {
	query := `UPDATE clients SET name = $1, is_active = $2 WHERE id = $3`
	isActive := 0
	if client.IsActive {
		isActive = 1
	}

	result, err := pgDB.Exec(query, client.Name, isActive, client.Id)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}
	return nil
}

func (p *PostgresDBLayer) DeleteClient(id int) error {
	result, err := pgDB.Exec(`DELETE FROM clients WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}
	return nil
}

func (p *PostgresDBLayer) DeactivateClient(id int) error {
	result, err := pgDB.Exec(`UPDATE clients SET is_active = 0 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}
	return nil
}

// Client rate operations

func (p *PostgresDBLayer) GetClientRates(clientId int) ([]ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
		FROM client_rates
		WHERE client_id = $1
		ORDER BY effective_date DESC, created_at DESC`

	rows, err := pgDB.Query(query, clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to query client rates: %w", err)
	}
	defer rows.Close()

	rates := make([]ClientRate, 0, 10)
	for rows.Next() {
		var rate ClientRate
		if err := rows.Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
			&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan client rate: %w", err)
		}
		rates = append(rates, rate)
	}
	return rates, rows.Err()
}

func (p *PostgresDBLayer) GetClientRateById(id int) (ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
		FROM client_rates WHERE id = $1`

	var rate ClientRate
	err := pgDB.QueryRow(query, id).Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
		&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ClientRate{}, fmt.Errorf("client rate not found")
		}
		return ClientRate{}, fmt.Errorf("failed to query client rate: %w", err)
	}
	return rate, nil
}

func (p *PostgresDBLayer) AddClientRate(rate ClientRate) error {
	query := `INSERT INTO client_rates (client_id, hourly_rate, effective_date, notes, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	_, err := pgDB.Exec(query, rate.ClientId, rate.HourlyRate, rate.EffectiveDate, rate.Notes, createdAt)
	if err != nil {
		return fmt.Errorf("failed to add client rate: %w", err)
	}
	return nil
}

func (p *PostgresDBLayer) UpdateClientRate(rate ClientRate) error {
	query := `UPDATE client_rates SET hourly_rate = $1, effective_date = $2, notes = $3 WHERE id = $4`
	result, err := pgDB.Exec(query, rate.HourlyRate, rate.EffectiveDate, rate.Notes, rate.Id)
	if err != nil {
		return fmt.Errorf("failed to update client rate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client rate not found")
	}
	return nil
}

func (p *PostgresDBLayer) DeleteClientRate(id int) error {
	result, err := pgDB.Exec(`DELETE FROM client_rates WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete client rate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("client rate not found")
	}
	return nil
}

func (p *PostgresDBLayer) GetClientRateForDate(clientId int, date string) (ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
		FROM client_rates
		WHERE client_id = $1 AND effective_date <= $2
		ORDER BY effective_date DESC, created_at DESC
		LIMIT 1`

	var rate ClientRate
	err := pgDB.QueryRow(query, clientId, date).Scan(&rate.Id, &rate.ClientId,
		&rate.HourlyRate, &rate.EffectiveDate, &rate.Notes, &rate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ClientRate{}, fmt.Errorf("no rate found for client on date %s", date)
		}
		return ClientRate{}, fmt.Errorf("failed to query client rate: %w", err)
	}
	return rate, nil
}

func (p *PostgresDBLayer) GetClientRateByName(clientName string, date string) (float64, error) {
	client, err := p.GetClientByName(clientName)
	if err != nil {
		return 0.0, nil
	}

	rate, err := p.GetClientRateForDate(client.Id, date)
	if err != nil {
		return 0.0, nil
	}
	return rate.HourlyRate, nil
}

// Earnings operations

// pgRateCache holds cached client and rate information for PostgreSQL
type pgRateCache struct {
	clientsByName map[string]int
	ratesByClient map[int][]ClientRate
}

func (p *PostgresDBLayer) buildRateCache() (*pgRateCache, error) {
	cache := &pgRateCache{
		clientsByName: make(map[string]int),
		ratesByClient: make(map[int][]ClientRate),
	}

	clients, err := p.GetAllClients()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}
	for _, client := range clients {
		cache.clientsByName[client.Name] = client.Id
	}

	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
		FROM client_rates
		ORDER BY client_id, effective_date DESC`

	rows, err := pgDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rate ClientRate
		if err := rows.Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
			&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		cache.ratesByClient[rate.ClientId] = append(cache.ratesByClient[rate.ClientId], rate)
	}
	return cache, nil
}

func (c *pgRateCache) getRateFromCache(clientName string, date string) float64 {
	clientId, ok := c.clientsByName[clientName]
	if !ok {
		return 0.0
	}

	rates, ok := c.ratesByClient[clientId]
	if !ok || len(rates) == 0 {
		return 0.0
	}

	for _, rate := range rates {
		if rate.EffectiveDate <= date {
			return rate.HourlyRate
		}
	}
	return 0.0
}

func (p *PostgresDBLayer) CalculateEarningsForYear(year int) (EarningsOverview, error) {
	cache, err := p.buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	entries, err := p.GetAllTimesheetEntries(year, 0)
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	earningsEntries := make([]EarningsEntry, 0, 300)
	var totalHours int
	var totalEarnings float64

	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		rate := cache.getRateFromCache(entry.Client_name, entry.Date)
		earnings := float64(entry.Client_hours) * rate

		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.Client_name,
			ClientHours: entry.Client_hours,
			HourlyRate:  rate,
			Earnings:    earnings,
		})

		totalHours += entry.Client_hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         0,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

func (p *PostgresDBLayer) CalculateEarningsSummaryForYear(year int) (EarningsOverview, error) {
	cache, err := p.buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	entries, err := p.GetAllTimesheetEntries(year, 0)
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	type ClientRateKey struct {
		ClientName string
		Rate       float64
	}
	aggregated := make(map[ClientRateKey]int)

	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		rate := cache.getRateFromCache(entry.Client_name, entry.Date)
		key := ClientRateKey{ClientName: entry.Client_name, Rate: rate}
		aggregated[key] += entry.Client_hours
	}

	earningsEntries := make([]EarningsEntry, 0, len(aggregated))
	var totalHours int
	var totalEarnings float64

	for key, hours := range aggregated {
		earnings := float64(hours) * key.Rate
		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        "",
			ClientName:  key.ClientName,
			ClientHours: hours,
			HourlyRate:  key.Rate,
			Earnings:    earnings,
		})
		totalHours += hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         0,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

func (p *PostgresDBLayer) CalculateEarningsForMonth(year int, month int) (EarningsOverview, error) {
	cache, err := p.buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	entries, err := p.GetAllTimesheetEntries(year, time.Month(month))
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	earningsEntries := make([]EarningsEntry, 0, 30)
	var totalHours int
	var totalEarnings float64

	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		rate := cache.getRateFromCache(entry.Client_name, entry.Date)
		earnings := float64(entry.Client_hours) * rate

		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.Client_name,
			ClientHours: entry.Client_hours,
			HourlyRate:  rate,
			Earnings:    earnings,
		})

		totalHours += entry.Client_hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         month,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

func (p *PostgresDBLayer) GetClientWithRates(clientId int) (ClientWithRates, error) {
	client, err := p.GetClientById(clientId)
	if err != nil {
		return ClientWithRates{}, err
	}

	rates, err := p.GetClientRates(clientId)
	if err != nil {
		return ClientWithRates{}, err
	}

	return ClientWithRates{
		Client: client,
		Rates:  rates,
	}, nil
}

// Health check

func (p *PostgresDBLayer) Ping() error {
	return PingPostgres()
}

// UpdateTimesheetEntryByIdPostgres updates a timesheet entry by ID for PostgreSQL
func UpdateTimesheetEntryByIdPostgres(id string, data map[string]any) error {
	allowedFields := map[string]bool{
		"client_hours":   true,
		"vacation_hours": true,
		"idle_hours":     true,
		"training_hours": true,
		"holiday_hours":  true,
		"sick_hours":     true,
	}

	query := "UPDATE timesheet SET "
	values := []any{}
	setStatements := []string{}
	argNum := 1

	for key, val := range data {
		if !allowedFields[key] {
			return fmt.Errorf("field %s is not allowed for update", key)
		}
		setStatements = append(setStatements, fmt.Sprintf("%s = $%d", key, argNum))
		values = append(values, val)
		argNum++
	}

	if len(setStatements) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	query += strings.Join(setStatements, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", argNum)
	values = append(values, id)

	result, err := pgDB.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no entry found with id %s", id)
	}
	return nil
}
