package db

import (
	"testing"
)

// resetRateCacheForTest clears any cache state left by earlier tests so
// each test starts from a known baseline.
func resetRateCacheForTest() {
	rateCacheSnap.Store(nil)
	rateCacheGen.Store(0)
}

func TestBuildRateCache_ReturnsSameInstanceOnRepeatCalls(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatalf("add client: %v", err)
	}
	if err := AddClientRate(ClientRate{ClientId: id, HourlyRate: 100, EffectiveDate: "2024-01-01"}); err != nil {
		t.Fatalf("add rate: %v", err)
	}

	c1, err := buildRateCache()
	if err != nil {
		t.Fatalf("build 1: %v", err)
	}
	c2, err := buildRateCache()
	if err != nil {
		t.Fatalf("build 2: %v", err)
	}
	if c1 != c2 {
		t.Fatalf("expected cache reuse, got fresh instance on second call")
	}
}

func TestBuildRateCache_RebuildsAfterAddClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	c1, err := buildRateCache()
	if err != nil {
		t.Fatalf("build 1: %v", err)
	}
	if _, ok := c1.clientsByName["NewClient"]; ok {
		t.Fatal("precondition: NewClient should not be present yet")
	}

	if _, err := AddClient(Client{Name: "NewClient", IsActive: true}); err != nil {
		t.Fatalf("add client: %v", err)
	}

	c2, err := buildRateCache()
	if err != nil {
		t.Fatalf("build 2: %v", err)
	}
	if c1 == c2 {
		t.Fatal("cache was not invalidated after AddClient")
	}
	if _, ok := c2.clientsByName["NewClient"]; !ok {
		t.Fatal("new client missing from rebuilt cache")
	}
}

func TestBuildRateCache_RebuildsAfterAddClientRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}

	c1, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if len(c1.ratesByClient[id]) != 0 {
		t.Fatal("precondition: no rates yet")
	}

	if err := AddClientRate(ClientRate{ClientId: id, HourlyRate: 200, EffectiveDate: "2024-01-01"}); err != nil {
		t.Fatal(err)
	}

	c2, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if c1 == c2 {
		t.Fatal("cache was not invalidated after AddClientRate")
	}
	if len(c2.ratesByClient[id]) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(c2.ratesByClient[id]))
	}
}

func TestBuildRateCache_RebuildsAfterUpdateClientRate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := AddClientRate(ClientRate{ClientId: id, HourlyRate: 100, EffectiveDate: "2024-01-01"}); err != nil {
		t.Fatal(err)
	}

	c1, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if c1.ratesByClient[id][0].HourlyRate != 100 {
		t.Fatalf("precondition: expected rate 100, got %v", c1.ratesByClient[id][0].HourlyRate)
	}

	rateId := c1.ratesByClient[id][0].Id
	if err := UpdateClientRate(ClientRate{
		Id:            rateId,
		ClientId:      id,
		HourlyRate:    150,
		EffectiveDate: "2024-01-01",
	}); err != nil {
		t.Fatal(err)
	}

	c2, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if c1 == c2 {
		t.Fatal("cache was not invalidated after UpdateClientRate")
	}
	if c2.ratesByClient[id][0].HourlyRate != 150 {
		t.Fatalf("expected updated rate 150, got %v", c2.ratesByClient[id][0].HourlyRate)
	}
}

func TestBuildRateCache_RebuildsAfterDeleteClient(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	id, err := AddClient(Client{Name: "ToDelete", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	c1, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c1.clientsByName["ToDelete"]; !ok {
		t.Fatal("precondition: client should be present")
	}

	if err := DeleteClient(id); err != nil {
		t.Fatal(err)
	}

	c2, err := buildRateCache()
	if err != nil {
		t.Fatal(err)
	}
	if c1 == c2 {
		t.Fatal("cache was not invalidated after DeleteClient")
	}
	if _, ok := c2.clientsByName["ToDelete"]; ok {
		t.Fatal("deleted client still in cache")
	}
}

func TestCalculateEarnings_ReflectsRateUpdate(t *testing.T) {
	dbPath := setupTestDB(t)
	defer teardownTestDB(t, dbPath)
	resetRateCacheForTest()

	id, err := AddClient(Client{Name: "Acme", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := AddClientRate(ClientRate{ClientId: id, HourlyRate: 100, EffectiveDate: "2025-01-01"}); err != nil {
		t.Fatal(err)
	}
	if err := AddTimesheetEntry(TimesheetEntry{
		Date: "2025-06-15", Client_name: "Acme", Client_hours: 8,
	}); err != nil {
		t.Fatal(err)
	}

	first, err := CalculateEarningsForYear(2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.TotalEarnings != 800 {
		t.Fatalf("first earnings: expected 800, got %v", first.TotalEarnings)
	}

	// Add a newer rate that supersedes the earlier one for entries on or after
	// 2025-06-01. The cache must be invalidated for this to take effect.
	if err := AddClientRate(ClientRate{ClientId: id, HourlyRate: 200, EffectiveDate: "2025-06-01"}); err != nil {
		t.Fatal(err)
	}

	second, err := CalculateEarningsForYear(2025)
	if err != nil {
		t.Fatal(err)
	}
	if second.TotalEarnings != 1600 {
		t.Fatalf("second earnings (after rate bump): expected 1600, got %v", second.TotalEarnings)
	}
}
