package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"timesheet/internal/config"
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
	defer Close()

	// InitializeDatabase already opens the connection, so we don't need to call Connect
	// Verify connection is working
	if err := Ping(); err != nil {
		t.Fatalf("Failed to ping database after initialization: %v", err)
	}

	// Verify indexes were created
	expectedIndexes := []string{
		"idx_client_name",
		"idx_timesheet_date",
		"idx_timesheet_date_client",
		"idx_training_date",
		"idx_clients_name",
		"idx_clients_active",
		"idx_client_rates_client",
		"idx_client_rates_date",
		"idx_client_rates_client_date",
		"idx_vacation_carryover_year",
	}

	for _, indexName := range expectedIndexes {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check for index %s: %v", indexName, err)
		}
		if count == 0 {
			t.Errorf("Expected index %s was not created", indexName)
		}
	}
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

// setupTestConfig creates a temporary config file with a given yearly target
// and returns a cleanup function.
func setupTestConfig(t *testing.T, yearlyTarget int) func() {
	t.Helper()
	tmpDir := t.TempDir()
	tmpConfigPath := filepath.Join(tmpDir, "config.json")
	testConfig := config.Config{
		VacationHours: config.VacationHours{
			YearlyTarget: yearlyTarget,
			Category:     "Vacation",
		},
	}
	config.SetConfigPathOverride(tmpConfigPath)
	if err := config.SaveConfig(testConfig); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	return func() {
		config.SetConfigPathOverride("")
		os.RemoveAll(tmpDir)
	}
}

func TestAutoCarryover_FromPreviousYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 187)
	defer cleanup()

	// Add 140 vacation hours in 2025
	entries := []TimesheetEntry{
		{Date: "2025-01-15", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-01-16", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-01-17", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-08-11", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-08-12", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-08-13", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-08-14", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-08-15", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-22", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-23", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-24", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-29", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-30", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-12-31", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-06-15", Client_name: "Vacation", Vacation_hours: 9},
		{Date: "2025-06-16", Client_name: "Vacation", Vacation_hours: 5}, // total = 140
	}
	for _, e := range entries {
		if err := AddTimesheetEntry(e); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	// Verify 2025 used hours
	used, err := GetVacationHoursForYear(2025)
	if err != nil {
		t.Fatalf("Failed to get 2025 vacation hours: %v", err)
	}
	if used != 140 {
		t.Fatalf("Expected 140 used hours in 2025, got %d", used)
	}

	// Get 2026 summary — no explicit carryover record exists, should auto-calculate
	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get 2026 vacation summary: %v", err)
	}

	// 2025 remaining = 187 - 140 = 47 (no carryover into 2025)
	expectedCarryover := 47
	if summary.CarryoverHours != expectedCarryover {
		t.Errorf("Expected auto-carryover of %d, got %d", expectedCarryover, summary.CarryoverHours)
	}
	if summary.TotalAvailable != 187+expectedCarryover {
		t.Errorf("Expected total available %d, got %d", 187+expectedCarryover, summary.TotalAvailable)
	}
}

func TestAutoCarryover_WithExplicitPrevYearCarryover(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 187)
	defer cleanup()

	// Set explicit carryover for 2025 (from 2024)
	err := SetVacationCarryover(VacationCarryover{
		Year:           2025,
		CarryoverHours: 14,
		SourceYear:     2024,
		Notes:          "Carryover from 2024",
	})
	if err != nil {
		t.Fatalf("Failed to set carryover: %v", err)
	}

	// Add 143 vacation hours in 2025
	for i := 0; i < 15; i++ {
		hours := 9
		if i == 14 {
			hours = 8 // 14*9 + 8 = 134... need 143
		}
		entry := TimesheetEntry{
			Date:           "2025-" + strconv.Itoa(i/28+1) + "-" + strconv.Itoa(i%28+1),
			Client_name:    "Vacation",
			Vacation_hours: hours,
		}
		if err := AddTimesheetEntry(entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	// That gives us 14*9 + 8 = 134, we need 143. Let me fix the math.
	// Actually let me just add precise entries.
	// Clean up and redo.
	db.Exec("DELETE FROM timesheet")

	// 15 entries of 9 = 135, plus one of 8 = 143
	for i := 0; i < 15; i++ {
		date := "2025-" + fmt.Sprintf("%02d", i/28+1) + "-" + fmt.Sprintf("%02d", i%28+1)
		entry := TimesheetEntry{
			Date:           date,
			Client_name:    "Vacation",
			Vacation_hours: 9,
		}
		if err := AddTimesheetEntry(entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}
	// Add one more entry with 8 hours: 15*9 + 8 = 143
	if err := AddTimesheetEntry(TimesheetEntry{
		Date: "2025-02-01", Client_name: "Vacation", Vacation_hours: 8,
	}); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	used, _ := GetVacationHoursForYear(2025)
	if used != 143 {
		t.Fatalf("Expected 143 used hours, got %d", used)
	}

	// 2026 auto-carryover: 187 + 14 (explicit 2025 carryover) - 143 = 58
	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get 2026 summary: %v", err)
	}

	if summary.CarryoverHours != 58 {
		t.Errorf("Expected auto-carryover of 58, got %d", summary.CarryoverHours)
	}
}

func TestAutoCarryover_ExplicitOverridesAuto(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 187)
	defer cleanup()

	// Add some vacation in 2025
	if err := AddTimesheetEntry(TimesheetEntry{
		Date: "2025-06-15", Client_name: "Vacation", Vacation_hours: 9,
	}); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Auto-carryover for 2026 would be 187 - 9 = 178
	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}
	if summary.CarryoverHours != 178 {
		t.Errorf("Expected auto-carryover of 178, got %d", summary.CarryoverHours)
	}

	// Now set explicit carryover that overrides auto-calculation
	err = SetVacationCarryover(VacationCarryover{
		Year:           2026,
		CarryoverHours: 50,
		SourceYear:     2025,
		Notes:          "Manual override",
	})
	if err != nil {
		t.Fatalf("Failed to set carryover: %v", err)
	}

	// Should now use the explicit value, not auto-calculated
	summary, err = GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}
	if summary.CarryoverHours != 50 {
		t.Errorf("Expected explicit carryover of 50, got %d", summary.CarryoverHours)
	}
}

func TestAutoCarryover_NegativeRemainingClampsToZero(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 50)
	defer cleanup()

	// Use more vacation than available in 2025 (60 > 50)
	for i := 0; i < 6; i++ {
		if err := AddTimesheetEntry(TimesheetEntry{
			Date:           "2025-0" + strconv.Itoa(i+1) + "-15",
			Client_name:    "Vacation",
			Vacation_hours: 10,
		}); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	// 2026 auto-carryover: 50 - 60 = -10, should clamp to 0
	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}
	if summary.CarryoverHours != 0 {
		t.Errorf("Expected 0 carryover (negative clamped), got %d", summary.CarryoverHours)
	}
}

func TestAutoCarryover_NoPreviousYearData(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 187)
	defer cleanup()

	// No entries for 2025 at all — remaining = 187, carryover into 2026 = 187
	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}
	if summary.CarryoverHours != 187 {
		t.Errorf("Expected 187 carryover (full unused year), got %d", summary.CarryoverHours)
	}
}
