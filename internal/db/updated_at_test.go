package db

import (
	"testing"
	"time"
)

// readTimestamp returns updated_at as a string for the given row.
func readTimestamp(t *testing.T, table, column, whereCol string, whereVal any) string {
	t.Helper()
	var v string
	q := "SELECT " + column + " FROM " + table + " WHERE " + whereCol + " = ?"
	if err := db.QueryRow(q, whereVal).Scan(&v); err != nil {
		t.Fatalf("read %s.%s: %v", table, column, err)
	}
	return v
}

// TestUpdatedAtBumpedOnTimesheetUpdate verifies that editing a timesheet row
// advances its updated_at column. Sync's conflict resolution depends on this:
// without a fresh updated_at, the writing side cannot win the merge and the
// edit silently fails to propagate.
func TestUpdatedAtBumpedOnTimesheetUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	entry := TimesheetEntry{
		Date:         "2024-01-15",
		Client_name:  "Client A",
		Client_hours: 8,
	}
	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("add: %v", err)
	}

	before := readTimestamp(t, "timesheet", "updated_at", "date", "2024-01-15")

	// CURRENT_TIMESTAMP / our formatted string has 1-second resolution. Sleep
	// past that boundary so a bumped timestamp is observably different.
	time.Sleep(1100 * time.Millisecond)

	entry.Client_hours = 4
	if err := UpdateTimesheetEntry(entry); err != nil {
		t.Fatalf("update: %v", err)
	}

	after := readTimestamp(t, "timesheet", "updated_at", "date", "2024-01-15")
	if after <= before {
		t.Fatalf("updated_at not bumped on timesheet UPDATE: before=%q after=%q", before, after)
	}
}

func TestUpdatedAtBumpedOnTimesheetUpdateById(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	entry := TimesheetEntry{Date: "2024-02-15", Client_name: "X", Client_hours: 8}
	if err := AddTimesheetEntry(entry); err != nil {
		t.Fatalf("add: %v", err)
	}
	row, err := GetTimesheetEntryByDate("2024-02-15")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	before := readTimestamp(t, "timesheet", "updated_at", "id", row.Id)
	time.Sleep(1100 * time.Millisecond)

	if err := UpdateTimesheetEntryById(
		// id is passed as string by the API handler
		intToStr(row.Id),
		map[string]any{"client_hours": 5},
	); err != nil {
		t.Fatalf("update by id: %v", err)
	}

	after := readTimestamp(t, "timesheet", "updated_at", "id", row.Id)
	if after <= before {
		t.Fatalf("updated_at not bumped on UpdateTimesheetEntryById: before=%q after=%q", before, after)
	}
}

func TestUpdatedAtBumpedOnClientUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}

	before := readTimestamp(t, "clients", "updated_at", "id", id)
	time.Sleep(1100 * time.Millisecond)

	if err := UpdateClient(Client{Id: id, Name: "Acme Inc", IsActive: true}); err != nil {
		t.Fatalf("update client: %v", err)
	}

	after := readTimestamp(t, "clients", "updated_at", "id", id)
	if after <= before {
		t.Fatalf("updated_at not bumped on clients UPDATE: before=%q after=%q", before, after)
	}
}

func TestUpdatedAtBumpedOnDeactivateClient(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}

	before := readTimestamp(t, "clients", "updated_at", "id", id)
	time.Sleep(1100 * time.Millisecond)

	if err := DeactivateClient(id); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	after := readTimestamp(t, "clients", "updated_at", "id", id)
	if after <= before {
		t.Fatalf("updated_at not bumped on DeactivateClient: before=%q after=%q", before, after)
	}
}

func TestUpdatedAtBumpedOnClientRateUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}
	if err := AddClientRate(ClientRate{
		ClientId:      id,
		HourlyRate:    100,
		EffectiveDate: "2024-01-01",
	}); err != nil {
		t.Fatalf("add rate: %v", err)
	}
	rates, err := GetClientRates(id)
	if err != nil || len(rates) == 0 {
		t.Fatalf("get rates: %v", err)
	}
	rateId := rates[0].Id

	before := readTimestamp(t, "client_rates", "updated_at", "id", rateId)
	time.Sleep(1100 * time.Millisecond)

	if err := UpdateClientRate(ClientRate{
		Id:            rateId,
		HourlyRate:    125,
		EffectiveDate: "2024-01-01",
	}); err != nil {
		t.Fatalf("update rate: %v", err)
	}

	after := readTimestamp(t, "client_rates", "updated_at", "id", rateId)
	if after <= before {
		t.Fatalf("updated_at not bumped on client_rates UPDATE: before=%q after=%q", before, after)
	}
}

func TestUpdatedAtBumpedOnTrainingBudgetUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	entry := TrainingBudgetEntry{
		Date:             "2024-03-01",
		Training_name:    "Kubernetes",
		Hours:            8,
		Cost_without_vat: 500,
	}
	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("add training budget: %v", err)
	}
	got, err := GetTrainingBudgetEntryByDate("2024-03-01")
	if err != nil {
		t.Fatalf("get training budget: %v", err)
	}

	before := readTimestamp(t, "training_budget", "updated_at", "id", got.Id)
	time.Sleep(1100 * time.Millisecond)

	got.Cost_without_vat = 600
	if err := UpdateTrainingBudgetEntry(got); err != nil {
		t.Fatalf("update training budget: %v", err)
	}

	after := readTimestamp(t, "training_budget", "updated_at", "id", got.Id)
	if after <= before {
		t.Fatalf("updated_at not bumped on training_budget UPDATE: before=%q after=%q", before, after)
	}
}

func intToStr(i int) string {
	// Local helper so the test file doesn't depend on strconv elsewhere.
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = digits[i%10]
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
