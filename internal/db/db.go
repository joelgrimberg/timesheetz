package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/logging"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Connect establishes a connection to the database
func Connect(dbPath string) error {
	// Close any existing connection
	if db != nil {
		db.Close()
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		// Close the connection if ping fails
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set pragmas for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to set journal mode: %w", err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	logging.Log("Connected to the database 🍺")
	return nil
}

// Close closes the database connection
func Close() {
	if db != nil {
		db.Close()
	}
	logging.Log("Disconnected from the database 🍺")
}

// TimesheetEntry represents a row in the timesheet table
type TimesheetEntry struct {
	Id             int
	Date           string
	Client_name    string
	Client_hours   int
	Vacation_hours int
	Idle_hours     int
	Training_hours int
	Total_hours    int
	Sick_hours     int
	Holiday_hours  int
}

// GetDBPath returns the path to the database file
func GetDBPath() string {
	// Check if development mode is enabled
	if config.GetDevelopmentMode() {
		// In development mode, use a local database file
		dbPath := "timesheet.db"
		logging.Log("Using development database at: %s", dbPath)
		return dbPath
	}

	// In production mode, use ~/.local/share/timesheetz/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	// Create timesheet directory if it doesn't exist
	timesheetDir := filepath.Join(homeDir, ".local", "share", "timesheetz")
	if err := os.MkdirAll(timesheetDir, 0755); err != nil {
		log.Fatalf("Failed to create timesheet directory: %v", err)
	}

	// Ensure directory has correct permissions
	if err := os.Chmod(timesheetDir, 0755); err != nil {
		log.Fatalf("Failed to set directory permissions: %v", err)
	}

	dbPath := filepath.Join(timesheetDir, "timesheet.db")
	logging.Log("Using production database at: %s", dbPath)
	return dbPath
}

// InitializeDatabase creates the database and tables if they don't exist
func InitializeDatabase(dbPath string) error {
	// Ensure the directory exists (skip for in-memory databases)
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" && dbPath != ":memory:" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for database: %w", err)
		}
		// Ensure directory has correct permissions
		if err := os.Chmod(dir, 0755); err != nil {
			return fmt.Errorf("failed to set directory permissions: %w", err)
		}
	}

	// Close any existing connection
	if db != nil {
		db.Close()
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Execute each statement separately to ensure all tables are created
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS timesheet (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			client_name TEXT NOT NULL,
			client_hours INTEGER DEFAULT NULL,
			vacation_hours INTEGER DEFAULT NULL,
			idle_hours INTEGER DEFAULT NULL,
			training_hours INTEGER DEFAULT NULL,
			sick_hours INTEGER DEFAULT NULL,
			holiday_hours INTEGER DEFAULT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_client_name ON timesheet(client_name);`,
		`CREATE TABLE IF NOT EXISTS training_budget (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			training_name TEXT NOT NULL,
			hours INTEGER NOT NULL,
			cost_without_vat DECIMAL(10,2) NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_training_date ON training_budget(date);`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nSQL: %s", err, stmt)
		}
	}

	// Set database permissions AFTER the file is created (skip for in-memory databases)
	if dbPath != ":memory:" {
		// Check if file exists before trying to chmod
		if _, err := os.Stat(dbPath); err == nil {
			if err := os.Chmod(dbPath, 0644); err != nil {
				// Log warning but don't fail - permissions might not be critical
				logging.Log("Warning: failed to set database permissions: %v", err)
			}
		}
	}

	logging.Log("Database initialized successfully 🍺")
	return nil
}

// GetAllTimesheetEntries retrieves entries from the timesheet table
// If year and month are provided (non-zero), it filters entries for that specific month
func GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error) {
	var query string
	var args []any

	baseQuery := "SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, " +
		"(client_hours + vacation_hours + idle_hours + training_hours + sick_hours + holiday_hours) AS total_hours " +
		"FROM timesheet"

	if year != 0 && month != 0 {
		// Filter by specific month and year
		startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.UTC).Format("2006-01-02")

		query = baseQuery + " WHERE date BETWEEN ? AND ?"
		args = []any{startDate, endDate}
	} else {
		// Get all entries
		query = baseQuery
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TimesheetEntry
	for rows.Next() {
		var entry TimesheetEntry
		if err := rows.Scan(&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours,
			&entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours, &entry.Sick_hours, &entry.Holiday_hours, &entry.Total_hours); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// GetTimesheetEntryByDate retrieves a single timesheet entry by date
func GetTimesheetEntryByDate(date string) (TimesheetEntry, error) {
	query := `SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours,
              (client_hours + vacation_hours + idle_hours + training_hours + holiday_hours + sick_hours) AS total_hours
              FROM timesheet WHERE date = ?`

	var entry TimesheetEntry
	err := db.QueryRow(query, date).Scan(
		&entry.Id,
		&entry.Date,
		&entry.Client_name,
		&entry.Client_hours,
		&entry.Vacation_hours,
		&entry.Idle_hours,
		&entry.Training_hours,
		&entry.Sick_hours,
		&entry.Holiday_hours,
		&entry.Total_hours,
	)
	if err != nil {
		return TimesheetEntry{}, err
	}

	return entry, nil
}

func AddTimesheetEntry(entry TimesheetEntry) error {
	// Remove debug output
	// fmt.Printf("DEBUG: AddTimesheetEntry - Date: %s, Client: %s, VacationHours: %d\n",
	// 	entry.Date, entry.Client_name, entry.Vacation_hours)

	query := `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query,
		entry.Date,
		entry.Client_name,
		entry.Client_hours,
		entry.Vacation_hours,
		entry.Idle_hours,
		entry.Training_hours,
		entry.Sick_hours,
		entry.Holiday_hours)
	if err != nil {
		// fmt.Printf("DEBUG: AddTimesheetEntry error: %v\n", err)
		return err
	}

	// fmt.Printf("DEBUG: AddTimesheetEntry success\n")
	return nil
}

// UpdateTimesheetEntry updates an existing Timesheet entry by date
func UpdateTimesheetEntry(entry TimesheetEntry) error {
	query := `UPDATE timesheet 
              SET client_name = ?, client_hours = ?, 
                  vacation_hours = ?, idle_hours = ?, training_hours = ?, holiday_hours = ?, sick_hours = ?
              WHERE date = ?`

	result, err := db.Exec(query,
		entry.Client_name,
		entry.Client_hours,
		entry.Vacation_hours,
		entry.Idle_hours,
		entry.Training_hours,
		entry.Holiday_hours,
		entry.Sick_hours,
		entry.Date)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no entry found with date %s", entry.Date)
	}

	return nil
}

// PutTimesheetEntry inserts a new timesheet entry with the current date
func PutTimesheetEntry(clientHours, vacationHours, idleHours, trainingHours, holidayHours, sickHours float64) (int64, error) {
	// Get current date in YYYY-MM-DD format
	currentDate := time.Now().Format("2006-01-02")

	// Use prepared statement to prevent SQL injection
	stmt, err := db.Prepare("INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, holiday_hours, sick_hours) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the statement with client name as parameter
	// Note: Replaced hardcoded value 1 with a client name
	result, err := stmt.Exec(currentDate, "default", clientHours, vacationHours, idleHours, trainingHours, holidayHours, sickHours)
	if err != nil {
		return 0, err
	}

	// Return the new entry's ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateTimesheetEntryById(id string, data map[string]any) error {
	// Validate allowed fields to prevent SQL injection
	allowedFields := map[string]bool{
		"client_hours":   true,
		"vacation_hours": true,
		"idle_hours":     true,
		"training_hours": true,
		"holiday_hours":  true,
		"sick_hours":     true,
	}

	// Start building the query
	query := "UPDATE timesheet SET "

	// Add each field to be updated
	values := []any{}
	setStatements := []string{}

	for key, val := range data {
		// Check if the field is allowed
		if !allowedFields[key] {
			return fmt.Errorf("field %s is not allowed for update", key)
		}
		setStatements = append(setStatements, key+" = ?")
		values = append(values, val)
	}

	if len(setStatements) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	query += strings.Join(setStatements, ", ")
	query += " WHERE id = ?"
	values = append(values, id)

	// Execute the query
	result, err := db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no entry found with id %s", id)
	}

	return nil
}

// DeleteTimesheetEntryByDate removes a timesheet entry by its date
func DeleteTimesheetEntryByDate(date string) error {
	// Use prepared statement to prevent SQL injection
	stmt, err := db.Prepare("DELETE FROM timesheet WHERE date = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(date)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// DeleteTimesheetEntry removes a timesheet entry by its ID
func DeleteTimesheetEntry(id string) error {
	// Use prepared statement to prevent SQL injection
	stmt, err := db.Prepare("DELETE FROM timesheet WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

func Ping() error {
	return db.Ping()
}

// GetLastClientName returns the client name from the most recent timesheet entry
func GetLastClientName() (string, error) {
	query := `SELECT client_name FROM timesheet ORDER BY date DESC LIMIT 1`
	var clientName string
	err := db.QueryRow(query).Scan(&clientName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Return empty string if no entries exist
		}
		return "", fmt.Errorf("failed to get last client name: %w", err)
	}
	return clientName, nil
}

// GetVacationEntriesForYear returns all vacation days with vacation_hours > 0 from the timesheet table
func GetVacationEntriesForYear(year int) ([]TimesheetEntry, error) {
	rows, err := db.Query(`
		SELECT id, date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, (client_hours + vacation_hours + idle_hours + training_hours + sick_hours + holiday_hours) AS total_hours
		FROM timesheet
		WHERE strftime('%Y', date) = ? AND vacation_hours > 0
		ORDER BY date DESC
	`, fmt.Sprintf("%d", year))
	if err != nil {
		return nil, fmt.Errorf("failed to query timesheet vacation entries: %w", err)
	}
	defer rows.Close()

	var entries []TimesheetEntry
	for rows.Next() {
		var entry TimesheetEntry
		if err := rows.Scan(&entry.Id, &entry.Date, &entry.Client_name, &entry.Client_hours, &entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours, &entry.Sick_hours, &entry.Holiday_hours, &entry.Total_hours); err != nil {
			return nil, fmt.Errorf("failed to scan timesheet vacation entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// GetVacationHoursForYear returns the total vacation hours used in a given year (from timesheet table only)
func GetVacationHoursForYear(year int) (int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(vacation_hours), 0)
		FROM timesheet
		WHERE strftime('%Y', date) = ? AND vacation_hours > 0
	`, fmt.Sprintf("%d", year)).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get vacation hours from timesheet table: %w", err)
	}
	return total, nil
}
