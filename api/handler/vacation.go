package handler

import (
    "net/http"
    "strconv"
    "time"
    "timesheet/internal/config"
    "timesheet/internal/db"

    "github.com/gin-gonic/gin"
)

// GetVacation returns all vacation entries for the current year
func GetVacation(c *gin.Context) {
    year := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
    yearInt, err := strconv.Atoi(year)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
        return
    }

    entries, err := db.GetVacationEntriesForYear(yearInt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Get the yearly target from config
    config, err := config.GetConfig()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get config"})
        return
    }

    // Get total hours used
    totalHours, err := db.GetVacationHoursForYear(yearInt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total hours"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "entries":     entries,
        "yearlyTarget": config.VacationHours.YearlyTarget,
        "totalHours":  totalHours,
        "remaining":   config.VacationHours.YearlyTarget - totalHours,
    })
}

// CreateVacation creates a new vacation entry
func CreateVacation(c *gin.Context) {
    var entry db.VacationEntry
    if err := c.ShouldBindJSON(&entry); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := db.AddVacationEntry(entry); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Vacation entry created"})
}

// UpdateVacation updates an existing vacation entry
func UpdateVacation(c *gin.Context) {
    id := c.Param("id")
    idInt, err := strconv.Atoi(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var entry db.VacationEntry
    if err := c.ShouldBindJSON(&entry); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    entry.Id = idInt

    if err := db.AddVacationEntry(entry); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Vacation entry updated"})
}

// DeleteVacation deletes a vacation entry
func DeleteVacation(c *gin.Context) {
    id := c.Param("id")
    idInt, err := strconv.Atoi(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    if err := db.DeleteVacationEntry(idInt); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Vacation entry deleted"})
} 