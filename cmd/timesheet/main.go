package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

func setupLogging() *os.File {
	// Create logs directory inside the user's .config/timesheet folder
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Warning: couldn't get home directory, using current directory for logs")
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, ".config/timesheet/logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Println("Warning: couldn't create logs directory:", err)
		logDir = "."
	}

	// Use just the date part for daily log files
	dailyTimestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("timesheet_%s.log", dailyTimestamp))

	// Open file with append flag - this will create it if it doesn't exist
	// or append to it if it does
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Don't close the file here - we want it to stay open for the duration of the program
	// Instead, we'll defer the close in main()

	log.SetOutput(f)
	log.Printf("Logging initialized at %s", time.Now().Format("15:04:05"))

	return f
}

func initDatabase() {
	// Add a --no-tui flag
	noTUI := flag.Bool("no-tui", false, "Run only the API server without the TUI")

	initFlag := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	// Connect to the database
	dbPath := getDBPath()
	err := db.Connect(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Check if --no-tui is set and handle it early
	if *noTUI {
		log.Println("Starting API server in --no-tui mode...")
		apiP := tea.NewProgram(ui.NewAppModel())
		handler.StartServer(apiP)

		// Keep the application running in the background
		select {}
	}

	// Check if initialization is requested
	if *initFlag {
		if err := db.InitializeDatabase(dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		log.Println("Database initialized successfully at:", dbPath)
		// If just initializing, exit after success
		if len(flag.Args()) == 0 {
			os.Exit(0)
		}
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
	// Set up logging first thing
	logFile := setupLogging()
	defer logFile.Close()

	// Initialize the database first
	initDatabase()
	defer db.Close()

	// Read configuration file (and create if it doesn't exist)
	config.RequireConfig()

	// Initialize the app with timesheet as the default view
	app := ui.NewAppModel()
	if config.GetStartAPIServer() {
		// Start API server in a goroutine before running the UI
		go func() {
			log.Println("Starting API server...")
			apiP := tea.NewProgram(ui.NewAppModel())
			handler.StartServer(apiP)
		}()

		// Give the API server a moment to start
		time.Sleep(100 * time.Millisecond)
	}

	// Run the UI program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
