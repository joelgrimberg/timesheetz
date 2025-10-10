package handler

import (
	"net/http"
	"strconv"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"

	"github.com/gin-gonic/gin"
)

// GetTimesheet handles GET requests for timesheet entries
func GetTimesheet(c *gin.Context) {
	entries, err := db.GetAllTimesheetEntries(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// CreateTimesheet handles POST requests to create a new timesheet entry
func CreateTimesheet(c *gin.Context) {
	var entry db.TimesheetEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.AddTimesheetEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// UpdateTimesheet handles PUT requests to update a timesheet entry
func UpdateTimesheet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var entry db.TimesheetEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.UpdateTimesheetEntryById(id, map[string]any{
		"client_name":    entry.Client_name,
		"client_hours":   entry.Client_hours,
		"vacation_hours": entry.Vacation_hours,
		"idle_hours":     entry.Idle_hours,
		"training_hours": entry.Training_hours,
		"holiday_hours":  entry.Holiday_hours,
		"sick_hours":     entry.Sick_hours,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// DeleteTimesheet handles DELETE requests to remove a timesheet entry
func DeleteTimesheet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	if err := db.DeleteTimesheetEntry(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted successfully"})
}

// ExportPDF handles GET requests to export timesheet as PDF
func ExportPDF(c *gin.Context) {
	// TODO: Implement PDF export
	c.JSON(http.StatusNotImplemented, gin.H{"error": "PDF export not implemented yet"})
}

// ExportExcel handles GET requests to export timesheet as Excel
func ExportExcel(c *gin.Context) {
	// TODO: Implement Excel export
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Excel export not implemented yet"})
}

// GetLastClientName handles GET requests for the last client name
func GetLastClientName(c *gin.Context) {
	clientName, err := db.GetLastClientName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_name": clientName})
}

// GetTrainingBudget handles GET requests for training budget entries
func GetTrainingBudget(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Year parameter is required"})
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	entries, err := db.GetTrainingBudgetEntriesForYear(yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// CreateTrainingBudget handles POST requests to create a new training budget entry
func CreateTrainingBudget(c *gin.Context) {
	var entry db.TrainingBudgetEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.AddTrainingBudgetEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// UpdateTrainingBudget handles PUT requests to update a training budget entry
func UpdateTrainingBudget(c *gin.Context) {
	var entry db.TrainingBudgetEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.UpdateTrainingBudgetEntry(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// DeleteTrainingBudget handles DELETE requests to remove a training budget entry
func DeleteTrainingBudget(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter is required"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID parameter"})
		return
	}

	if err := db.DeleteTrainingBudgetEntry(idInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted successfully"})
}

// GetTrainingHours handles GET requests for total training hours
func GetTrainingHours(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Year parameter is required"})
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	// Get spent hours from timesheet entries
	entries, err := db.GetTrainingEntriesForYear(yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var usedHours int
	for _, entry := range entries {
		usedHours += entry.Training_hours
	}

	// Get total hours from config
	config, err := config.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read configuration"})
		return
	}

	totalHours := config.TrainingHours.YearlyTarget
	availableHours := totalHours - usedHours

	// Return all hours information
	c.JSON(http.StatusOK, gin.H{
		"year":            yearInt,
		"total_hours":     totalHours,
		"used_hours":      usedHours,
		"available_hours": availableHours,
	})
}

// GetVacationHours handles GET requests for total vacation hours
func GetVacationHours(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Year parameter is required"})
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	// Get spent hours from timesheet entries
	usedHours, err := db.GetVacationHoursForYear(yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total hours from config
	config, err := config.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read configuration"})
		return
	}

	totalHours := config.VacationHours.YearlyTarget
	availableHours := totalHours - usedHours

	// Return all hours information
	c.JSON(http.StatusOK, gin.H{
		"year":            yearInt,
		"total_hours":     totalHours,
		"used_hours":      usedHours,
		"available_hours": availableHours,
	})
}

// GetOverview handles GET requests for overview data (training and vacation days left)
func GetOverview(c *gin.Context) {
	year := c.Query("year")
	var yearInt int
	var err error

	if year == "" {
		// Default to current year
		yearInt = time.Now().Year()
	} else {
		yearInt, err = strconv.Atoi(year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
			return
		}
	}

	// Get config
	cfg, err := config.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read configuration"})
		return
	}

	// Calculate training hours
	trainingEntries, err := db.GetTrainingEntriesForYear(yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get training entries"})
		return
	}

	var totalTrainingHours int
	for _, entry := range trainingEntries {
		totalTrainingHours += entry.Training_hours
	}

	trainingHoursLeft := cfg.TrainingHours.YearlyTarget - totalTrainingHours
	trainingDaysLeft := float64(trainingHoursLeft) / 9.0

	// Calculate vacation hours
	vacationEntries, err := db.GetVacationEntriesForYear(yearInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vacation entries"})
		return
	}

	var totalVacationHours int
	for _, entry := range vacationEntries {
		totalVacationHours += entry.Vacation_hours
	}

	vacationHoursLeft := cfg.VacationHours.YearlyTarget - totalVacationHours
	vacationDaysLeft := float64(vacationHoursLeft) / 9.0

	// Return overview data
	c.JSON(http.StatusOK, gin.H{
		"year": yearInt,
		"training": gin.H{
			"total_hours":     cfg.TrainingHours.YearlyTarget,
			"used_hours":      totalTrainingHours,
			"available_hours": trainingHoursLeft,
			"days_left":       trainingDaysLeft,
		},
		"vacation": gin.H{
			"total_hours":     cfg.VacationHours.YearlyTarget,
			"used_hours":      totalVacationHours,
			"available_hours": vacationHoursLeft,
			"days_left":       vacationDaysLeft,
		},
	})
}
