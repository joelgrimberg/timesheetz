package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/logging"
	"timesheet/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

// Command line flags
type flags struct {
	noTUI   bool
	init    bool
	help    bool
	verbose bool
}

// setupFlags defines and parses command line flags
func setupFlags() flags {
	// Define flags
	noTUI := flag.Bool("no-tui", false, "Run only the API server without the TUI")
	initFlag := flag.Bool("init", false, "Initialize the database")
	helpFlag := flag.Bool("help", false, "Show help message")
	verboseFlag := flag.Bool("verbose", false, "Show detailed output")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --init          Initialize the database\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --no-tui        Run only the API server\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --help          Show this help message\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose       Show detailed output\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	return flags{
		noTUI:   *noTUI,
		init:    *initFlag,
		help:    *helpFlag,
		verbose: *verboseFlag,
	}
}

func main() {
	// Setup and parse flags
	flags := setupFlags()

	// Show help and exit if --help is used
	if flags.help {
		flag.Usage()
		os.Exit(0)
	}

	// Clear the screen
	fmt.Print("\033[H\033[2J")

	// Set up logging
	logFile := logging.SetupLogging()
	defer logFile.Close()

	// Set verbose mode
	logging.SetVerbose(flags.verbose)

	// Initialize the database
	dbPath := db.GetDBPath()
	if err := db.Connect(dbPath); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Handle database initialization if requested
	if flags.init {
		if err := db.InitializeDatabase(dbPath); err != nil {
			log.Fatalf("Error initializing database: %v", err)
		}
		log.Println("Database initialized successfully")
		// If just initializing, exit after success
		if len(flag.Args()) == 0 {
			os.Exit(0)
		}
	}

	// Handle no-tui mode
	if flags.noTUI {
		log.Println("Starting API server in --no-tui mode...")
		apiP := tea.NewProgram(ui.NewAppModel())
		handler.StartServer(apiP)
		// Keep the application running in the background
		select {}
	}

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
