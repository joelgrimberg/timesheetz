package db

import (
	"fmt"
	"strings"

	"timesheet/internal/logging"
)

// InitializePostgresDatabase creates the database tables if they don't exist
func InitializePostgresDatabase() error {
	if pgDB == nil {
		return fmt.Errorf("postgres connection not established")
	}

	stmts := []string{
		// Clients table (must be created before timesheet due to foreign key)
		`CREATE TABLE IF NOT EXISTS clients (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_active INTEGER DEFAULT 1
		)`,
		`CREATE INDEX IF NOT EXISTS idx_clients_name ON clients(name)`,
		`CREATE INDEX IF NOT EXISTS idx_clients_active ON clients(is_active)`,

		// Timesheet table
		`CREATE TABLE IF NOT EXISTS timesheet (
			id SERIAL PRIMARY KEY,
			date TEXT NOT NULL,
			client_name TEXT NOT NULL,
			client_hours INTEGER DEFAULT NULL,
			vacation_hours INTEGER DEFAULT NULL,
			idle_hours INTEGER DEFAULT NULL,
			training_hours INTEGER DEFAULT NULL,
			sick_hours INTEGER DEFAULT NULL,
			holiday_hours INTEGER DEFAULT NULL,
			client_id INTEGER REFERENCES clients(id),
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_client_name ON timesheet(client_name)`,
		`CREATE INDEX IF NOT EXISTS idx_timesheet_date ON timesheet(date)`,
		`CREATE INDEX IF NOT EXISTS idx_timesheet_date_client ON timesheet(date, client_name)`,

		// Training budget table
		`CREATE TABLE IF NOT EXISTS training_budget (
			id SERIAL PRIMARY KEY,
			date TEXT NOT NULL,
			training_name TEXT NOT NULL,
			hours INTEGER NOT NULL,
			cost_without_vat DECIMAL(10,2) NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_training_date ON training_budget(date)`,

		// Client rates table
		`CREATE TABLE IF NOT EXISTS client_rates (
			id SERIAL PRIMARY KEY,
			client_id INTEGER NOT NULL,
			hourly_rate DECIMAL(10,2) NOT NULL,
			effective_date TEXT NOT NULL,
			notes TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_client_rates_client ON client_rates(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_client_rates_date ON client_rates(effective_date)`,
		`CREATE INDEX IF NOT EXISTS idx_client_rates_client_date ON client_rates(client_id, effective_date)`,

		// Vacation carryover table
		`CREATE TABLE IF NOT EXISTS vacation_carryover (
			id SERIAL PRIMARY KEY,
			year INTEGER NOT NULL UNIQUE,
			carryover_hours INTEGER NOT NULL,
			source_year INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			notes TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_vacation_carryover_year ON vacation_carryover(year)`,
	}

	for _, stmt := range stmts {
		if _, err := pgDB.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nSQL: %s", err, stmt)
		}
	}

	// Migration: Add updated_at columns for sync support (for existing tables)
	migrations := []struct {
		table  string
		column string
	}{
		{"timesheet", "created_at"},
		{"timesheet", "updated_at"},
		{"training_budget", "created_at"},
		{"training_budget", "updated_at"},
		{"clients", "updated_at"},
		{"client_rates", "updated_at"},
	}

	for _, m := range migrations {
		sql := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s TEXT DEFAULT CURRENT_TIMESTAMP`, m.table, m.column)
		_, err := pgDB.Exec(sql)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			logging.Log("Note: Could not add %s.%s column: %v", m.table, m.column, err)
		}
	}

	// Set default values for existing rows that have NULL timestamps
	pgDB.Exec(`UPDATE timesheet SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL`)
	pgDB.Exec(`UPDATE timesheet SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL`)
	pgDB.Exec(`UPDATE training_budget SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL`)
	pgDB.Exec(`UPDATE training_budget SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL`)
	pgDB.Exec(`UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL`)
	pgDB.Exec(`UPDATE client_rates SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL`)

	logging.Log("PostgreSQL database initialized successfully")
	return nil
}
