package db

import (
	"testing"
)

func TestUpsertAndGetBufferEntries(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 2, Hours: 40, Notes: "Feb crunch"}); err != nil {
		t.Fatalf("UpsertBufferEntry: %v", err)
	}
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 4, Hours: 25, Notes: "Apr oncall"}); err != nil {
		t.Fatalf("UpsertBufferEntry: %v", err)
	}

	entries, err := GetBufferEntriesForYear(2026)
	if err != nil {
		t.Fatalf("GetBufferEntriesForYear: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Month != 2 || entries[0].Hours != 40 {
		t.Errorf("expected first entry Feb/40, got %d/%d", entries[0].Month, entries[0].Hours)
	}
	if entries[1].Month != 4 || entries[1].Hours != 25 {
		t.Errorf("expected second entry Apr/25, got %d/%d", entries[1].Month, entries[1].Hours)
	}

	total, err := GetBufferTotalForYear(2026)
	if err != nil {
		t.Fatalf("GetBufferTotalForYear: %v", err)
	}
	if total != 65 {
		t.Errorf("expected total 65, got %d", total)
	}
}

func TestUpsertBufferEntry_UpdatesExisting(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 2, Hours: 40}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 2, Hours: 55, Notes: "corrected"}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	entries, _ := GetBufferEntriesForYear(2026)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after upsert, got %d", len(entries))
	}
	if entries[0].Hours != 55 || entries[0].Notes != "corrected" {
		t.Errorf("entry not updated, got %+v", entries[0])
	}
}

func TestUpsertBufferEntry_RejectsNegative(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 3, Hours: -1})
	if err == nil {
		t.Fatal("expected error for negative hours, got nil")
	}
}

func TestUpsertBufferEntry_RejectsBadMonth(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 0, Hours: 10}); err == nil {
		t.Error("expected error for month 0")
	}
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 13, Hours: 10}); err == nil {
		t.Error("expected error for month 13")
	}
}

func TestDeleteBufferEntry(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)

	_ = UpsertBufferEntry(BufferEntry{Year: 2026, Month: 2, Hours: 40})
	_ = UpsertBufferEntry(BufferEntry{Year: 2026, Month: 4, Hours: 25})

	if err := DeleteBufferEntry(2026, 2); err != nil {
		t.Fatalf("DeleteBufferEntry: %v", err)
	}
	entries, _ := GetBufferEntriesForYear(2026)
	if len(entries) != 1 || entries[0].Month != 4 {
		t.Errorf("expected only Apr after delete, got %+v", entries)
	}
}

// VacationSummary should include BufferHours in TotalAvailable and follow the
// deduction order: carryover → buffer → current-year allowance.
func TestVacationSummary_BufferIncludedAndCascadeOrder(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 187)
	defer cleanup()

	// Explicit carryover for 2026: 20 hours
	if err := SetVacationCarryover(VacationCarryover{
		Year:           2026,
		CarryoverHours: 20,
		SourceYear:     2025,
	}); err != nil {
		t.Fatalf("SetVacationCarryover: %v", err)
	}

	// Buffer banked in 2026: 40 hours (Feb) + 10 hours (Mar) = 50
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 2, Hours: 40}); err != nil {
		t.Fatalf("upsert feb: %v", err)
	}
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 3, Hours: 10}); err != nil {
		t.Fatalf("upsert mar: %v", err)
	}

	// Use 35 vacation hours in 2026. Cascade: 20 (carryover) → 15 (buffer) → 0 (current).
	if err := AddTimesheetEntry(TimesheetEntry{
		Date: "2026-04-01", Client_name: "Vacation", Vacation_hours: 35,
	}); err != nil {
		t.Fatalf("AddTimesheetEntry: %v", err)
	}

	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatalf("GetVacationSummaryForYear: %v", err)
	}

	if summary.CarryoverHours != 20 {
		t.Errorf("CarryoverHours: want 20, got %d", summary.CarryoverHours)
	}
	if summary.BufferHours != 50 {
		t.Errorf("BufferHours: want 50, got %d", summary.BufferHours)
	}
	if summary.TotalAvailable != 187+20+50 {
		t.Errorf("TotalAvailable: want %d, got %d", 187+20+50, summary.TotalAvailable)
	}
	if summary.UsedFromCarryover != 20 {
		t.Errorf("UsedFromCarryover: want 20, got %d", summary.UsedFromCarryover)
	}
	if summary.UsedFromBuffer != 15 {
		t.Errorf("UsedFromBuffer: want 15, got %d", summary.UsedFromBuffer)
	}
	if summary.UsedFromCurrent != 0 {
		t.Errorf("UsedFromCurrent: want 0, got %d", summary.UsedFromCurrent)
	}
	if summary.RemainingTotal != 187+20+50-35 {
		t.Errorf("RemainingTotal: want %d, got %d", 187+20+50-35, summary.RemainingTotal)
	}
}

// When usage exceeds carryover+buffer, the remainder should spill into UsedFromCurrent.
func TestVacationSummary_BufferSpillsToCurrent(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	cleanup := setupTestConfig(t, 100)
	defer cleanup()

	if err := SetVacationCarryover(VacationCarryover{Year: 2026, CarryoverHours: 10, SourceYear: 2025}); err != nil {
		t.Fatal(err)
	}
	if err := UpsertBufferEntry(BufferEntry{Year: 2026, Month: 1, Hours: 20}); err != nil {
		t.Fatal(err)
	}
	if err := AddTimesheetEntry(TimesheetEntry{
		Date: "2026-05-01", Client_name: "Vacation", Vacation_hours: 45,
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := GetVacationSummaryForYear(2026)
	if err != nil {
		t.Fatal(err)
	}
	if summary.UsedFromCarryover != 10 {
		t.Errorf("UsedFromCarryover: want 10, got %d", summary.UsedFromCarryover)
	}
	if summary.UsedFromBuffer != 20 {
		t.Errorf("UsedFromBuffer: want 20, got %d", summary.UsedFromBuffer)
	}
	if summary.UsedFromCurrent != 15 {
		t.Errorf("UsedFromCurrent: want 15, got %d", summary.UsedFromCurrent)
	}
}
