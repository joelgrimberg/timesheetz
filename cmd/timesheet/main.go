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
	dev     bool
	port    int
}

// setupFlags defines and parses command line flags
func setupFlags() flags {
	// Define flags
	noTUI := flag.Bool("no-tui", false, "Run only the API server without the TUI")
	initFlag := flag.Bool("init", false, "Initialize the database")
	helpFlag := flag.Bool("help", false, "Show help message")
	verboseFlag := flag.Bool("verbose", false, "Show detailed output")
	devFlag := flag.Bool("dev", false, "Run in development mode (uses local database)")
	portFlag := flag.Int("port", 0, "Specify the port for the API server")

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
		fmt.Fprintf(os.Stderr, "  %s --dev           Run in development mode\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 3000     Run API server on port 3000\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	return flags{
		noTUI:   *noTUI,
		init:    *initFlag,
		help:    *helpFlag,
		verbose: *verboseFlag,
		dev:     *devFlag,
		port:    *portFlag,
	}
}

func main() {
	// Setup and parse flags
	flags := setupFlags()
	log.Println("Flags parsed successfully")

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
	log.Println("Logging setup complete")

	// Set verbose mode
	logging.SetVerbose(flags.verbose)
	log.Println("Verbose mode set to:", flags.verbose)

	// Read configuration file (and create if it doesn't exist)
	config.RequireConfig()
	log.Println("Config file checked/created")

	// If dev flag is set, set runtime development mode
	if flags.dev {
		log.Println("Development mode flag detected")
		config.SetRuntimeDevMode(true)
	}

	// If port flag is set, set runtime port
	if flags.port != 0 {
		log.Println("Port flag detected:", flags.port)
		config.SetRuntimePort(flags.port)
	}

	// Initialize the database
	dbPath := db.GetDBPath()
	log.Printf("Database path: %s", dbPath)

	// Check if database exists, if not initialize it
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Println("Database does not exist, initializing...")
		if err := db.InitializeDatabase(dbPath); err != nil {
			log.Fatalf("Error initializing database: %v", err)
		}
		log.Println("Database initialized successfully")
	} else {
		log.Println("Database file exists")
	}

	log.Println("Attempting to connect to database...")
	if err := db.Connect(dbPath); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Handle database initialization if requested
	if flags.init {
		log.Println("Init flag detected, reinitializing database...")
		if err := db.InitializeDatabase(dbPath); err != nil {
			log.Fatalf("Error initializing database: %v", err)
		}
		log.Println("Database reinitialized successfully")
		// If just initializing, exit after success
		if len(flag.Args()) == 0 {
			os.Exit(0)
		}
	}

	// Initialize the app with timesheet as the default view
	log.Println("Initializing UI...")
	app := ui.NewAppModel()
	refreshChan := app.GetRefreshChan()
	log.Println("UI initialized")

	// Create the UI program first
	p := tea.NewProgram(app)
	log.Println("UI program created")

	// Handle no-tui mode
	if flags.noTUI {
		log.Println("Starting API server in --no-tui mode...")
		// Start the API server
		handler.StartServer(p, refreshChan)
		// The server will keep running until interrupted
		// No need for select{} as StartServer blocks
	}

	if config.GetStartAPIServer() {
		// Start API server in a goroutine before running the UI
		go func() {
			log.Println("Starting API server...")
			handler.StartServer(p, refreshChan)
		}()

		// Give the API server a moment to start
		time.Sleep(100 * time.Millisecond)
		log.Println("API server started")
	}

	// Start a goroutine to handle refresh messages
	go func() {
		log.Println("Starting refresh message handler...")
		for {
			select {
			case <-refreshChan:
				// Send refresh message to the UI program
				p.Send(ui.RefreshMsg{})
			}
		}
	}()

	// Run the UI program
	log.Println("Starting UI program...")
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
