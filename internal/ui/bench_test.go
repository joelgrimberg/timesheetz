package ui

import (
	"testing"
	"time"

	"timesheet/internal/db"
)

// setupBenchDB wires the shared db package to an in-memory SQLite and
// seeds one month of timesheet rows so the UI models have data to render.
func setupBenchDB(b *testing.B) {
	b.Helper()
	if err := db.InitializeDatabase(":memory:"); err != nil {
		b.Fatalf("init db: %v", err)
	}
	b.Cleanup(func() { db.Close() })

	if _, err := db.AddClient(db.Client{Name: "Acme", IsActive: true}); err != nil {
		b.Fatalf("seed client: %v", err)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for d := 0; d < 28; d++ {
		date := start.AddDate(0, 0, d).Format("2006-01-02")
		if err := db.AddTimesheetEntry(db.TimesheetEntry{
			Date:         date,
			Client_name:  "Acme",
			Client_hours: 8,
		}); err != nil {
			b.Fatalf("seed entry %d: %v", d, err)
		}
	}
}

func BenchmarkGenerateMonthTable(b *testing.B) {
	setupBenchDB(b)
	now := time.Now()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := generateMonthTable(now.Year(), now.Month())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTimesheetView(b *testing.B) {
	setupBenchDB(b)
	m := InitialTimesheetModel()

	// Render once outside the timer to warm any lazy state, then measure
	// the steady-state per-frame cost.
	_ = m.View()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := m.View()
		if len(s) == 0 {
			b.Fatal("empty view")
		}
	}
}

