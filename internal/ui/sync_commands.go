package ui

import (
	"fmt"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/sync"

	tea "github.com/charmbracelet/bubbletea"
)

const syncInterval = 1 * time.Minute

// InitSyncServiceCmd initializes the sync service if both databases are available
// Returns nil if sync is not possible (no postgres URL configured)
func InitSyncServiceCmd() tea.Cmd {
	return func() tea.Msg {
		// Check if PostgreSQL is configured
		postgresURL := config.GetPostgresURL()
		if postgresURL == "" {
			return syncInitResultMsg{enabled: false}
		}

		// Get the SQLite connection (always need this for sync)
		// If running in postgres mode, SQLite may not be connected yet
		sqliteDB := db.GetSQLiteDB()
		if sqliteDB == nil {
			// Try to connect to SQLite
			dbPath := db.GetDBPath()
			if err := db.Connect(dbPath); err != nil {
				return syncInitResultMsg{enabled: false, err: "Failed to connect to SQLite: " + err.Error()}
			}
			// Initialize SQLite schema to ensure sync columns exist
			if err := db.InitializeDatabase(dbPath); err != nil {
				return syncInitResultMsg{enabled: false, err: "Failed to initialize SQLite: " + err.Error()}
			}
			sqliteDB = db.GetSQLiteDB()
		}

		if sqliteDB == nil {
			return syncInitResultMsg{enabled: false, err: "SQLite database not connected"}
		}

		// Try to connect to PostgreSQL if not already connected
		postgresDB := db.GetPostgresDB()
		if postgresDB == nil {
			// Try to connect
			if err := db.ConnectPostgres(postgresURL); err != nil {
				return syncInitResultMsg{enabled: false, err: "Failed to connect to PostgreSQL: " + err.Error()}
			}
			// Initialize PostgreSQL schema
			if err := db.InitializePostgresDatabase(); err != nil {
				return syncInitResultMsg{enabled: false, err: "Failed to initialize PostgreSQL: " + err.Error()}
			}
			postgresDB = db.GetPostgresDB()
		}

		if postgresDB == nil {
			return syncInitResultMsg{enabled: false, err: "PostgreSQL database not connected"}
		}

		// Create the sync service
		svc := sync.NewSyncService(sqliteDB, postgresDB, syncInterval)
		return syncInitResultMsg{enabled: true, service: svc}
	}
}

// syncInitResultMsg contains the result of sync service initialization
type syncInitResultMsg struct {
	enabled bool
	service *sync.SyncService
	err     string
}

// SyncTickCmd returns a command that triggers periodic sync
func SyncTickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return SyncTickMsg{}
	})
}

// DoSyncCmd performs the actual sync operation in the background
func DoSyncCmd(svc *sync.SyncService) tea.Cmd {
	return func() tea.Msg {
		if svc == nil {
			return SyncCompleteMsg{Err: nil}
		}

		err := svc.Sync(sync.SyncBidirectional)
		return SyncCompleteMsg{
			Stats: svc.GetLastSyncStats(),
			Err:   err,
		}
	}
}

// FormatSyncStatus returns a human-readable sync status string
func FormatSyncStatus(lastSync time.Time, isSyncing bool, hasError bool) string {
	if isSyncing {
		return "Syncing..."
	}

	if hasError {
		return "Sync error"
	}

	if lastSync.IsZero() {
		return "Not synced"
	}

	// Calculate time since last sync
	elapsed := time.Since(lastSync)

	if elapsed < time.Minute {
		return "Synced just now"
	} else if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		if minutes == 1 {
			return "Synced 1m ago"
		}
		return fmt.Sprintf("Synced %dm ago", minutes)
	} else {
		hours := int(elapsed.Hours())
		if hours == 1 {
			return "Synced 1h ago"
		}
		return fmt.Sprintf("Synced %dh ago", hours)
	}
}
