package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var db *sql.DB

func Connect(user, password string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/timesheet", user, password)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Connected to the database 🍺")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
	fmt.Println("Disconnected from the database 🍺")
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
}

// GetAllTimesheetEntries retrieves entries from the timesheet table
// If year and month are provided (non-zero), it filters entries for that specific month
func GetAllTimesheetEntries(year int, month time.Month) ([]TimesheetEntry, error) {
	var query string
	var args []any

	baseQuery := "SELECT t.id, t.date, c.client_name, t.client_hours, t.vacation_hours, t.idle_hours, t.training_hours, " +
		"(t.client_hours + t.vacation_hours + t.idle_hours + t.training_hours) AS total_hours " +
		"FROM timesheet t JOIN clients c ON t.client_id = c.client_id"

	if year != 0 && month != 0 {
		// Filter by specific month and year
		startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.UTC).Format("2006-01-02")

		query = baseQuery + " WHERE t.date BETWEEN ? AND ?"
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
			&entry.Vacation_hours, &entry.Idle_hours, &entry.Training_hours, &entry.Total_hours); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// / PutTimesheetEntry inserts a new timesheet entry with the current date
// and client_id fixed to '1'
func PutTimesheetEntry(clientHours, vacationHours, idleHours, trainingHours float64) (int64, error) {
	// Get current date in YYYY-MM-DD format
	currentDate := time.Now().Format("2006-01-02")

	// Use prepared statement to prevent SQL injection
	stmt, err := db.Prepare("INSERT INTO timesheet (date, client_id, client_hours, vacation_hours, idle_hours, training_hours) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Execute the statement with fixed client_id=1
	result, err := stmt.Exec(currentDate, 1, clientHours, vacationHours, idleHours, trainingHours)
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

func UpdateTimesheetEntry(id string, data map[string]any) error {
	// Validate allowed fields to prevent SQL injection
	allowedFields := map[string]bool{
		"client_hours":   true,
		"vacation_hours": true,
		"idle_hours":     true,
		"training_hours": true,
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
