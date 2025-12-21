package handler

import (
	"net/http"
	"strconv"
	"time"
	"timesheet/internal/db"
	"timesheet/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetClients handles GET /api/clients
// Returns all clients or only active clients if ?active=true query param is provided
func GetClients(c *gin.Context) {
	activeOnly := c.Query("active") == "true"

	var clients []db.Client
	var err error

	if activeOnly {
		clients, err = db.GetActiveClients()
	} else {
		clients, err = db.GetAllClients()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, clients)
}

// GetClient handles GET /api/clients/:id
// Returns a specific client by ID
func GetClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	client, err := db.GetClientById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// CreateClient handles POST /api/clients
// Creates a new client
func CreateClient(c *gin.Context) {
	var client db.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := db.AddClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the created client with its ID
	client.Id = id
	c.JSON(http.StatusCreated, client)
}

// UpdateClient handles PUT /api/clients/:id
// Updates an existing client
func UpdateClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var client db.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID from the URL is used
	client.Id = id

	if err := db.UpdateClient(client); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client)
}

// DeleteClient handles DELETE /api/clients/:id
// Deletes a client (or deactivates if you prefer soft delete)
func DeleteClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	// Use deactivate instead of hard delete to preserve historical data
	if err := db.DeactivateClient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deactivated successfully"})
}

// GetClientRates handles GET /api/clients/:id/rates
// Returns all rates for a specific client
func GetClientRates(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	rates, err := db.GetClientRates(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rates)
}

// CreateClientRate handles POST /api/clients/:id/rates
// Adds a new rate for a client
func CreateClientRate(c *gin.Context) {
	idStr := c.Param("id")
	clientId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var rate db.ClientRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the client_id from the URL is used
	rate.ClientId = clientId

	if err := db.AddClientRate(rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rate)
}

// UpdateClientRate handles PUT /api/client-rates/:id
// Updates an existing rate
func UpdateClientRate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rate ID"})
		return
	}

	var rate db.ClientRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the ID from the URL is used
	rate.Id = id

	if err := db.UpdateClientRate(rate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
}

// DeleteClientRate handles DELETE /api/client-rates/:id
// Deletes a specific rate
func DeleteClientRate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rate ID"})
		return
	}

	if err := db.DeleteClientRate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rate deleted successfully"})
}

// GetEarnings handles GET /api/earnings?year=YYYY&month=MM
// Returns earnings overview for a year or specific month
func GetEarnings(c *gin.Context) {
	yearStr := c.Query("year")
	if yearStr == "" {
		// Default to current year
		yearStr = strconv.Itoa(time.Now().Year())
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	monthStr := c.Query("month")
	summaryStr := c.Query("summary")
	var overview db.EarningsOverview

	if monthStr != "" {
		// Calculate for specific month
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month (must be 1-12)"})
			return
		}

		overview, err = db.CalculateEarningsForMonth(year, month)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if summaryStr == "true" {
		// Calculate summary for entire year (grouped by client and rate)
		overview, err = db.CalculateEarningsSummaryForYear(year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Calculate detailed for entire year
		overview, err = db.CalculateEarningsForYear(year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Format response with Euro formatting
	response := formatEarningsResponse(overview)
	c.JSON(http.StatusOK, response)
}

// formatEarningsResponse formats the earnings overview with Euro currency formatting
func formatEarningsResponse(overview db.EarningsOverview) gin.H {
	// Format individual entries
	var formattedEntries []gin.H
	for _, entry := range overview.Entries {
		formattedEntries = append(formattedEntries, gin.H{
			"date":         entry.Date,
			"client_name":  entry.ClientName,
			"client_hours": entry.ClientHours,
			"hourly_rate":  utils.FormatEuro(entry.HourlyRate),
			"earnings":     utils.FormatEuro(entry.Earnings),
		})
	}

	return gin.H{
		"year":           overview.Year,
		"month":          overview.Month,
		"total_hours":    overview.TotalHours,
		"total_earnings": utils.FormatEuro(overview.TotalEarnings),
		"entries":        formattedEntries,
	}
}
