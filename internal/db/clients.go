package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Client represents a client record
type Client struct {
	Id        int
	Name      string
	CreatedAt string
	IsActive  bool
}

// ClientRate represents a rate for a client at a specific date
type ClientRate struct {
	Id            int
	ClientId      int
	HourlyRate    float64
	EffectiveDate string // YYYY-MM-DD format
	Notes         string
	CreatedAt     string
}

// ClientWithRates combines client with their rate history
type ClientWithRates struct {
	Client
	Rates []ClientRate
}

// EarningsEntry represents earnings for a specific timesheet entry
type EarningsEntry struct {
	Date        string
	ClientName  string
	ClientHours int
	HourlyRate  float64
	Earnings    float64
}

// EarningsOverview represents aggregated earnings for a period
type EarningsOverview struct {
	Year          int
	Month         int // 0 for yearly, 1-12 for monthly
	TotalHours    int
	TotalEarnings float64
	Entries       []EarningsEntry
}

// Client CRUD Operations

// GetAllClients retrieves all clients from the database
func GetAllClients() ([]Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients ORDER BY name ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity for typical number of clients
	clients := make([]Client, 0, 10)
	for rows.Next() {
		var client Client
		var isActive int
		if err := rows.Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		client.IsActive = isActive == 1
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

// GetActiveClients retrieves only active clients
func GetActiveClients() ([]Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE is_active = 1 ORDER BY name ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active clients: %w", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity for typical number of active clients
	clients := make([]Client, 0, 10)
	for rows.Next() {
		var client Client
		var isActive int
		if err := rows.Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		client.IsActive = isActive == 1
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

// GetClientById retrieves a specific client by ID
func GetClientById(id int) (Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE id = ?`

	var client Client
	var isActive int
	err := db.QueryRow(query, id).Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, fmt.Errorf("client not found")
		}
		return Client{}, fmt.Errorf("failed to query client: %w", err)
	}
	client.IsActive = isActive == 1

	return client, nil
}

// GetClientByName retrieves a specific client by name
func GetClientByName(name string) (Client, error) {
	query := `SELECT id, name, created_at, is_active FROM clients WHERE name = ?`

	var client Client
	var isActive int
	err := db.QueryRow(query, name).Scan(&client.Id, &client.Name, &client.CreatedAt, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, fmt.Errorf("client not found")
		}
		return Client{}, fmt.Errorf("failed to query client: %w", err)
	}
	client.IsActive = isActive == 1

	return client, nil
}

// AddClient creates a new client and returns the new client ID
func AddClient(client Client) (int, error) {
	query := `INSERT INTO clients (name, created_at, is_active) VALUES (?, ?, ?)`

	createdAt := time.Now().Format("2006-01-02 15:04:05")
	isActive := 0
	if client.IsActive {
		isActive = 1
	}

	result, err := db.Exec(query, client.Name, createdAt, isActive)
	if err != nil {
		return 0, fmt.Errorf("failed to add client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

// UpdateClient updates an existing client
func UpdateClient(client Client) error {
	query := `UPDATE clients SET name = ?, is_active = ? WHERE id = ?`

	isActive := 0
	if client.IsActive {
		isActive = 1
	}

	result, err := db.Exec(query, client.Name, isActive, client.Id)
	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// DeleteClient permanently deletes a client
func DeleteClient(id int) error {
	query := `DELETE FROM clients WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// DeactivateClient sets a client to inactive instead of deleting
func DeactivateClient(id int) error {
	query := `UPDATE clients SET is_active = 0 WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// Client Rate Operations

// GetClientRates retrieves all rates for a specific client
// Returns rates in descending order by effective_date (newest first)
func GetClientRates(clientId int) ([]ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
	          FROM client_rates
	          WHERE client_id = ?
	          ORDER BY effective_date DESC, created_at DESC`

	rows, err := db.Query(query, clientId)
	if err != nil {
		return nil, fmt.Errorf("failed to query client rates: %w", err)
	}
	defer rows.Close()

	// Pre-allocate slice with reasonable capacity for typical number of rate changes
	rates := make([]ClientRate, 0, 10)
	for rows.Next() {
		var rate ClientRate
		if err := rows.Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
			&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan client rate: %w", err)
		}
		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

// GetClientRateById retrieves a specific rate by ID
func GetClientRateById(id int) (ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
	          FROM client_rates WHERE id = ?`

	var rate ClientRate
	err := db.QueryRow(query, id).Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
		&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ClientRate{}, fmt.Errorf("client rate not found")
		}
		return ClientRate{}, fmt.Errorf("failed to query client rate: %w", err)
	}

	return rate, nil
}

// AddClientRate adds a new rate for a client
func AddClientRate(rate ClientRate) error {
	query := `INSERT INTO client_rates (client_id, hourly_rate, effective_date, notes, created_at)
	          VALUES (?, ?, ?, ?, ?)`

	createdAt := time.Now().Format("2006-01-02 15:04:05")

	_, err := db.Exec(query, rate.ClientId, rate.HourlyRate, rate.EffectiveDate, rate.Notes, createdAt)
	if err != nil {
		return fmt.Errorf("failed to add client rate: %w", err)
	}

	return nil
}

// UpdateClientRate updates an existing rate
func UpdateClientRate(rate ClientRate) error {
	query := `UPDATE client_rates
	          SET hourly_rate = ?, effective_date = ?, notes = ?
	          WHERE id = ?`

	result, err := db.Exec(query, rate.HourlyRate, rate.EffectiveDate, rate.Notes, rate.Id)
	if err != nil {
		return fmt.Errorf("failed to update client rate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("client rate not found")
	}

	return nil
}

// DeleteClientRate deletes a specific rate
func DeleteClientRate(id int) error {
	query := `DELETE FROM client_rates WHERE id = ?`

	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client rate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("client rate not found")
	}

	return nil
}

// Rate Lookup Functions

// GetClientRateForDate returns the rate that was effective on the given date
// If multiple rates exist for the same date, returns the most recently created one
func GetClientRateForDate(clientId int, date string) (ClientRate, error) {
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
	          FROM client_rates
	          WHERE client_id = ? AND effective_date <= ?
	          ORDER BY effective_date DESC, created_at DESC
	          LIMIT 1`

	var rate ClientRate
	err := db.QueryRow(query, clientId, date).Scan(&rate.Id, &rate.ClientId,
		&rate.HourlyRate, &rate.EffectiveDate, &rate.Notes, &rate.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return ClientRate{}, fmt.Errorf("no rate found for client on date %s", date)
		}
		return ClientRate{}, fmt.Errorf("failed to query client rate: %w", err)
	}

	return rate, nil
}

// GetClientRateByName is a convenience function that combines client lookup and rate lookup
func GetClientRateByName(clientName string, date string) (float64, error) {
	// First, look up the client by name
	client, err := GetClientByName(clientName)
	if err != nil {
		// Client doesn't exist in clients table - return 0 rate
		return 0.0, nil
	}

	// Then get the rate for the date
	rate, err := GetClientRateForDate(client.Id, date)
	if err != nil {
		// No rate found - return 0
		return 0.0, nil
	}

	return rate.HourlyRate, nil
}

// Earnings Calculation Functions

// rateCache holds cached client and rate information for efficient lookups
type rateCache struct {
	clientsByName map[string]int              // clientName -> clientId
	ratesByClient map[int][]ClientRate        // clientId -> sorted rates (newest first)
}

// buildRateCache creates a cache of all clients and their rates
// This eliminates N+1 queries by loading all data upfront
func buildRateCache() (*rateCache, error) {
	cache := &rateCache{
		clientsByName: make(map[string]int),
		ratesByClient: make(map[int][]ClientRate),
	}

	// Load all clients into cache
	clients, err := GetAllClients()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}
	for _, client := range clients {
		cache.clientsByName[client.Name] = client.Id
	}

	// Load all rates for all clients
	query := `SELECT id, client_id, hourly_rate, effective_date, notes, created_at
	          FROM client_rates
	          ORDER BY client_id, effective_date DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rate ClientRate
		if err := rows.Scan(&rate.Id, &rate.ClientId, &rate.HourlyRate,
			&rate.EffectiveDate, &rate.Notes, &rate.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan rate: %w", err)
		}
		cache.ratesByClient[rate.ClientId] = append(cache.ratesByClient[rate.ClientId], rate)
	}

	return cache, nil
}

// getRateFromCache gets the rate for a client on a specific date from the cache
// Returns the rate that was effective on the given date (most recent rate where effective_date <= date)
func (c *rateCache) getRateFromCache(clientName string, date string) float64 {
	// Get client ID
	clientId, ok := c.clientsByName[clientName]
	if !ok {
		return 0.0
	}

	// Get rates for this client
	rates, ok := c.ratesByClient[clientId]
	if !ok || len(rates) == 0 {
		return 0.0
	}

	// Find the most recent rate where effective_date <= date
	// Rates are sorted by effective_date DESC (newest first)
	for _, rate := range rates {
		if rate.EffectiveDate <= date {
			return rate.HourlyRate
		}
	}

	// No rate found for this date
	return 0.0
}

// CalculateEarningsForYear calculates total earnings for a specific year
func CalculateEarningsForYear(year int) (EarningsOverview, error) {
	// Build rate cache once for all lookups - eliminates N+1 query problem
	cache, err := buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	// Get all timesheet entries for the year with client_hours > 0
	entries, err := GetAllTimesheetEntries(year, 0)
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	// Pre-allocate slice with capacity for typical year's work days (250-365)
	earningsEntries := make([]EarningsEntry, 0, 300)
	var totalHours int
	var totalEarnings float64

	// For each entry, calculate earnings
	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		// Get the rate from cache (no database query!)
		rate := cache.getRateFromCache(entry.Client_name, entry.Date)

		earnings := float64(entry.Client_hours) * rate

		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.Client_name,
			ClientHours: entry.Client_hours,
			HourlyRate:  rate,
			Earnings:    earnings,
		})

		totalHours += entry.Client_hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         0,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

// CalculateEarningsSummaryForYear calculates earnings grouped by client and rate
func CalculateEarningsSummaryForYear(year int) (EarningsOverview, error) {
	// Build rate cache once for all lookups - eliminates N+1 query problem
	cache, err := buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	// Get all timesheet entries for the year with client_hours > 0
	entries, err := GetAllTimesheetEntries(year, 0)
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	// Map to aggregate: key = "ClientName|Rate", value = total hours
	type ClientRateKey struct {
		ClientName string
		Rate       float64
	}
	aggregated := make(map[ClientRateKey]int)

	// Aggregate hours by client and rate
	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		// Get the rate from cache (no database query!)
		rate := cache.getRateFromCache(entry.Client_name, entry.Date)

		key := ClientRateKey{
			ClientName: entry.Client_name,
			Rate:       rate,
		}
		aggregated[key] += entry.Client_hours
	}

	// Convert aggregated data to EarningsEntry slice
	// Pre-allocate for number of unique client-rate combinations
	earningsEntries := make([]EarningsEntry, 0, len(aggregated))
	var totalHours int
	var totalEarnings float64

	for key, hours := range aggregated {
		earnings := float64(hours) * key.Rate
		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        "", // No specific date in summary view
			ClientName:  key.ClientName,
			ClientHours: hours,
			HourlyRate:  key.Rate,
			Earnings:    earnings,
		})
		totalHours += hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         0,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

// CalculateEarningsForMonth calculates total earnings for a specific month
func CalculateEarningsForMonth(year int, month int) (EarningsOverview, error) {
	// Build rate cache once for all lookups - eliminates N+1 query problem
	cache, err := buildRateCache()
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to build rate cache: %w", err)
	}

	// Get all timesheet entries for the month
	entries, err := GetAllTimesheetEntries(year, time.Month(month))
	if err != nil {
		return EarningsOverview{}, fmt.Errorf("failed to get timesheet entries: %w", err)
	}

	// Pre-allocate slice with capacity for typical month's work days (20-30)
	earningsEntries := make([]EarningsEntry, 0, 30)
	var totalHours int
	var totalEarnings float64

	// For each entry, calculate earnings
	for _, entry := range entries {
		if entry.Client_hours <= 0 {
			continue
		}

		// Get the rate from cache (no database query!)
		rate := cache.getRateFromCache(entry.Client_name, entry.Date)

		earnings := float64(entry.Client_hours) * rate

		earningsEntries = append(earningsEntries, EarningsEntry{
			Date:        entry.Date,
			ClientName:  entry.Client_name,
			ClientHours: entry.Client_hours,
			HourlyRate:  rate,
			Earnings:    earnings,
		})

		totalHours += entry.Client_hours
		totalEarnings += earnings
	}

	return EarningsOverview{
		Year:          year,
		Month:         month,
		TotalHours:    totalHours,
		TotalEarnings: totalEarnings,
		Entries:       earningsEntries,
	}, nil
}

// GetClientWithRates retrieves a client along with all their rate history
func GetClientWithRates(clientId int) (ClientWithRates, error) {
	client, err := GetClientById(clientId)
	if err != nil {
		return ClientWithRates{}, err
	}

	rates, err := GetClientRates(clientId)
	if err != nil {
		return ClientWithRates{}, err
	}

	return ClientWithRates{
		Client: client,
		Rates:  rates,
	}, nil
}
