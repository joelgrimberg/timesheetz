package handler

import (
	"log"
	"net/http"
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

	router.DELETE("/api/entry/:id", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")

		// Use it in logs
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

	router.POST("/api/entry", func(context *gin.Context) {
		requestID, _ := context.Get("RequestID")

		// Use it in logs
		log.Printf("[%s] Processing entry request", requestID)

		// Define a struct to bind the JSON request body
		type EntryRequest struct {
			ClientHours   float64 `json:"client_hours"`
			VacationHours float64 `json:"vacation_hours"`
			IdleHours     float64 `json:"idle_hours"`
			TrainingHours float64 `json:"training_hours"`
		}

		var req EntryRequest
		if err := context.ShouldBindJSON(&req); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Add validation
		if req.ClientHours < 0 || req.VacationHours < 0 || req.IdleHours < 0 || req.TrainingHours < 0 {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Hours cannot be negative"})
			return
		}

		// Call the PutTimesheetEntry function
		id, err := db.PutTimesheetEntry(req.ClientHours, req.VacationHours, req.IdleHours, req.TrainingHours)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create timesheet entry: " + err.Error()})
			return
		}

		// Return success response with the new entry ID
		context.JSON(http.StatusCreated, gin.H{"id": id, "message": "Timesheet entry created successfully"})
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

	router.GET("/api/all", func(c *gin.Context) {
		entries, err := db.GetAllTimesheetEntries()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timesheet entries"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"entries": entries})
	})

	router.PUT("/api/entry/:id", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")

		// Use it in logs
		log.Printf("[%s] Processing update entry request", requestID)

		// Get the ID from the URL parameter
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Entry ID is required"})
			return
		}

		// Define a struct to bind the JSON request body
		type EntryUpdateRequest struct {
			ClientHours   *float64 `json:"client_hours"`
			VacationHours *float64 `json:"vacation_hours"`
			IdleHours     *float64 `json:"idle_hours"`
			TrainingHours *float64 `json:"training_hours"`
		}

		var req EntryUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Convert the request to a map of fields to update
		updates := make(map[string]any)
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

		// Call the UpdateTimesheetEntry function
		err := db.UpdateTimesheetEntry(id, updates)
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
