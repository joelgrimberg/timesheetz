package db

import (
	"fmt"
	"testing"
	"time"
)

// seedYearOfEntries populates the timesheet, clients, and client_rates
// tables with one work-year of data for benchmarking earnings/rate paths.
func seedYearOfEntries(b *testing.B, year int, clientCount int) {
	b.Helper()

	for i := 0; i < clientCount; i++ {
		name := fmt.Sprintf("Client-%03d", i)
		id, err := AddClient(Client{Name: name, IsActive: true})
		if err != nil {
			b.Fatalf("seed client: %v", err)
		}
		if err := AddClientRate(ClientRate{
			ClientId:      id,
			HourlyRate:    100.0 + float64(i),
			EffectiveDate: fmt.Sprintf("%d-01-01", year),
			Notes:         "seed",
		}); err != nil {
			b.Fatalf("seed rate: %v", err)
		}
	}

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	for d := 0; d < 365; d++ {
		date := start.AddDate(0, 0, d).Format("2006-01-02")
		clientName := fmt.Sprintf("Client-%03d", d%clientCount)
		if err := AddTimesheetEntry(TimesheetEntry{
			Date:         date,
			Client_name:  clientName,
			Client_hours: 8,
		}); err != nil {
			b.Fatalf("seed entry: %v", err)
		}
	}
}

func BenchmarkBuildRateCache(b *testing.B) {
	if err := InitializeDatabase(":memory:"); err != nil {
		b.Fatalf("init db: %v", err)
	}
	defer Close()
	seedYearOfEntries(b, 2025, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cache, err := buildRateCache()
		if err != nil {
			b.Fatal(err)
		}
		_ = cache
	}
}

func BenchmarkCalculateEarningsForYear(b *testing.B) {
	if err := InitializeDatabase(":memory:"); err != nil {
		b.Fatalf("init db: %v", err)
	}
	defer Close()
	seedYearOfEntries(b, 2025, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := CalculateEarningsForYear(2025); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAllTimesheetEntries(b *testing.B) {
	if err := InitializeDatabase(":memory:"); err != nil {
		b.Fatalf("init db: %v", err)
	}
	defer Close()
	seedYearOfEntries(b, 2025, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		entries, err := GetAllTimesheetEntries(0, 0)
		if err != nil {
			b.Fatal(err)
		}
		_ = entries
	}
}
