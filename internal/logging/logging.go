package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	verbose bool
	logFile *os.File
)

// SetVerbose sets the verbose mode
func SetVerbose(v bool) {
	verbose = v
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// Log prints a message if verbose mode is enabled
func Log(format string, v ...interface{}) {
	if verbose {
		// Print to console
		fmt.Printf(format+"\n", v...)
		// Also log to file (only if logFile is not stderr)
		if logFile != nil && logFile != os.Stderr {
			log.Printf(format, v...)
		}
	}
}

// SetupLogging initializes logging and returns the log file.
func SetupLogging() *os.File {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Warning: couldn't get home directory, using current directory for logs")
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, ".local", "state", "timesheetz", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Println("Warning: couldn't create logs directory:", err)
		logDir = "."
	}

	dailyTimestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("timesheet_%s.log", dailyTimestamp))

	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		// In non-interactive environments (like Docker), fall back to stderr
		// Don't use log.Fatalf as it would cause the program to exit
		log.Printf("Warning: couldn't open log file %s: %v, using stderr", logPath, err)
		return os.Stderr
	}

	// Create a multi-writer to write to both file and console
	log.SetOutput(f)
	log.Printf("Logging initialized at %s", time.Now().Format("15:04:05"))

	return f
}
