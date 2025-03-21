package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/db"
	"timesheet/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

func initDatabase() {
	dbPath := getDBPath()

	initFlag := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	// Check if initialization is requested
	if *initFlag {
		if err := db.InitializeDatabase(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database initialized successfully at:", dbPath)
		// If just initializing, exit after success
		if len(flag.Args()) == 0 {
			os.Exit(0)
		}
	}

	// Connect to the database
	err := db.Connect(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
}

// getDBPath returns the path to the SQLite database file
func getDBPath() string {
	// Default path in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If there's an error getting home dir, use current directory
		return "timesheet.db"
	}

	// Create the .timesheet directory if it doesn't exist
	dbDir := filepath.Join(homeDir, ".config/timesheet")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		// If directory creation fails, use current directory
		return "timesheet.db"
	}

	return filepath.Join(dbDir, "timesheet.db")
}

func main() {
	// Initialize the database first
	initDatabase()
	defer db.Close()

	// Initialize the app with timesheet as the default view
	app := ui.NewAppModel()

	// Start API server in a goroutine before running the UI
	go func() {
		fmt.Println("Starting API server...")
		apiP := tea.NewProgram(ui.NewAppModel())
		handler.StartServer(apiP)
	}()

	// Give the API server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Run the UI program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
