package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

