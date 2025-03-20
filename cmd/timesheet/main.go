package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/db"
	"timesheet/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/go-sql-driver/mysql"
)

func initDatabase() {
	dbUser, dbPassword := GetDBCredentials()

	initFlag := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	// Check if initialization is requested
	if *initFlag {
		if err := db.InitializeDatabase(dbUser, dbPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		// If just initializing, exit after success
		if len(flag.Args()) == 0 {
			os.Exit(0)
		}
	}

	if dbUser == "" || dbPassword == "" {
		fmt.Println("Error: Database username or password is empty")
		os.Exit(1)
	}

	err := db.Connect(dbUser, dbPassword)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
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
