package db

import (
	"strconv"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) string {
	// Use in-memory database for testing
	dbPath := ":memory:"

	// For in-memory databases, InitializeDatabase already opens the connection
	// so we don't need to call Connect separately
	if err := InitializeDatabase(dbPath); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return dbPath
}

func teardownTestDB(t *testing.T, dbPath string) {
	Close()
	// No need to remove in-memory database
}

func TestConnect(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Test that we can ping the database
	if err := Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestInitializeDatabase(t *testing.T) {
	// Use in-memory database
	dbPath := ":memory:"

	err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify tables were created by trying to connect
	if err := Connect(dbPath); err != nil {
		t.Fatalf("Failed to connect after initialization: %v", err)
	}
	defer Close()
}

func TestGetAllTimesheetEntries(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add test entries
	entry1 := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   6,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTimesheetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Test getting all entries
	entries, err := GetAllTimesheetEntries(0, 0)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Test filtering by month
	entries, err = GetAllTimesheetEntries(2024, time.January)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry for January, got %d", len(entries))
	}
	if entries[0].Date != "2024-01-15" {
		t.Errorf("Expected date 2024-01-15, got %s", entries[0].Date)
	}
}

func TestGetTimesheetEntryByDate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Test getting entry by date
	result, err := GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Client_name != "Client A" {
		t.Errorf("Expected Client A, got %s", result.Client_name)
	}

	// Test non-existent date
	_, err = GetTimesheetEntryByDate("2024-01-16")
	if err == nil {
		t.Error("Expected error for non-existent date")
	}
}

func TestAddTimesheetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 2,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	err := AddTimesheetEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Verify entry was added
	result, err := GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Client_name != "Client A" {
		t.Errorf("Expected Client A, got %s", result.Client_name)
	}
	if result.Vacation_hours != 2 {
		t.Errorf("Expected 2 vacation hours, got %d", result.Vacation_hours)
	}
}

func TestUpdateTimesheetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Update the entry
	entry.Client_hours = 6
	entry.Vacation_hours = 2
	err := UpdateTimesheetEntry(entry)
	if err != nil {
		t.Fatalf("Failed to update entry: %v", err)
	}

	// Verify update
	result, err := GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Client_hours != 6 {
		t.Errorf("Expected 6 client hours, got %d", result.Client_hours)
	}
	if result.Vacation_hours != 2 {
		t.Errorf("Expected 2 vacation hours, got %d", result.Vacation_hours)
	}

	// Test updating non-existent entry
	entry.Date = "2024-01-16"
	err = UpdateTimesheetEntry(entry)
	if err == nil {
		t.Error("Expected error for non-existent entry")
	}
}

func TestUpdateTimesheetEntryById(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get the entry to get its ID
	result, err := GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	// Update by ID
	data := map[string]any{
		"client_hours":   6,
		"vacation_hours": 2,
	}
	err = UpdateTimesheetEntryById(strconv.Itoa(result.Id), data)
	if err != nil {
		t.Fatalf("Failed to update entry: %v", err)
	}

	// Test invalid field
	data["invalid_field"] = "value"
	err = UpdateTimesheetEntryById(strconv.Itoa(result.Id), data)
	if err == nil {
		t.Error("Expected error for invalid field")
	}

	// Test empty data
	err = UpdateTimesheetEntryById(strconv.Itoa(result.Id), map[string]any{})
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Test non-existent ID
	err = UpdateTimesheetEntryById("999", data)
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}
}

func TestDeleteTimesheetEntryByDate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Delete the entry
	err := DeleteTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	_, err = GetTimesheetEntryByDate("2024-01-15")
	if err == nil {
		t.Error("Expected error for deleted entry")
	}
}

func TestDeleteTimesheetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get the entry to get its ID
	result, err := GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	// Delete by ID
	err = DeleteTimesheetEntry(strconv.Itoa(result.Id))
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	_, err = GetTimesheetEntryByDate("2024-01-15")
	if err == nil {
		t.Error("Expected error for deleted entry")
	}
}

func TestGetLastClientName(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Test with no entries
	name, err := GetLastClientName()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if name != "" {
		t.Errorf("Expected empty string, got %s", name)
	}

	// Add entries
	entry1 := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   6,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTimesheetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Test getting last client name
	name, err = GetLastClientName()
	if err != nil {
		t.Fatalf("Failed to get last client name: %v", err)
	}
	if name != "Client B" {
		t.Errorf("Expected Client B, got %s", name)
	}
}

func TestGetVacationEntriesForYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry1 := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 8,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTimesheetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	entries, err := GetVacationEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Failed to get vacation entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 vacation entry, got %d", len(entries))
	}
	if entries[0].Vacation_hours != 8 {
		t.Errorf("Expected 8 vacation hours, got %d", entries[0].Vacation_hours)
	}
}

func TestGetVacationHoursForYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry1 := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 8,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   0,
		Vacation_hours: 4,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTimesheetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	total, err := GetVacationHoursForYear(2024)
	if err != nil {
		t.Fatalf("Failed to get vacation hours: %v", err)
	}
	if total != 12 {
		t.Errorf("Expected 12 vacation hours, got %d", total)
	}
}

func TestPing(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	err := Ping()
	if err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestClose(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	Close()

	// Try to ping after close (should fail)
	err := Ping()
	if err == nil {
		t.Error("Expected error after closing database")
	}
}

func TestGetTrainingEntriesForYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry1 := TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 4,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   8,
		Training_hours: 0,
		Vacation_hours: 0,
		Idle_hours:     0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	if err := AddTimesheetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTimesheetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	entries, err := GetTrainingEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Failed to get training entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 training entry, got %d", len(entries))
	}
	if entries[0].Training_hours != 4 {
		t.Errorf("Expected 4 training hours, got %d", entries[0].Training_hours)
	}
}

func TestGetTrainingBudgetEntriesForYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry1 := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}
	entry2 := TrainingBudgetEntry{
		Date:             "2024-02-15",
		Training_name:    "Training B",
		Hours:            4,
		Cost_without_vat: 50.0,
	}

	if err := AddTrainingBudgetEntry(entry1); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := AddTrainingBudgetEntry(entry2); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	entries, err := GetTrainingBudgetEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Failed to get training budget entries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestAddTrainingBudgetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	err := AddTrainingBudgetEntry(entry)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Verify entry was added
	result, err := GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Training_name != "Training A" {
		t.Errorf("Expected Training A, got %s", result.Training_name)
	}
}

func TestUpdateTrainingBudgetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get the entry to get its ID
	result, err := GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	// Update the entry
	result.Training_name = "Training B"
	result.Hours = 10
	err = UpdateTrainingBudgetEntry(result)
	if err != nil {
		t.Fatalf("Failed to update entry: %v", err)
	}

	// Verify update
	updated, err := GetTrainingBudgetEntry(result.Id)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if updated.Training_name != "Training B" {
		t.Errorf("Expected Training B, got %s", updated.Training_name)
	}
	if updated.Hours != 10 {
		t.Errorf("Expected 10 hours, got %d", updated.Hours)
	}
}

func TestDeleteTrainingBudgetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get the entry to get its ID
	result, err := GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	// Delete the entry
	err = DeleteTrainingBudgetEntry(result.Id)
	if err != nil {
		t.Fatalf("Failed to delete entry: %v", err)
	}

	// Verify deletion
	_, err = GetTrainingBudgetEntryByDate("2024-01-15")
	if err == nil {
		t.Error("Expected error for deleted entry")
	}
}

func TestGetTrainingBudgetEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Get all entries to find the ID
	entries, err := GetTrainingBudgetEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("No entries found")
	}

	// Get entry by ID
	result, err := GetTrainingBudgetEntry(entries[0].Id)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Training_name != "Training A" {
		t.Errorf("Expected Training A, got %s", result.Training_name)
	}

	// Test non-existent ID
	_, err = GetTrainingBudgetEntry(999)
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}
}

func TestGetTrainingBudgetEntryByDate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	entry := TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Test getting entry by date
	result, err := GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}
	if result.Training_name != "Training A" {
		t.Errorf("Expected Training A, got %s", result.Training_name)
	}

	// Test non-existent date
	_, err = GetTrainingBudgetEntryByDate("2024-01-16")
	if err == nil {
		t.Error("Expected error for non-existent date")
	}
}
