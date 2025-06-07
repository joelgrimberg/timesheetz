package handler

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"timesheet/api/middleware"
	"timesheet/internal/config"
	"timesheet/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gin-gonic/gin"
)

type ApiMsg struct {
	IP string
}

// IsAPIRunning checks if the API is running on the specified port
func IsAPIRunning(port int) bool {
	// Try to connect to the health endpoint
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// StartServer starts the API server
func StartServer(p *tea.Program, refreshChan chan ui.RefreshMsg) {
	// Get the configured port
	port := config.GetAPIPort()
	addr := fmt.Sprintf("0.0.0.0:%d", port)

	// Check if port is available
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("\nError: Port %d is already in use. This might be because:\n", port)
		fmt.Printf("1. Another instance of Timesheetz is already running\n")
		fmt.Printf("2. Another application is using this port\n\n")
		fmt.Printf("To fix this, you can:\n")
		fmt.Printf("1. Stop the other instance\n")
		fmt.Printf("2. Run Timesheetz with a different port: --port <number>\n")
		fmt.Printf("   Example: ./bin/timesheet --port 8081\n\n")
		log.Fatalf("Port %d is not available: %v", port, err)
	}
	listener.Close()

	// Set Gin to Release Mode
	gin.SetMode(gin.ReleaseMode)

	// Create a custom logger that writes to a file instead of stdout
	logFile, err := os.OpenFile("gin.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open gin log file: %v", err)
	}
	defer logFile.Close()

	// Create a custom logger for Gin
	ginLogger := gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    logFile,
		SkipPaths: []string{"/health"}, // Skip logging for health checks
	})

	router := gin.New() // Use New() instead of Default() to avoid default middleware
	router.Use(ginLogger)
	router.Use(gin.Recovery())

	// disable trusted proxies functionality
	router.SetTrustedProxies(nil)

	// Add the request ID middleware early in the chain
	router.Use(middleware.RequestIDMiddleware())

	// Apply security headers middleware
	router.Use(middleware.SecurityHeaders())

	// Middleware to extract and convert IP address to IPv4 if necessary
	router.Use(middleware.RetreiveIP())

	// Helper function to send refresh message
	sendRefresh := func() {
		select {
		case refreshChan <- ui.RefreshMsg{}:
		default:
			// Channel is full or closed, ignore
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Timesheet routes
		api.GET("/timesheet", func(c *gin.Context) {
			GetTimesheet(c)
		})
		api.POST("/timesheet", func(c *gin.Context) {
			CreateTimesheet(c)
			sendRefresh()
		})
		api.PUT("/timesheet/:id", func(c *gin.Context) {
			UpdateTimesheet(c)
			sendRefresh()
		})
		api.DELETE("/timesheet/:id", func(c *gin.Context) {
			DeleteTimesheet(c)
			sendRefresh()
		})

		// Get last client name
		api.GET("/last-client", GetLastClientName)

		// Export routes
		api.GET("/export/pdf", ExportPDF)
		api.GET("/export/excel", ExportExcel)
	}

	// Start the server
	fmt.Printf("\nTimesheet API started on http://localhost:%d\n\n", port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
