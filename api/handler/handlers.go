package handler

import (
	"net/http"
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
