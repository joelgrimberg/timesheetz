package db

import (
	"testing"
	"time"
)

func TestAddClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	client := Client{Name: "Acme Corp", IsActive: true}
	id, err := AddClient(client)
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Verify client can be retrieved
	retrieved, err := GetClientById(id)
	if err != nil {
		t.Fatalf("GetClientById failed: %v", err)
	}
	if retrieved.Name != "Acme Corp" {
		t.Errorf("Expected name 'Acme Corp', got '%s'", retrieved.Name)
	}
	if !retrieved.IsActive {
		t.Errorf("Expected client to be active")
	}
}

func TestGetClientByName(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	client := Client{Name: "Test Client", IsActive: true}
	_, err := AddClient(client)
	if err != nil {
		t.Fatalf("AddClient failed: %v", err)
	}

	retrieved, err := GetClientByName("Test Client")
	if err != nil {
		t.Fatalf("GetClientByName failed: %v", err)
	}
	if retrieved.Name != "Test Client" {
		t.Errorf("Expected name 'Test Client', got '%s'", retrieved.Name)
	}
}

func TestGetAllClients(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add multiple clients
	clients := []string{"Client A", "Client B", "Client C"}
	for _, name := range clients {
		_, err := AddClient(Client{Name: name, IsActive: true})
		if err != nil {
			t.Fatalf("AddClient failed: %v", err)
		}
	}

	allClients, err := GetAllClients()
	if err != nil {
		t.Fatalf("GetAllClients failed: %v", err)
	}
	if len(allClients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(allClients))
	}
}

func TestGetActiveClients(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add active and inactive clients
	id1, _ := AddClient(Client{Name: "Active Client", IsActive: true})
	id2, _ := AddClient(Client{Name: "Inactive Client", IsActive: false})
	_, _ = id1, id2

	activeClients, err := GetActiveClients()
	if err != nil {
		t.Fatalf("GetActiveClients failed: %v", err)
	}
	if len(activeClients) != 1 {
		t.Errorf("Expected 1 active client, got %d", len(activeClients))
	}
	if activeClients[0].Name != "Active Client" {
		t.Errorf("Expected 'Active Client', got '%s'", activeClients[0].Name)
	}
}

func TestUpdateClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	id, _ := AddClient(Client{Name: "Original Name", IsActive: true})

	client, _ := GetClientById(id)
	client.Name = "Updated Name"
	client.IsActive = false

	err := UpdateClient(client)
	if err != nil {
		t.Fatalf("UpdateClient failed: %v", err)
	}

	updated, _ := GetClientById(id)
	if updated.Name != "Updated Name" {
		t.Errorf("Expected 'Updated Name', got '%s'", updated.Name)
	}
	if updated.IsActive {
		t.Errorf("Expected client to be inactive")
	}
}

func TestDeactivateClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	id, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	err := DeactivateClient(id)
	if err != nil {
		t.Fatalf("DeactivateClient failed: %v", err)
	}

	client, _ := GetClientById(id)
	if client.IsActive {
		t.Errorf("Expected client to be inactive")
	}
}

func TestDeleteClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	id, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	err := DeleteClient(id)
	if err != nil {
		t.Fatalf("DeleteClient failed: %v", err)
	}

	// Verify client is deleted
	_, err = GetClientById(id)
	if err == nil {
		t.Errorf("Expected error when getting deleted client")
	}
}

// Client Rate Tests

func TestAddClientRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	rate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.50,
		EffectiveDate: "2024-01-01",
		Notes:         "Initial rate",
	}

	err := AddClientRate(rate)
	if err != nil {
		t.Fatalf("AddClientRate failed: %v", err)
	}

	// Verify rate can be retrieved
	rates, err := GetClientRates(clientId)
	if err != nil {
		t.Fatalf("GetClientRates failed: %v", err)
	}
	if len(rates) != 1 {
		t.Errorf("Expected 1 rate, got %d", len(rates))
	}
	if rates[0].HourlyRate != 100.50 {
		t.Errorf("Expected rate 100.50, got %.2f", rates[0].HourlyRate)
	}
}

func TestGetClientRateForDate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	// Add multiple rates with different effective dates
	rates := []ClientRate{
		{ClientId: clientId, HourlyRate: 50.00, EffectiveDate: "2024-01-01", Notes: "Rate 1"},
		{ClientId: clientId, HourlyRate: 60.00, EffectiveDate: "2024-06-01", Notes: "Rate 2"},
		{ClientId: clientId, HourlyRate: 70.00, EffectiveDate: "2024-12-01", Notes: "Rate 3"},
	}

	for _, rate := range rates {
		err := AddClientRate(rate)
		if err != nil {
			t.Fatalf("AddClientRate failed: %v", err)
		}
	}

	// Test: Get rate for date before any rates (should error)
	_, err := GetClientRateForDate(clientId, "2023-12-31")
	if err == nil {
		t.Errorf("Expected error for date before any rates")
	}

	// Test: Get rate for date in first period (should return 50.00)
	rate1, err := GetClientRateForDate(clientId, "2024-05-15")
	if err != nil {
		t.Fatalf("GetClientRateForDate failed: %v", err)
	}
	if rate1.HourlyRate != 50.00 {
		t.Errorf("Expected rate 50.00, got %.2f", rate1.HourlyRate)
	}

	// Test: Get rate for date in second period (should return 60.00)
	rate2, err := GetClientRateForDate(clientId, "2024-08-01")
	if err != nil {
		t.Fatalf("GetClientRateForDate failed: %v", err)
	}
	if rate2.HourlyRate != 60.00 {
		t.Errorf("Expected rate 60.00, got %.2f", rate2.HourlyRate)
	}

	// Test: Get rate for date in third period (should return 70.00)
	rate3, err := GetClientRateForDate(clientId, "2024-12-15")
	if err != nil {
		t.Fatalf("GetClientRateForDate failed: %v", err)
	}
	if rate3.HourlyRate != 70.00 {
		t.Errorf("Expected rate 70.00, got %.2f", rate3.HourlyRate)
	}

	// Test: Get rate for exact effective date (should return that rate)
	rateExact, err := GetClientRateForDate(clientId, "2024-06-01")
	if err != nil {
		t.Fatalf("GetClientRateForDate failed: %v", err)
	}
	if rateExact.HourlyRate != 60.00 {
		t.Errorf("Expected rate 60.00 for exact date, got %.2f", rateExact.HourlyRate)
	}
}

func TestGetClientRateByName(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	rate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Test rate",
	}
	AddClientRate(rate)

	// Test: Get rate by client name
	hourlyRate, err := GetClientRateByName("Test Client", "2024-06-01")
	if err != nil {
		t.Fatalf("GetClientRateByName failed: %v", err)
	}
	if hourlyRate != 100.00 {
		t.Errorf("Expected rate 100.00, got %.2f", hourlyRate)
	}

	// Test: Get rate for non-existent client (should return 0)
	hourlyRate2, _ := GetClientRateByName("Non-Existent Client", "2024-06-01")
	if hourlyRate2 != 0.0 {
		t.Errorf("Expected rate 0.00 for non-existent client, got %.2f", hourlyRate2)
	}
}

func TestUpdateClientRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	rate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Original note",
	}
	AddClientRate(rate)

	rates, _ := GetClientRates(clientId)
	rateToUpdate := rates[0]
	rateToUpdate.HourlyRate = 150.00
	rateToUpdate.Notes = "Updated note"

	err := UpdateClientRate(rateToUpdate)
	if err != nil {
		t.Fatalf("UpdateClientRate failed: %v", err)
	}

	updated, _ := GetClientRateById(rateToUpdate.Id)
	if updated.HourlyRate != 150.00 {
		t.Errorf("Expected rate 150.00, got %.2f", updated.HourlyRate)
	}
	if updated.Notes != "Updated note" {
		t.Errorf("Expected 'Updated note', got '%s'", updated.Notes)
	}
}

func TestDeleteClientRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})

	rate := ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	}
	AddClientRate(rate)

	rates, _ := GetClientRates(clientId)
	rateId := rates[0].Id

	err := DeleteClientRate(rateId)
	if err != nil {
		t.Fatalf("DeleteClientRate failed: %v", err)
	}

	// Verify rate is deleted
	_, err = GetClientRateById(rateId)
	if err == nil {
		t.Errorf("Expected error when getting deleted rate")
	}
}

// Earnings Calculation Tests

func TestCalculateEarningsForYear(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add client with rate
	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})
	AddClientRate(ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	})

	// Add timesheet entries
	entries := []TimesheetEntry{
		{Date: "2024-01-15", Client_name: "Test Client", Client_hours: 8},
		{Date: "2024-02-15", Client_name: "Test Client", Client_hours: 10},
		{Date: "2024-03-15", Client_name: "Test Client", Client_hours: 5},
	}

	for _, entry := range entries {
		AddTimesheetEntry(entry)
	}

	// Calculate earnings
	earnings, err := CalculateEarningsForYear(2024)
	if err != nil {
		t.Fatalf("CalculateEarningsForYear failed: %v", err)
	}

	expectedHours := 23
	expectedEarnings := 2300.00

	if earnings.TotalHours != expectedHours {
		t.Errorf("Expected %d hours, got %d", expectedHours, earnings.TotalHours)
	}
	if earnings.TotalEarnings != expectedEarnings {
		t.Errorf("Expected earnings %.2f, got %.2f", expectedEarnings, earnings.TotalEarnings)
	}
	if len(earnings.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(earnings.Entries))
	}
}

func TestCalculateEarningsWithRateChange(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add client with multiple rates
	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})
	AddClientRate(ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	})
	AddClientRate(ClientRate{
		ClientId:      clientId,
		HourlyRate:    150.00,
		EffectiveDate: "2024-07-01",
	})

	// Add timesheet entries before and after rate change
	entries := []TimesheetEntry{
		{Date: "2024-05-15", Client_name: "Test Client", Client_hours: 10}, // Should use 100.00 rate
		{Date: "2024-08-15", Client_name: "Test Client", Client_hours: 10}, // Should use 150.00 rate
	}

	for _, entry := range entries {
		AddTimesheetEntry(entry)
	}

	// Calculate earnings
	earnings, err := CalculateEarningsForYear(2024)
	if err != nil {
		t.Fatalf("CalculateEarningsForYear failed: %v", err)
	}

	// May entry: 10 * 100 = 1000
	// August entry: 10 * 150 = 1500
	// Total: 2500
	expectedEarnings := 2500.00

	if earnings.TotalEarnings != expectedEarnings {
		t.Errorf("Expected earnings %.2f, got %.2f", expectedEarnings, earnings.TotalEarnings)
	}

	// Verify individual entry rates
	for _, entry := range earnings.Entries {
		if entry.Date == "2024-05-15" && entry.HourlyRate != 100.00 {
			t.Errorf("Expected rate 100.00 for May entry, got %.2f", entry.HourlyRate)
		}
		if entry.Date == "2024-08-15" && entry.HourlyRate != 150.00 {
			t.Errorf("Expected rate 150.00 for August entry, got %.2f", entry.HourlyRate)
		}
	}
}

func TestCalculateEarningsForMonth(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add client with rate
	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})
	AddClientRate(ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	})

	// Add timesheet entries for different months
	entries := []TimesheetEntry{
		{Date: "2024-01-15", Client_name: "Test Client", Client_hours: 8},
		{Date: "2024-02-15", Client_name: "Test Client", Client_hours: 10},
		{Date: "2024-02-20", Client_name: "Test Client", Client_hours: 5},
	}

	for _, entry := range entries {
		AddTimesheetEntry(entry)
	}

	// Calculate earnings for February only
	earnings, err := CalculateEarningsForMonth(2024, int(time.February))
	if err != nil {
		t.Fatalf("CalculateEarningsForMonth failed: %v", err)
	}

	expectedHours := 15    // 10 + 5
	expectedEarnings := 1500.00 // 15 * 100

	if earnings.TotalHours != expectedHours {
		t.Errorf("Expected %d hours, got %d", expectedHours, earnings.TotalHours)
	}
	if earnings.TotalEarnings != expectedEarnings {
		t.Errorf("Expected earnings %.2f, got %.2f", expectedEarnings, earnings.TotalEarnings)
	}
	if earnings.Month != int(time.February) {
		t.Errorf("Expected month %d, got %d", time.February, earnings.Month)
	}
}

func TestEarningsWithNoRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add timesheet entry for client WITHOUT a rate
	entry := TimesheetEntry{Date: "2024-01-15", Client_name: "Client Without Rate", Client_hours: 8}
	AddTimesheetEntry(entry)

	// Calculate earnings
	earnings, err := CalculateEarningsForYear(2024)
	if err != nil {
		t.Fatalf("CalculateEarningsForYear failed: %v", err)
	}

	// Should have 0 earnings for client without rate
	if earnings.TotalEarnings != 0.00 {
		t.Errorf("Expected 0.00 earnings for client without rate, got %.2f", earnings.TotalEarnings)
	}
}

func TestGetClientWithRates(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	// Add client with multiple rates
	clientId, _ := AddClient(Client{Name: "Test Client", IsActive: true})
	rates := []ClientRate{
		{ClientId: clientId, HourlyRate: 100.00, EffectiveDate: "2024-01-01"},
		{ClientId: clientId, HourlyRate: 150.00, EffectiveDate: "2024-07-01"},
	}

	for _, rate := range rates {
		AddClientRate(rate)
	}

	// Get client with rates
	clientWithRates, err := GetClientWithRates(clientId)
	if err != nil {
		t.Fatalf("GetClientWithRates failed: %v", err)
	}

	if clientWithRates.Name != "Test Client" {
		t.Errorf("Expected name 'Test Client', got '%s'", clientWithRates.Name)
	}
	if len(clientWithRates.Rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(clientWithRates.Rates))
	}
}
