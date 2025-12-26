package db

import (
	"fmt"
)

// GetTrainingEntriesForYear retrieves all training entries for a specific year
func GetTrainingEntriesForYear(year int) ([]TimesheetEntry, error) {
	// Calculate start and end dates for the year
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	// Query the database
	rows, err := db.Query(`
		SELECT id, date, client_name, client_hours, training_hours, vacation_hours, 
		       idle_hours, holiday_hours, sick_hours,
		       (client_hours + training_hours + vacation_hours + idle_hours + holiday_hours + sick_hours) as total_hours
		FROM timesheet
		WHERE date BETWEEN ? AND ?
		AND training_hours > 0
		ORDER BY date DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process the results
	// Pre-allocate slice with capacity for typical training days per year
	entries := make([]TimesheetEntry, 0, 50)
	for rows.Next() {
		var entry TimesheetEntry
		err := rows.Scan(
			&entry.Id,
			&entry.Date,
			&entry.Client_name,
			&entry.Client_hours,
			&entry.Training_hours,
			&entry.Vacation_hours,
			&entry.Idle_hours,
			&entry.Holiday_hours,
			&entry.Sick_hours,
			&entry.Total_hours,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// TrainingBudgetEntry represents a training budget entry
type TrainingBudgetEntry struct {
	Id               int
	Date             string
	Training_name    string
	Hours            int
	Cost_without_vat float64
}

// GetTrainingBudgetEntriesForYear retrieves all training budget entries for a specific year
func GetTrainingBudgetEntriesForYear(year int) ([]TrainingBudgetEntry, error) {
	// Calculate start and end dates for the year
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	// Query the database
	rows, err := db.Query(`
		SELECT id, date, training_name, hours, cost_without_vat
		FROM training_budget
		WHERE date BETWEEN ? AND ?
		ORDER BY date DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process the results
	// Pre-allocate slice with capacity for typical training budget entries per year
	entries := make([]TrainingBudgetEntry, 0, 50)
	for rows.Next() {
		var entry TrainingBudgetEntry
		err := rows.Scan(
			&entry.Id,
			&entry.Date,
			&entry.Training_name,
			&entry.Hours,
			&entry.Cost_without_vat,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// AddTrainingBudgetEntry adds a new training budget entry
func AddTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	query := `INSERT INTO training_budget (date, training_name, hours, cost_without_vat)
              VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query,
		entry.Date,
		entry.Training_name,
		entry.Hours,
		entry.Cost_without_vat)
	return err
}

// UpdateTrainingBudgetEntry updates an existing training budget entry
func UpdateTrainingBudgetEntry(entry TrainingBudgetEntry) error {
	query := `UPDATE training_budget 
              SET date = ?, training_name = ?, hours = ?, cost_without_vat = ?
              WHERE id = ?`
	_, err := db.Exec(query,
		entry.Date,
		entry.Training_name,
		entry.Hours,
		entry.Cost_without_vat,
		entry.Id)
	return err
}

// DeleteTrainingBudgetEntry removes a training budget entry
func DeleteTrainingBudgetEntry(id int) error {
	query := `DELETE FROM training_budget WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// GetTrainingBudgetEntry retrieves a single training budget entry by ID
func GetTrainingBudgetEntry(id int) (TrainingBudgetEntry, error) {
	query := `SELECT id, date, training_name, hours, cost_without_vat
              FROM training_budget WHERE id = ?`

	var entry TrainingBudgetEntry
	err := db.QueryRow(query, id).Scan(
		&entry.Id,
		&entry.Date,
		&entry.Training_name,
		&entry.Hours,
		&entry.Cost_without_vat,
	)
	if err != nil {
		return TrainingBudgetEntry{}, err
	}

	return entry, nil
}

// GetTrainingBudgetEntryByDate retrieves a single training budget entry by date
func GetTrainingBudgetEntryByDate(date string) (TrainingBudgetEntry, error) {
	query := `SELECT id, date, training_name, hours, cost_without_vat
              FROM training_budget WHERE date = ?`

	var entry TrainingBudgetEntry
	err := db.QueryRow(query, date).Scan(
		&entry.Id,
		&entry.Date,
		&entry.Training_name,
		&entry.Hours,
		&entry.Cost_without_vat,
	)
	if err != nil {
		return TrainingBudgetEntry{}, err
	}

	return entry, nil
}
