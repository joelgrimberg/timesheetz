package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/logging"
	"timesheet/internal/sync"
	"timesheet/internal/ui"
	"timesheet/internal/version"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

// Command line flags
type flags struct {
	noTUI       bool
	tuiOnly     bool
	add         bool
	init        bool
	help        bool
	verbose     bool
	dev         bool
	port        int
	dbType      string
	postgresURL string
	syncCmd     bool
}

// setupFlags defines and parses command line flags
func setupFlags() flags {
	// Define flags
	noTUI := flag.Bool("no-tui", false, "Run only the API server without the TUI")
	tuiOnly := flag.Bool("tui-only", false, "Run only the TUI without the API server")
	addFlag := flag.Bool("add", false, "Add a new entry for today and exit")
	initFlag := flag.Bool("init", false, "Initialize the database")
	helpFlag := flag.Bool("help", false, "Show help message")
	verboseFlag := flag.Bool("verbose", false, "Show detailed output")
	devFlag := flag.Bool("dev", false, "Run in development mode (uses local database)")
	portFlag := flag.Int("port", 0, "Specify the port for the API server")
	dbTypeFlag := flag.String("db-type", "", "Database type: sqlite or postgres")
	postgresURLFlag := flag.String("postgres-url", "", "PostgreSQL connection URL")
	versionFlag := flag.Bool("version", false, "Show version and exit")
	syncFlag := flag.Bool("sync", false, "Sync SQLite and PostgreSQL databases (requires both to be configured)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --init          Initialize the database\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --no-tui        Run only the API server\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --tui-only      Run only the TUI without the API server\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --add           Add a new entry for today and exit\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --help          Show this help message\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose       Show detailed output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dev           Run in development mode\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 3000     Run API server on port 3000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --db-type postgres --postgres-url \"postgres://...\"  Use PostgreSQL\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --sync --postgres-url \"postgres://...\"  Sync SQLite to PostgreSQL\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	// Check for version flag
	if *versionFlag {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	return flags{
		noTUI:       *noTUI,
		tuiOnly:     *tuiOnly,
		add:         *addFlag,
		init:        *initFlag,
		help:        *helpFlag,
		verbose:     *verboseFlag,
		dev:         *devFlag,
		port:        *portFlag,
		dbType:      *dbTypeFlag,
		postgresURL: *postgresURLFlag,
		syncCmd:     *syncFlag,
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

	// Clear the screen (only if we have a terminal)
	if !flags.noTUI {
		fmt.Print("\033[H\033[2J")
	}

	// Set up logging
	logFile := logging.SetupLogging()
	if logFile != nil {
		defer logFile.Close()
	}
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

	// Add panic recovery at the top level
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "PANIC RECOVERED: %v\n", r)
			log.Printf("Panic recovered: %v", r)
			os.Exit(1)
		}
	}()

	// If port flag is set, set runtime port
	if flags.port != 0 {
		log.Println("Port flag detected:", flags.port)
		config.SetRuntimePort(flags.port)
	}

	// Handle database type selection
	if flags.dbType != "" {
		log.Println("Database type flag detected:", flags.dbType)
		config.SetRuntimeDBType(flags.dbType)
	}
	if flags.postgresURL != "" {
		log.Println("PostgreSQL URL flag detected")
		config.SetRuntimePostgresURL(flags.postgresURL)
	}

	// Get the database type to use
	dbType := config.GetDBType()
	log.Printf("Using database type: %s", dbType)

	// Initialize database based on type
	if dbType == "postgres" {
		// PostgreSQL mode
		postgresURL := config.GetPostgresURL()
		if postgresURL == "" {
			log.Fatal("PostgreSQL URL required when using postgres db type. Set via --postgres-url, TIMESHEETZ_POSTGRES_URL, or config file.")
		}

		log.Println("Attempting to connect to PostgreSQL...")
		if err := db.ConnectPostgres(postgresURL); err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer db.ClosePostgres()
		log.Println("PostgreSQL connected successfully")

		// Initialize schema
		log.Println("Initializing PostgreSQL schema...")
		if err := db.InitializePostgresDatabase(); err != nil {
			log.Fatalf("Error initializing PostgreSQL database: %v", err)
		}
		log.Println("PostgreSQL schema initialized successfully")

		// Handle --init flag for postgres
		if flags.init {
			log.Println("PostgreSQL database reinitialized")
			if len(flag.Args()) == 0 {
				os.Exit(0)
			}
		}
	} else {
		// SQLite mode (default)
		dbPath := config.GetDBPath()
		log.Printf("Database path: %s", dbPath)

		// Check if database exists, if not initialize it
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			log.Println("Database does not exist, initializing...")
			if err := db.InitializeDatabase(dbPath); err != nil {
				log.Fatalf("Error initializing database: %v", err)
			}
			log.Println("Database initialized successfully")
		} else if err != nil {
			log.Fatalf("Error checking database: %v", err)
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
	}

	// Handle --sync command: sync between SQLite and PostgreSQL
	// This needs special handling because we need BOTH databases
	if flags.syncCmd {
		log.Println("Sync command detected")

		// For sync, we need both databases connected
		postgresURL := config.GetPostgresURL()
		if postgresURL == "" {
			log.Fatal("PostgreSQL URL required for sync. Set via --postgres-url, TIMESHEETZ_POSTGRES_URL, or config file.")
		}

		// Always connect to SQLite for sync (regardless of db-type setting)
		dbPath := db.GetDBPath()
		log.Printf("Connecting to SQLite for sync at: %s", dbPath)
		if err := db.Connect(dbPath); err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
		defer db.Close()

		// Run SQLite migrations to ensure sync columns exist
		if err := db.InitializeDatabase(dbPath); err != nil {
			log.Fatalf("Failed to initialize SQLite: %v", err)
		}

		// Always connect to PostgreSQL for sync
		log.Println("Connecting to PostgreSQL for sync...")
		if err := db.ConnectPostgres(postgresURL); err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer db.ClosePostgres()

		// Initialize PostgreSQL schema
		if err := db.InitializePostgresDatabase(); err != nil {
			log.Fatalf("Error initializing PostgreSQL database: %v", err)
		}

		// Create sync service and run sync
		fmt.Println("Starting database sync...")
		syncService := sync.NewSyncService(db.GetSQLiteDB(), db.GetPostgresDB(), time.Minute)

		if err := syncService.Sync(sync.SyncBidirectional); err != nil {
			log.Fatalf("Sync failed: %v", err)
		}

		stats := syncService.GetLastSyncStats()
		fmt.Printf("Sync completed in %v\n", stats.Duration)
		fmt.Printf("  Records pushed (local -> remote): %d\n", stats.RecordsPushed)
		fmt.Printf("  Records pulled (remote -> local): %d\n", stats.RecordsPulled)
		fmt.Printf("  Tables processed: %d\n", stats.TablesProcessed)
		if len(stats.Errors) > 0 {
			fmt.Printf("  Errors: %d\n", len(stats.Errors))
			for _, e := range stats.Errors {
				fmt.Printf("    - %s\n", e)
			}
		}
		os.Exit(0)
	}

	// Start the TUI if requested
	if flags.tuiOnly {
		log.Println("Starting TUI only mode...")
		model := ui.NewAppModel(flags.add)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
		os.Exit(0)
	}

	// If --no-tui flag is set, start only the API server
	if flags.noTUI {
		log.Println("Starting API server only mode...")
		refreshChan := make(chan ui.RefreshMsg)
		handler.StartServer(nil, refreshChan)
		// Keep the server running
		select {}
	}

	// Initialize the app with timesheet as the default view
	log.Println("Initializing UI...")
	app := ui.NewAppModel(flags.add)
	refreshChan := app.GetRefreshChan()
	log.Println("UI initialized")

	// Create the UI program first
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	log.Println("UI program created")

	// Start API server if not in tui-only mode or add mode
	if !flags.tuiOnly && !flags.add && config.GetStartAPIServer() {
		// Check if API is already running on the configured port
		port := config.GetAPIPort()
		apiRunning := false

		// Try to connect to the API to check if it's already running
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
		if err == nil {
			resp.Body.Close()
			apiRunning = true
			log.Printf("API server already running on port %d, skipping startup", port)
		}

		// Only start API server if it's not already running
		if !apiRunning {
			// Start API server in a goroutine before running the UI
			go func() {
				log.Println("Starting API server...")
				handler.StartServer(p, refreshChan)
			}()

			// Give the API server a moment to start
			time.Sleep(100 * time.Millisecond)
			log.Println("API server started")
		}
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

	// If --add flag is set, start in form mode for today
	if flags.add {
		// Switch to form mode
		app.ActiveMode = ui.FormMode
		// Initialize form for today
		app.FormModel = ui.InitialFormModel()
	}

	// Run the UI program
	log.Println("Starting UI program...")
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		os.Exit(1)
	}

	// Clean up the terminal
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[2J")   // Clear screen
	fmt.Print("\033[H")    // Move cursor to top-left
}
