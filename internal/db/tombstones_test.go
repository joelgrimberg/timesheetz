package db

import (
	"strconv"
	"testing"
)

// tombstoneExists reports whether a (table, key) tombstone is present in the
// SQLite test DB. Used by every Delete* tombstone test below.
func tombstoneExists(t *testing.T, table, key string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM tombstones WHERE table_name = ? AND record_key = ?`,
		table, key,
	).Scan(&n)
	if err != nil {
		t.Fatalf("query tombstone: %v", err)
	}
	return n > 0
}

func TestDeleteTimesheetEntryByDate_WritesTombstone(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	if err := AddTimesheetEntry(TimesheetEntry{
		Date:        "2026-06-14",
		Client_name: "Acme",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := DeleteTimesheetEntryByDate("2026-06-14"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if !tombstoneExists(t, TombstoneTableTimesheet, "2026-06-14") {
		t.Fatal("expected tombstone for timesheet/2026-06-14")
	}
}

func TestDeleteTimesheetEntryByDate_NoRowNoTombstone(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	if err := DeleteTimesheetEntryByDate("2026-06-14"); err != nil {
		t.Fatalf("delete (no row): %v", err)
	}

	if tombstoneExists(t, TombstoneTableTimesheet, "2026-06-14") {
		t.Fatal("did not expect tombstone when no row was deleted")
	}
}

func TestDeleteTimesheetEntry_WritesTombstoneByDate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	if err := AddTimesheetEntry(TimesheetEntry{
		Date:        "2026-06-14",
		Client_name: "Acme",
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	var id int
	if err := db.QueryRow(`SELECT id FROM timesheet WHERE date = ?`, "2026-06-14").Scan(&id); err != nil {
		t.Fatalf("lookup id: %v", err)
	}

	if err := DeleteTimesheetEntry(strconv.Itoa(id)); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if !tombstoneExists(t, TombstoneTableTimesheet, "2026-06-14") {
		t.Fatal("expected tombstone for timesheet/2026-06-14")
	}
}

func TestDeleteVacationCarryover_WritesTombstone(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	if err := SetVacationCarryover(VacationCarryover{
		Year:           2025,
		CarryoverHours: 8,
		SourceYear:     2024,
	}); err != nil {
		t.Fatalf("set carryover: %v", err)
	}

	if err := DeleteVacationCarryover(2025); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if !tombstoneExists(t, TombstoneTableVacationCarryover, TombstoneKeyVacationCarryover(2025)) {
		t.Fatal("expected tombstone for vacation_carryover/2025")
	}
}

func TestDeleteBufferEntry_WritesTombstone(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 6, Hours: 4}); err != nil {
		t.Fatalf("upsert buffer: %v", err)
	}

	if err := DeleteBufferEntry(2026, 6); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if !tombstoneExists(t, TombstoneTableBufferHours, TombstoneKeyBufferHours(2026, 6)) {
		t.Fatal("expected tombstone for buffer_hours/2026-06")
	}
}

func TestDeleteTrainingBudgetEntry_WritesTombstone(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	entry := TrainingBudgetEntry{
		Date:             "2026-06-14",
		Training_name:    "Kubernetes Deep Dive",
		Hours:            8,
		Cost_without_vat: 750.0,
	}
	if err := AddTrainingBudgetEntry(entry); err != nil {
		t.Fatalf("add: %v", err)
	}

	var id int
	if err := db.QueryRow(
		`SELECT id FROM training_budget WHERE date = ? AND training_name = ?`,
		entry.Date, entry.Training_name,
	).Scan(&id); err != nil {
		t.Fatalf("lookup id: %v", err)
	}

	if err := DeleteTrainingBudgetEntry(id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	wantKey := TombstoneKeyTrainingBudget(entry.Date, entry.Training_name)
	if !tombstoneExists(t, TombstoneTableTrainingBudget, wantKey) {
		t.Fatalf("expected tombstone for training_budget/%s", wantKey)
	}
}

func TestDeleteClient_WritesTombstoneAndCascadeRates(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	id, err := AddClient(Client{Name: "Acme"})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}
	if err := AddClientRate(ClientRate{
		ClientId:      id,
		HourlyRate:    100,
		EffectiveDate: "2026-01-01",
		Notes:         "first",
	}); err != nil {
		t.Fatalf("add rate 1: %v", err)
	}
	if err := AddClientRate(ClientRate{
		ClientId:      id,
		HourlyRate:    120,
		EffectiveDate: "2026-07-01",
		Notes:         "raise",
	}); err != nil {
		t.Fatalf("add rate 2: %v", err)
	}

	if err := DeleteClient(id); err != nil {
		t.Fatalf("delete client: %v", err)
	}

	if !tombstoneExists(t, TombstoneTableClients, "Acme") {
		t.Fatal("expected tombstone for clients/Acme")
	}
	for _, d := range []string{"2026-01-01", "2026-07-01"} {
		k := TombstoneKeyClientRate("Acme", d)
		if !tombstoneExists(t, TombstoneTableClientRates, k) {
			t.Fatalf("expected cascaded tombstone for client_rates/%s", k)
		}
	}
}

func TestDeleteClientRate_WritesTombstoneWithCompositeKey(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t, "")

	id, err := AddClient(Client{Name: "Acme"})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}
	if err := AddClientRate(ClientRate{
		ClientId:      id,
		HourlyRate:    100,
		EffectiveDate: "2026-01-01",
	}); err != nil {
		t.Fatalf("add rate: %v", err)
	}

	var rateId int
	if err := db.QueryRow(
		`SELECT id FROM client_rates WHERE client_id = ? AND effective_date = ?`,
		id, "2026-01-01",
	).Scan(&rateId); err != nil {
		t.Fatalf("lookup rate id: %v", err)
	}

	if err := DeleteClientRate(rateId); err != nil {
		t.Fatalf("delete rate: %v", err)
	}

	wantKey := TombstoneKeyClientRate("Acme", "2026-01-01")
	if !tombstoneExists(t, TombstoneTableClientRates, wantKey) {
		t.Fatalf("expected tombstone for client_rates/%s", wantKey)
	}
}
