package sync

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"timesheet/internal/db"
)

// newBenchPair mirrors newSyncPair but uses testing.B.
func newBenchPair(b *testing.B) (*SyncService, *sql.DB, *sql.DB) {
	b.Helper()

	localDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("open local: %v", err)
	}
	b.Cleanup(func() { localDB.Close() })
	if err := db.ApplySQLiteSchema(localDB); err != nil {
		b.Fatalf("init local schema: %v", err)
	}

	remoteDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatalf("open remote: %v", err)
	}
	b.Cleanup(func() { remoteDB.Close() })
	if err := db.ApplySQLiteSchema(remoteDB); err != nil {
		b.Fatalf("init remote schema: %v", err)
	}

	return NewSyncService(localDB, remoteDB, time.Minute), localDB, remoteDB
}

// seedBenchRows inserts N timesheet rows into both sides with matching
// timestamps so a bidirectional Sync becomes a no-op comparison pass.
func seedBenchRows(b *testing.B, local, remote *sql.DB, n int) {
	b.Helper()
	ts := "2024-01-01 00:00:00"
	for i := 0; i < n; i++ {
		date := fmt.Sprintf("2024-%02d-%02d", (i/28)+1, (i%28)+1)
		q := `INSERT INTO timesheet (date, client_name, client_hours, vacation_hours, idle_hours, training_hours, sick_hours, holiday_hours, created_at, updated_at) VALUES (?, ?, 8, 0, 0, 0, 0, 0, ?, ?)`
		if _, err := local.Exec(q, date, "Acme", ts, ts); err != nil {
			b.Fatalf("seed local: %v", err)
		}
		if _, err := remote.Exec(q, date, "Acme", ts, ts); err != nil {
			b.Fatalf("seed remote: %v", err)
		}
	}
}

// BenchmarkSync_Idle measures the cost of a Sync() when both sides already
// agree — this is the common steady-state case and the one that pays the
// biggest price for full-table sync.
func BenchmarkSync_Idle(b *testing.B) {
	svc, local, remote := newBenchPair(b)
	seedBenchRows(b, local, remote, 100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := svc.Sync(SyncBidirectional); err != nil {
			b.Fatal(err)
		}
	}
}
