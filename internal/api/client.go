package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"
	"timesheet/internal/logging"
)

// Client is an HTTP client for the timesheet API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// makeRequest makes an HTTP request and returns the response body
func (c *Client) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetAllTimesheetEntries retrieves all timesheet entries
func (c *Client) GetAllTimesheetEntries(year int, month time.Month) ([]db.TimesheetEntry, error) {
	endpoint := "/api/timesheet"
	if year != 0 && month != 0 {
		// Note: The API currently doesn't support year/month filtering
		// We'll get all entries and filter client-side if needed
		// This could be enhanced later
	}

	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var entries []db.TimesheetEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Filter by year/month if specified
	if year != 0 && month != 0 {
		filtered := []db.TimesheetEntry{}
		startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		endDate := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.UTC).Format("2006-01-02")
		for _, entry := range entries {
			if entry.Date >= startDate && entry.Date <= endDate {
				filtered = append(filtered, entry)
			}
		}
		return filtered, nil
	}

	return entries, nil
}

// GetTimesheetEntryByDate retrieves a timesheet entry by date
func (c *Client) GetTimesheetEntryByDate(date string) (db.TimesheetEntry, error) {
	// Get all entries and find the one with matching date
	entries, err := c.GetAllTimesheetEntries(0, 0)
	if err != nil {
		return db.TimesheetEntry{}, err
	}

	for _, entry := range entries {
		if entry.Date == date {
			return entry, nil
		}
	}

	return db.TimesheetEntry{}, fmt.Errorf("entry not found for date %s", date)
}

// AddTimesheetEntry creates a new timesheet entry
func (c *Client) AddTimesheetEntry(entry db.TimesheetEntry) error {
	_, err := c.makeRequest("POST", "/api/timesheet", entry)
	return err
}

// UpdateTimesheetEntry updates an existing timesheet entry
func (c *Client) UpdateTimesheetEntry(entry db.TimesheetEntry) error {
	if entry.Id == 0 {
		return fmt.Errorf("entry ID is required for update")
	}
	_, err := c.makeRequest("PUT", fmt.Sprintf("/api/timesheet/%d", entry.Id), entry)
	return err
}

// UpdateTimesheetEntryById updates specific fields of a timesheet entry by ID
func (c *Client) UpdateTimesheetEntryById(id string, data map[string]any) error {
	// Convert to a partial entry that the API expects
	_, err := c.makeRequest("PUT", fmt.Sprintf("/api/timesheet/%s", id), data)
	return err
}

// DeleteTimesheetEntryByDate deletes a timesheet entry by date
func (c *Client) DeleteTimesheetEntryByDate(date string) error {
	// First, get the entry to find its ID
	entry, err := c.GetTimesheetEntryByDate(date)
	if err != nil {
		return err
	}
	return c.DeleteTimesheetEntry(strconv.Itoa(entry.Id))
}

// DeleteTimesheetEntry deletes a timesheet entry by ID
func (c *Client) DeleteTimesheetEntry(id string) error {
	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/timesheet/%s", id), nil)
	return err
}

// GetLastClientName returns the last client name
func (c *Client) GetLastClientName() (string, error) {
	data, err := c.makeRequest("GET", "/api/last-client", nil)
	if err != nil {
		return "", err
	}

	var result struct {
		ClientName string `json:"client_name"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.ClientName, nil
}

// GetTrainingEntriesForYear retrieves training entries for a year
func (c *Client) GetTrainingEntriesForYear(year int) ([]db.TimesheetEntry, error) {
	// Get all entries and filter for training hours > 0
	entries, err := c.GetAllTimesheetEntries(0, 0)
	if err != nil {
		return nil, err
	}

	filtered := []db.TimesheetEntry{}
	yearStr := strconv.Itoa(year)
	for _, entry := range entries {
		if len(entry.Date) >= 4 && entry.Date[:4] == yearStr && entry.Training_hours > 0 {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetVacationEntriesForYear retrieves vacation entries for a year
func (c *Client) GetVacationEntriesForYear(year int) ([]db.TimesheetEntry, error) {
	// Get all entries and filter for vacation hours > 0
	entries, err := c.GetAllTimesheetEntries(0, 0)
	if err != nil {
		return nil, err
	}

	filtered := []db.TimesheetEntry{}
	yearStr := strconv.Itoa(year)
	for _, entry := range entries {
		if len(entry.Date) >= 4 && entry.Date[:4] == yearStr && entry.Vacation_hours > 0 {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetVacationHoursForYear returns total vacation hours for a year
func (c *Client) GetVacationHoursForYear(year int) (int, error) {
	entries, err := c.GetVacationEntriesForYear(year)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, entry := range entries {
		total += entry.Vacation_hours
	}

	return total, nil
}

// GetVacationCarryoverForYear retrieves carryover hours for a specific year
func (c *Client) GetVacationCarryoverForYear(year int) (db.VacationCarryover, error) {
	endpoint := fmt.Sprintf("/api/vacation-carryover?year=%d", year)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return db.VacationCarryover{}, err
	}

	var carryover db.VacationCarryover
	if err := json.Unmarshal(data, &carryover); err != nil {
		return db.VacationCarryover{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return carryover, nil
}

// SetVacationCarryover creates or updates carryover for a year
func (c *Client) SetVacationCarryover(carryover db.VacationCarryover) error {
	endpoint := "/api/vacation-carryover"
	_, err := c.makeRequest("POST", endpoint, carryover)
	return err
}

// DeleteVacationCarryover removes carryover for a year
func (c *Client) DeleteVacationCarryover(year int) error {
	endpoint := fmt.Sprintf("/api/vacation-carryover?year=%d", year)
	_, err := c.makeRequest("DELETE", endpoint, nil)
	return err
}

// GetVacationSummaryForYear retrieves comprehensive vacation info for a year
func (c *Client) GetVacationSummaryForYear(year int) (db.VacationSummary, error) {
	endpoint := fmt.Sprintf("/api/vacation-summary?year=%d", year)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return db.VacationSummary{}, err
	}

	var summary db.VacationSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return db.VacationSummary{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return summary, nil
}

// GetTrainingBudgetEntriesForYear retrieves training budget entries for a year
func (c *Client) GetTrainingBudgetEntriesForYear(year int) ([]db.TrainingBudgetEntry, error) {
	endpoint := fmt.Sprintf("/api/training-budget?year=%d", year)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var entries []db.TrainingBudgetEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return entries, nil
}

// AddTrainingBudgetEntry creates a new training budget entry
func (c *Client) AddTrainingBudgetEntry(entry db.TrainingBudgetEntry) error {
	_, err := c.makeRequest("POST", "/api/training-budget", entry)
	return err
}

// UpdateTrainingBudgetEntry updates an existing training budget entry
func (c *Client) UpdateTrainingBudgetEntry(entry db.TrainingBudgetEntry) error {
	_, err := c.makeRequest("PUT", "/api/training-budget", entry)
	return err
}

// DeleteTrainingBudgetEntry deletes a training budget entry
func (c *Client) DeleteTrainingBudgetEntry(id int) error {
	endpoint := fmt.Sprintf("/api/training-budget?id=%d", id)
	_, err := c.makeRequest("DELETE", endpoint, nil)
	return err
}

// GetTrainingBudgetEntry retrieves a training budget entry by ID
func (c *Client) GetTrainingBudgetEntry(id int) (db.TrainingBudgetEntry, error) {
	// Get all entries for the year and find the one with matching ID
	// We need to get entries from a reasonable year range
	currentYear := time.Now().Year()
	entries, err := c.GetTrainingBudgetEntriesForYear(currentYear)
	if err != nil {
		// Try previous year as fallback
		entries, err = c.GetTrainingBudgetEntriesForYear(currentYear - 1)
		if err != nil {
			return db.TrainingBudgetEntry{}, err
		}
	}

	for _, entry := range entries {
		if entry.Id == id {
			return entry, nil
		}
	}

	return db.TrainingBudgetEntry{}, fmt.Errorf("training budget entry not found with id %d", id)
}

// GetTrainingBudgetEntryByDate retrieves a training budget entry by date
func (c *Client) GetTrainingBudgetEntryByDate(date string) (db.TrainingBudgetEntry, error) {
	// Extract year from date
	if len(date) < 4 {
		return db.TrainingBudgetEntry{}, fmt.Errorf("invalid date format")
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return db.TrainingBudgetEntry{}, fmt.Errorf("invalid year in date: %w", err)
	}

	entries, err := c.GetTrainingBudgetEntriesForYear(year)
	if err != nil {
		return db.TrainingBudgetEntry{}, err
	}

	for _, entry := range entries {
		if entry.Date == date {
			return entry, nil
		}
	}

	return db.TrainingBudgetEntry{}, fmt.Errorf("training budget entry not found for date %s", date)
}

// Client Management Methods

// GetAllClients retrieves all clients
func (c *Client) GetAllClients() ([]db.Client, error) {
	data, err := c.makeRequest("GET", "/api/clients", nil)
	if err != nil {
		return nil, err
	}

	var clients []db.Client
	if err := json.Unmarshal(data, &clients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return clients, nil
}

// GetActiveClients retrieves only active clients
func (c *Client) GetActiveClients() ([]db.Client, error) {
	data, err := c.makeRequest("GET", "/api/clients?active=true", nil)
	if err != nil {
		return nil, err
	}

	var clients []db.Client
	if err := json.Unmarshal(data, &clients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return clients, nil
}

// GetClientById retrieves a specific client by ID
func (c *Client) GetClientById(id int) (db.Client, error) {
	data, err := c.makeRequest("GET", fmt.Sprintf("/api/clients/%d", id), nil)
	if err != nil {
		return db.Client{}, err
	}

	var client db.Client
	if err := json.Unmarshal(data, &client); err != nil {
		return db.Client{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return client, nil
}

// GetClientByName retrieves a specific client by name
func (c *Client) GetClientByName(name string) (db.Client, error) {
	// Get all clients and find by name (API doesn't have direct name lookup)
	clients, err := c.GetAllClients()
	if err != nil {
		return db.Client{}, err
	}

	for _, client := range clients {
		if client.Name == name {
			return client, nil
		}
	}

	return db.Client{}, fmt.Errorf("client not found: %s", name)
}

// AddClient creates a new client
func (c *Client) AddClient(client db.Client) (int, error) {
	data, err := c.makeRequest("POST", "/api/clients", client)
	if err != nil {
		return 0, err
	}

	var result db.Client
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Id, nil
}

// UpdateClient updates an existing client
func (c *Client) UpdateClient(client db.Client) error {
	_, err := c.makeRequest("PUT", fmt.Sprintf("/api/clients/%d", client.Id), client)
	return err
}

// DeleteClient deletes a client
func (c *Client) DeleteClient(id int) error {
	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/clients/%d", id), nil)
	return err
}

// DeactivateClient deactivates a client
func (c *Client) DeactivateClient(id int) error {
	// The API DeleteClient actually does deactivation
	return c.DeleteClient(id)
}

// Client Rate Methods

// GetClientRates retrieves all rates for a specific client
func (c *Client) GetClientRates(clientId int) ([]db.ClientRate, error) {
	data, err := c.makeRequest("GET", fmt.Sprintf("/api/clients/%d/rates", clientId), nil)
	if err != nil {
		return nil, err
	}

	var rates []db.ClientRate
	if err := json.Unmarshal(data, &rates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return rates, nil
}

// GetClientRateById retrieves a specific rate by ID
func (c *Client) GetClientRateById(id int) (db.ClientRate, error) {
	// Get all clients and search for the rate
	// This is inefficient but works without a dedicated endpoint
	clients, err := c.GetAllClients()
	if err != nil {
		return db.ClientRate{}, err
	}

	for _, client := range clients {
		rates, err := c.GetClientRates(client.Id)
		if err != nil {
			continue
		}

		for _, rate := range rates {
			if rate.Id == id {
				return rate, nil
			}
		}
	}

	return db.ClientRate{}, fmt.Errorf("rate not found with id %d", id)
}

// AddClientRate adds a new rate for a client
func (c *Client) AddClientRate(rate db.ClientRate) error {
	_, err := c.makeRequest("POST", fmt.Sprintf("/api/clients/%d/rates", rate.ClientId), rate)
	return err
}

// UpdateClientRate updates an existing rate
func (c *Client) UpdateClientRate(rate db.ClientRate) error {
	_, err := c.makeRequest("PUT", fmt.Sprintf("/api/client-rates/%d", rate.Id), rate)
	return err
}

// DeleteClientRate deletes a specific rate
func (c *Client) DeleteClientRate(id int) error {
	_, err := c.makeRequest("DELETE", fmt.Sprintf("/api/client-rates/%d", id), nil)
	return err
}

// GetClientRateForDate returns the rate that was effective on the given date
func (c *Client) GetClientRateForDate(clientId int, date string) (db.ClientRate, error) {
	rates, err := c.GetClientRates(clientId)
	if err != nil {
		return db.ClientRate{}, err
	}

	// Find the most recent rate that's effective on or before the date
	var validRate db.ClientRate
	found := false

	for _, rate := range rates {
		if rate.EffectiveDate <= date {
			if !found || rate.EffectiveDate > validRate.EffectiveDate {
				validRate = rate
				found = true
			}
		}
	}

	if !found {
		return db.ClientRate{}, fmt.Errorf("no rate found for client %d on date %s", clientId, date)
	}

	return validRate, nil
}

// GetClientRateByName is a convenience function that combines client lookup and rate lookup
func (c *Client) GetClientRateByName(clientName string, date string) (float64, error) {
	client, err := c.GetClientByName(clientName)
	if err != nil {
		return 0.0, nil // Client doesn't exist, return 0 rate
	}

	rate, err := c.GetClientRateForDate(client.Id, date)
	if err != nil {
		return 0.0, nil // No rate found, return 0
	}

	return rate.HourlyRate, nil
}

// Earnings Methods

// CalculateEarningsForYear calculates total earnings for a specific year
func (c *Client) CalculateEarningsForYear(year int) (db.EarningsOverview, error) {
	endpoint := fmt.Sprintf("/api/earnings?year=%d", year)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return db.EarningsOverview{}, err
	}

	// The API returns formatted data, we need to parse it
	var response struct {
		Year          int    `json:"year"`
		Month         int    `json:"month"`
		TotalHours    int    `json:"total_hours"`
		TotalEarnings string `json:"total_earnings"` // Formatted as Euro string
		Entries       []struct {
			Date        string `json:"date"`
			ClientName  string `json:"client_name"`
			ClientHours int    `json:"client_hours"`
			HourlyRate  string `json:"hourly_rate"` // Formatted as Euro string
			Earnings    string `json:"earnings"`    // Formatted as Euro string
		} `json:"entries"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return db.EarningsOverview{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert back to EarningsOverview with parsed Euro values
	overview := db.EarningsOverview{
		Year:       response.Year,
		Month:      response.Month,
		TotalHours: response.TotalHours,
	}

	// Parse total earnings
	totalEarnings, _ := parseEuroFromAPI(response.TotalEarnings)
	overview.TotalEarnings = totalEarnings

	// Parse entries
	for _, entry := range response.Entries {
		hourlyRate, _ := parseEuroFromAPI(entry.HourlyRate)
		earnings, _ := parseEuroFromAPI(entry.Earnings)

		overview.Entries = append(overview.Entries, db.EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.ClientName,
			ClientHours: entry.ClientHours,
			HourlyRate:  hourlyRate,
			Earnings:    earnings,
		})
	}

	return overview, nil
}

// CalculateEarningsSummaryForYear calculates earnings summary grouped by client and rate
func (c *Client) CalculateEarningsSummaryForYear(year int) (db.EarningsOverview, error) {
	endpoint := fmt.Sprintf("/api/earnings?year=%d&summary=true", year)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return db.EarningsOverview{}, err
	}

	// Same parsing logic as CalculateEarningsForYear
	var response struct {
		Year          int    `json:"year"`
		Month         int    `json:"month"`
		TotalHours    int    `json:"total_hours"`
		TotalEarnings string `json:"total_earnings"`
		Entries       []struct {
			Date        string `json:"date"`
			ClientName  string `json:"client_name"`
			ClientHours int    `json:"client_hours"`
			HourlyRate  string `json:"hourly_rate"`
			Earnings    string `json:"earnings"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return db.EarningsOverview{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert back to EarningsOverview with parsed Euro values
	overview := db.EarningsOverview{
		Year:       response.Year,
		Month:      response.Month,
		TotalHours: response.TotalHours,
	}

	// Parse total earnings
	totalEarnings, _ := parseEuroFromAPI(response.TotalEarnings)
	overview.TotalEarnings = totalEarnings

	// Parse entries
	for _, entry := range response.Entries {
		hourlyRate, _ := parseEuroFromAPI(entry.HourlyRate)
		earnings, _ := parseEuroFromAPI(entry.Earnings)

		overview.Entries = append(overview.Entries, db.EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.ClientName,
			ClientHours: entry.ClientHours,
			HourlyRate:  hourlyRate,
			Earnings:    earnings,
		})
	}

	return overview, nil
}

// CalculateEarningsForMonth calculates total earnings for a specific month
func (c *Client) CalculateEarningsForMonth(year int, month int) (db.EarningsOverview, error) {
	endpoint := fmt.Sprintf("/api/earnings?year=%d&month=%d", year, month)
	data, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return db.EarningsOverview{}, err
	}

	// Same parsing logic as CalculateEarningsForYear
	var response struct {
		Year          int    `json:"year"`
		Month         int    `json:"month"`
		TotalHours    int    `json:"total_hours"`
		TotalEarnings string `json:"total_earnings"`
		Entries       []struct {
			Date        string `json:"date"`
			ClientName  string `json:"client_name"`
			ClientHours int    `json:"client_hours"`
			HourlyRate  string `json:"hourly_rate"`
			Earnings    string `json:"earnings"`
		} `json:"entries"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return db.EarningsOverview{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	overview := db.EarningsOverview{
		Year:       response.Year,
		Month:      response.Month,
		TotalHours: response.TotalHours,
	}

	totalEarnings, _ := parseEuroFromAPI(response.TotalEarnings)
	overview.TotalEarnings = totalEarnings

	for _, entry := range response.Entries {
		hourlyRate, _ := parseEuroFromAPI(entry.HourlyRate)
		earnings, _ := parseEuroFromAPI(entry.Earnings)

		overview.Entries = append(overview.Entries, db.EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.ClientName,
			ClientHours: entry.ClientHours,
			HourlyRate:  hourlyRate,
			Earnings:    earnings,
		})
	}

	return overview, nil
}

// GetClientWithRates retrieves a client along with all their rate history
func (c *Client) GetClientWithRates(clientId int) (db.ClientWithRates, error) {
	client, err := c.GetClientById(clientId)
	if err != nil {
		return db.ClientWithRates{}, err
	}

	rates, err := c.GetClientRates(clientId)
	if err != nil {
		return db.ClientWithRates{}, err
	}

	return db.ClientWithRates{
		Client: client,
		Rates:  rates,
	}, nil
}

// parseEuroFromAPI parses a Euro string from the API (e.g., "€100,50") to float64
func parseEuroFromAPI(euroStr string) (float64, error) {
	// We can just use the existing ParseEuro function from utils
	// But since we're in the api package and want to avoid circular imports,
	// we'll implement a simple version here

	// Remove € symbol (works with UTF-8)
	cleanStr := strings.TrimSpace(euroStr)
	cleanStr = strings.TrimPrefix(cleanStr, "€")
	cleanStr = strings.TrimSpace(cleanStr)

	// Replace comma with dot
	cleanStr = strings.Replace(cleanStr, ",", ".", 1)

	var value float64
	_, err := fmt.Sscanf(cleanStr, "%f", &value)
	return value, err
}

// Ping checks if the API is accessible
func (c *Client) Ping() error {
	_, err := c.makeRequest("GET", "/health", nil)
	return err
}

// GetClient returns a configured API client or nil if not in remote mode
func GetClient() (*Client, error) {
	apiMode := config.GetAPIMode()
	if apiMode == "local" {
		return nil, nil
	}

	baseURL := config.GetAPIBaseURL()
	if baseURL == "" {
		return nil, fmt.Errorf("apiMode is '%s' but apiBaseURL is not configured", apiMode)
	}

	client := NewClient(baseURL)

	// Test connection
	if err := client.Ping(); err != nil {
		logging.Log("Warning: Failed to ping remote API at %s: %v", baseURL, err)
		// Don't fail here, let the caller decide
	}

	return client, nil
}
