package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"timesheet/api/middleware"
	"timesheet/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gin-gonic/gin"
)

type ApiMsg struct {
	IP string
}

func StartServer(p *tea.Program) {
	router := gin.Default()

	// disable trusted proxies functionality
	router.SetTrustedProxies(nil)

	// Add the request ID middleware early in the chain
	router.Use(middleware.RequestIDMiddleware())

	// Apply security headers middleware
	router.Use(middleware.SecurityHeaders())

	// Middleware to extract and convert IP address to IPv4 if necessary
	router.Use(middleware.RetreiveIP())

	router.GET("/health", func(context *gin.Context) {
		// Check DB connection
		if err := db.Ping(); err != nil {
			context.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "reason": "database unavailable"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.POST("/api/entry", func(context *gin.Context) {
		requestID, _ := context.Get("RequestID")
		log.Printf("[%s] Processing new entry request", requestID)

		// Define a struct to bind the JSON request body
		type EntryRequest struct {
			Date          string `json:"date"`
			ClientName    string `json:"client_name"`
			ClientHours   int    `json:"client_hours"`
			VacationHours int    `json:"vacation_hours"`
			IdleHours     int    `json:"idle_hours"`
			TrainingHours int    `json:"training_hours"`
			HolidayHours  int    `json:"holiday_hours"`
			SickHours     int    `json:"sick_hours"`
		}

		var req EntryRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Use current date if not provided
		if req.Date == "" {
			req.Date = time.Now().Format("2006-01-02")
		}

		// Validate date format
		_, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}

		// Add validation
		if req.ClientHours < 0 || req.VacationHours < 0 || req.IdleHours < 0 || req.TrainingHours < 0 {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Hours cannot be negative"})
			return
		}

		// Create entry struct
		entry := db.TimesheetEntry{
			Date:           req.Date,
			Client_name:    req.ClientName,
			Client_hours:   req.ClientHours,
			Vacation_hours: req.VacationHours,
			Idle_hours:     req.IdleHours,
			Training_hours: req.TrainingHours,
			Holiday_hours:  req.HolidayHours,
			Sick_hours:     req.SickHours,
		}

		// Call the AddTimesheetEntry function
		err = db.AddTimesheetEntry(entry)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create timesheet entry: " + err.Error()})
			return
		}

		// Return success response
		context.JSON(http.StatusCreated, gin.H{"message": "🎉 Timesheet entry created successfully"})
	})

	router.DELETE("/api/entry/:id", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")
		log.Printf("[%s] Processing delete entry request", requestID)

		// Get the ID from the URL parameter
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Entry ID is required"})
			return
		}

		// Call the DeleteTimesheetEntry function
		err := db.DeleteTimesheetEntry(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete timesheet entry: " + err.Error()})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Timesheet entry deleted successfully"})
	})

	router.GET("/api", func(c *gin.Context) {
		IP, exists := c.Get("clientIP")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve IP address"})
			return
		}

		// Send a message to the bubbletea program to update the view
		teaMsg := ApiMsg{IP: IP.(string)}
		p.Send(teaMsg)
		c.JSON(http.StatusOK, gin.H{"response": gin.H{"ip": IP}})
	})

	router.GET("/api/entries", func(c *gin.Context) {
		// Extract year and month from query parameters if provided
		yearStr := c.Query("year")
		monthStr := c.Query("month")

		var year int = 0
		var month time.Month = 0
		var err error

		if yearStr != "" {
			year, err = strconv.Atoi(yearStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year format"})
				return
			}
		}

		if monthStr != "" {
			monthInt, err := strconv.Atoi(monthStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
				return
			}
			month = time.Month(monthInt)
		}

		entries, err := db.GetAllTimesheetEntries(year, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timesheet entries: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"entries": entries})
	})

	router.GET("/api/entry/:date", func(c *gin.Context) {
		date := c.Param("date")
		if date == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date is required"})
			return
		}

		entry, err := db.GetTimesheetEntryByDate(date)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found for date: " + date})
			return
		}

		c.JSON(http.StatusOK, entry)
	})

	router.PUT("/api/entry/:id", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")
		log.Printf("[%s] Processing update entry request", requestID)

		// Get the ID from the URL parameter
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Entry ID is required"})
			return
		}

		// Define a struct to bind the JSON request body
		type EntryUpdateRequest struct {
			ClientHours   *int    `json:"client_hours"`
			VacationHours *int    `json:"vacation_hours"`
			IdleHours     *int    `json:"idle_hours"`
			TrainingHours *int    `json:"training_hours"`
			ClientName    *string `json:"client_name,omitempty"`
			HolidayHours  *int    `json:"holiday_hours,omitempty"`
			SickHours     *int    `json:"sick_hours,omitempty"`
		}

		var req EntryUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Convert the request to a map of fields to update
		updates := make(map[string]any)

		// Add client_name if provided
		if req.ClientName != nil {
			updates["client_name"] = *req.ClientName
		}

		if req.ClientHours != nil {
			if *req.ClientHours < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Client hours cannot be negative"})
				return
			}
			updates["client_hours"] = *req.ClientHours
		}
		if req.VacationHours != nil {
			if *req.VacationHours < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Vacation hours cannot be negative"})
				return
			}
			updates["vacation_hours"] = *req.VacationHours
		}
		if req.IdleHours != nil {
			if *req.IdleHours < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Idle hours cannot be negative"})
				return
			}
			updates["idle_hours"] = *req.IdleHours
		}
		if req.TrainingHours != nil {
			if *req.TrainingHours < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Training hours cannot be negative"})
				return
			}
			updates["training_hours"] = *req.TrainingHours
		}

		// Check if there's anything to update
		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		// Call the UpdateTimesheetEntryById function
		err := db.UpdateTimesheetEntryById(id, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update timesheet entry: " + err.Error()})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Timesheet entry updated successfully"})
	})

	router.PUT("/api/entry/date/:date", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")
		log.Printf("[%s] Processing update entry by date request", requestID)

		// Get the date from the URL parameter
		date := c.Param("date")
		if date == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date is required"})
			return
		}

		// Define a struct to bind the JSON request body
		type EntryFullUpdateRequest struct {
			ClientName    string `json:"client_name"`
			ClientHours   int    `json:"client_hours"`
			VacationHours int    `json:"vacation_hours"`
			IdleHours     int    `json:"idle_hours"`
			TrainingHours int    `json:"training_hours"`
			HolidayHours  int    `json:"holiday_hours"`
			SickHours     int    `json:"sick_hours"`
		}

		var req EntryFullUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Validate hours
		if req.ClientHours < 0 || req.VacationHours < 0 || req.IdleHours < 0 || req.TrainingHours < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hours cannot be negative"})
			return
		}

		// Create the entry object
		entry := db.TimesheetEntry{
			Date:           date,
			Client_name:    req.ClientName,
			Client_hours:   req.ClientHours,
			Vacation_hours: req.VacationHours,
			Idle_hours:     req.IdleHours,
			Training_hours: req.TrainingHours,
			Holiday_hours:  req.HolidayHours,
			Sick_hours:     req.SickHours,
		}

		// Call the UpdateTimesheetEntry function
		err := db.UpdateTimesheetEntry(entry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update timesheet entry: " + err.Error()})
			return
		}

		// Return success response
		c.JSON(http.StatusOK, gin.H{"message": "Timesheet entry updated successfully"})
	})

	log.Println("Starting server on 0.0.0.0:8080")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
