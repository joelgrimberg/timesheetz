package db

import (
	"testing"
	"time"
)

// TestEndToEndClientRatesAndEarnings tests the complete flow:
// 1. Create a client
// 2. Add rates with different effective dates
// 3. Add timesheet entries
// 4. Calculate earnings with correct historical rates
func TestEndToEndClientRatesAndEarnings(t *testing.T) {
	// Setup test database
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// 1. Create a client
	client := Client{
		Name:     "Acme Corp",
		IsActive: true,
	}
	clientId, err := AddClient(client)
	if err != nil {
		t.Fatalf("Failed to add client: %v", err)
	}

	// 2. Add rates with different effective dates
	rate1 := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Initial rate",
	}
	rate2 := ClientRate{
		ClientId:      clientId,
		HourlyRate:    120.00,
		EffectiveDate: "2024-07-01",
		Notes:         "Mid-year increase",
	}
	rate3 := ClientRate{
		ClientId:      clientId,
		HourlyRate:    150.00,
		EffectiveDate: "2025-01-01",
		Notes:         "2025 rate",
	}

	if err := AddClientRate(rate1); err != nil {
		t.Fatalf("Failed to add rate1: %v", err)
	}
	if err := AddClientRate(rate2); err != nil {
		t.Fatalf("Failed to add rate2: %v", err)
	}
	if err := AddClientRate(rate3); err != nil {
		t.Fatalf("Failed to add rate3: %v", err)
	}

	// 3. Add timesheet entries across different rate periods
	entries := []TimesheetEntry{
		{
			Date:         "2024-03-15", // Should use €100 rate
			Client_name:  "Acme Corp",
			Client_hours: 8,
		},
		{
			Date:         "2024-08-15", // Should use €120 rate
			Client_name:  "Acme Corp",
			Client_hours: 8,
		},
		{
			Date:         "2025-02-15", // Should use €150 rate
			Client_name:  "Acme Corp",
			Client_hours: 8,
		},
	}

	for _, entry := range entries {
		if err := AddTimesheetEntry(entry); err != nil {
			t.Fatalf("Failed to add timesheet entry: %v", err)
		}
	}

	// 4. Calculate earnings for 2024
	overview2024, err := CalculateEarningsForYear(2024)
	if err != nil {
		t.Fatalf("Failed to calculate 2024 earnings: %v", err)
	}

	// Verify 2024 results
	if overview2024.Year != 2024 {
		t.Errorf("Expected year 2024, got %d", overview2024.Year)
	}
	if overview2024.TotalHours != 16 {
		t.Errorf("Expected 16 total hours in 2024, got %d", overview2024.TotalHours)
	}
	// 8 hours * €100 + 8 hours * €120 = €800 + €960 = €1760
	expectedEarnings2024 := 1760.00
	if overview2024.TotalEarnings != expectedEarnings2024 {
		t.Errorf("Expected €%.2f earnings in 2024, got €%.2f", expectedEarnings2024, overview2024.TotalEarnings)
	}

	// Verify individual entries have correct rates
	if len(overview2024.Entries) != 2 {
		t.Fatalf("Expected 2 entries in 2024, got %d", len(overview2024.Entries))
	}

	// Find March entry (should be €100/hour)
	var marchEntry *EarningsEntry
	for i := range overview2024.Entries {
		if overview2024.Entries[i].Date == "2024-03-15" {
			marchEntry = &overview2024.Entries[i]
			break
		}
	}
	if marchEntry == nil {
		t.Fatal("March entry not found")
	}
	if marchEntry.HourlyRate != 100.00 {
		t.Errorf("Expected March rate €100, got €%.2f", marchEntry.HourlyRate)
	}
	if marchEntry.Earnings != 800.00 {
		t.Errorf("Expected March earnings €800, got €%.2f", marchEntry.Earnings)
	}

	// Find August entry (should be €120/hour)
	var augustEntry *EarningsEntry
	for i := range overview2024.Entries {
		if overview2024.Entries[i].Date == "2024-08-15" {
			augustEntry = &overview2024.Entries[i]
			break
		}
	}
	if augustEntry == nil {
		t.Fatal("August entry not found")
	}
	if augustEntry.HourlyRate != 120.00 {
		t.Errorf("Expected August rate €120, got €%.2f", augustEntry.HourlyRate)
	}
	if augustEntry.Earnings != 960.00 {
		t.Errorf("Expected August earnings €960, got €%.2f", augustEntry.Earnings)
	}

	// 5. Calculate earnings for 2025
	overview2025, err := CalculateEarningsForYear(2025)
	if err != nil {
		t.Fatalf("Failed to calculate 2025 earnings: %v", err)
	}

	// Verify 2025 results
	if overview2025.Year != 2025 {
		t.Errorf("Expected year 2025, got %d", overview2025.Year)
	}
	if overview2025.TotalHours != 8 {
		t.Errorf("Expected 8 total hours in 2025, got %d", overview2025.TotalHours)
	}
	// 8 hours * €150 = €1200
	expectedEarnings2025 := 1200.00
	if overview2025.TotalEarnings != expectedEarnings2025 {
		t.Errorf("Expected €%.2f earnings in 2025, got €%.2f", expectedEarnings2025, overview2025.TotalEarnings)
	}

	// 6. Test monthly earnings calculation
	monthlyOverview, err := CalculateEarningsForMonth(2024, 8)
	if err != nil {
		t.Fatalf("Failed to calculate monthly earnings: %v", err)
	}

	if monthlyOverview.Month != 8 {
		t.Errorf("Expected month 8, got %d", monthlyOverview.Month)
	}
	if monthlyOverview.TotalHours != 8 {
		t.Errorf("Expected 8 hours in August, got %d", monthlyOverview.TotalHours)
	}
	if monthlyOverview.TotalEarnings != 960.00 {
		t.Errorf("Expected €960 in August, got €%.2f", monthlyOverview.TotalEarnings)
	}
}

// TestRateLookupEdgeCases tests edge cases in rate lookup logic
func TestRateLookupEdgeCases(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Create client
	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	// Add rate with future effective date
	futureRate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    200.00,
		EffectiveDate: "2030-01-01",
		Notes:         "Future rate",
	}
	AddClientRate(futureRate)

	// Try to get rate for today (should return 0.0 since future rate shouldn't apply)
	today := time.Now().Format("2006-01-02")
	rate, err := GetClientRateForDate(clientId, today)

	// Should return error since no rate is effective yet
	if err == nil {
		t.Error("Expected error when no rate is effective, but got nil")
	}

	// Add a current rate
	currentRate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Current rate",
	}
	AddClientRate(currentRate)

	// Now should get the current rate (future rate should be ignored)
	rate, err = GetClientRateForDate(clientId, today)
	if err != nil {
		t.Fatalf("Failed to get rate: %v", err)
	}
	if rate.HourlyRate != 100.00 {
		t.Errorf("Expected rate €100, got €%.2f (future rate should be ignored)", rate.HourlyRate)
	}
}

// TestClientWithoutRatesEarnings tests earnings calculation when client has no rates
func TestClientWithoutRatesEarnings(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Create client without any rates
	AddClient(Client{Name: "No Rates Client", IsActive: true})

	// Add timesheet entry
	AddTimesheetEntry(TimesheetEntry{
		Date:         "2024-01-15",
		Client_name:  "No Rates Client",
		Client_hours: 8,
	})

	// Calculate earnings
	overview, err := CalculateEarningsForYear(2024)
	if err != nil {
		t.Fatalf("Failed to calculate earnings: %v", err)
	}

	// Should have entry with 0 earnings
	if len(overview.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(overview.Entries))
	}
	if overview.Entries[0].HourlyRate != 0.0 {
		t.Errorf("Expected rate €0, got €%.2f", overview.Entries[0].HourlyRate)
	}
	if overview.Entries[0].Earnings != 0.0 {
		t.Errorf("Expected earnings €0, got €%.2f", overview.Entries[0].Earnings)
	}
	if overview.TotalEarnings != 0.0 {
		t.Errorf("Expected total earnings €0, got €%.2f", overview.TotalEarnings)
	}
}
