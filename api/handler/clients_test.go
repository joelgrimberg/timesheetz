package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"timesheet/internal/db"

	"github.com/gin-gonic/gin"
)

func TestGetClients(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add test clients
	client1 := db.Client{
		Name:     "Acme Corp",
		IsActive: true,
	}
	client2 := db.Client{
		Name:     "Inactive Corp",
		IsActive: false,
	}
	db.AddClient(client1)
	db.AddClient(client2)

	// Test getting all clients
	req := httptest.NewRequest("GET", "/api/clients", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetClients(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var clients []db.Client
	if err := json.Unmarshal(w.Body.Bytes(), &clients); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}

	// Test getting only active clients
	req = httptest.NewRequest("GET", "/api/clients?active=true", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetClients(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &clients); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(clients) != 1 {
		t.Errorf("Expected 1 active client, got %d", len(clients))
	}
	if !clients[0].IsActive {
		t.Error("Expected active client")
	}
}

func TestGetClient(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add test client
	client := db.Client{
		Name:     "Acme Corp",
		IsActive: true,
	}
	id, _ := db.AddClient(client)

	// Test getting client by ID
	req := httptest.NewRequest("GET", "/api/clients/"+strconv.Itoa(id), nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(id)}}

	GetClient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result db.Client
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Name != "Acme Corp" {
		t.Errorf("Expected Acme Corp, got %s", result.Name)
	}

	// Test invalid ID
	req = httptest.NewRequest("GET", "/api/clients/invalid", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

	GetClient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Test non-existent ID
	req = httptest.NewRequest("GET", "/api/clients/9999", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "9999"}}

	GetClient(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateClient(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	client := db.Client{
		Name:     "New Client",
		IsActive: true,
	}

	body, _ := json.Marshal(client)
	req := httptest.NewRequest("POST", "/api/clients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	CreateClient(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result db.Client
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Name != "New Client" {
		t.Errorf("Expected New Client, got %s", result.Name)
	}
	if result.Id == 0 {
		t.Error("Expected non-zero ID")
	}

	// Test invalid JSON
	req = httptest.NewRequest("POST", "/api/clients", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	CreateClient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateClient(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client first
	client := db.Client{
		Name:     "Old Name",
		IsActive: true,
	}
	id, _ := db.AddClient(client)

	// Update client
	client.Id = id
	client.Name = "New Name"
	client.IsActive = false

	body, _ := json.Marshal(client)
	req := httptest.NewRequest("PUT", "/api/clients/"+strconv.Itoa(id), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(id)}}

	UpdateClient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify update
	updated, _ := db.GetClientById(id)
	if updated.Name != "New Name" {
		t.Errorf("Expected New Name, got %s", updated.Name)
	}
	if updated.IsActive {
		t.Error("Expected client to be inactive")
	}
}

func TestDeleteClient(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client first
	client := db.Client{
		Name:     "To Delete",
		IsActive: true,
	}
	id, _ := db.AddClient(client)

	req := httptest.NewRequest("DELETE", "/api/clients/"+strconv.Itoa(id), nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(id)}}

	DeleteClient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deactivation (soft delete)
	deactivated, _ := db.GetClientById(id)
	if deactivated.IsActive {
		t.Error("Expected client to be deactivated")
	}
}

func TestGetClientRates(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client
	client := db.Client{
		Name:     "Client A",
		IsActive: true,
	}
	clientId, _ := db.AddClient(client)

	// Add rates
	rate1 := db.ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Initial rate",
	}
	rate2 := db.ClientRate{
		ClientId:      clientId,
		HourlyRate:    120.00,
		EffectiveDate: "2024-07-01",
		Notes:         "Rate increase",
	}
	db.AddClientRate(rate1)
	db.AddClientRate(rate2)

	req := httptest.NewRequest("GET", "/api/clients/"+strconv.Itoa(clientId)+"/rates", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(clientId)}}

	GetClientRates(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var rates []db.ClientRate
	if err := json.Unmarshal(w.Body.Bytes(), &rates); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(rates))
	}
}

func TestCreateClientRate(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client
	client := db.Client{
		Name:     "Client A",
		IsActive: true,
	}
	clientId, _ := db.AddClient(client)

	rate := db.ClientRate{
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Test rate",
	}

	body, _ := json.Marshal(rate)
	req := httptest.NewRequest("POST", "/api/clients/"+strconv.Itoa(clientId)+"/rates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(clientId)}}

	CreateClientRate(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result db.ClientRate
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.ClientId != clientId {
		t.Errorf("Expected client_id %d, got %d", clientId, result.ClientId)
	}
	if result.HourlyRate != 100.00 {
		t.Errorf("Expected rate 100.00, got %.2f", result.HourlyRate)
	}
}

func TestUpdateClientRate(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client and rate
	client := db.Client{
		Name:     "Client A",
		IsActive: true,
	}
	clientId, _ := db.AddClient(client)

	rate := db.ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
		Notes:         "Original",
	}
	db.AddClientRate(rate)

	// Get rate to get ID
	rates, _ := db.GetClientRates(clientId)
	if len(rates) == 0 {
		t.Fatal("No rates found")
	}
	rateId := rates[0].Id

	// Update rate
	rate.Id = rateId
	rate.HourlyRate = 120.00
	rate.Notes = "Updated"

	body, _ := json.Marshal(rate)
	req := httptest.NewRequest("PUT", "/api/client-rates/"+strconv.Itoa(rateId), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(rateId)}}

	UpdateClientRate(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify update
	updated, _ := db.GetClientRateById(rateId)
	if updated.HourlyRate != 120.00 {
		t.Errorf("Expected rate 120.00, got %.2f", updated.HourlyRate)
	}
	if updated.Notes != "Updated" {
		t.Errorf("Expected notes 'Updated', got %s", updated.Notes)
	}
}

func TestDeleteClientRate(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client and rate
	client := db.Client{
		Name:     "Client A",
		IsActive: true,
	}
	clientId, _ := db.AddClient(client)

	rate := db.ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	}
	db.AddClientRate(rate)

	// Get rate to get ID
	rates, _ := db.GetClientRates(clientId)
	if len(rates) == 0 {
		t.Fatal("No rates found")
	}
	rateId := rates[0].Id

	req := httptest.NewRequest("DELETE", "/api/client-rates/"+strconv.Itoa(rateId), nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: strconv.Itoa(rateId)}}

	DeleteClientRate(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deletion
	rates, _ = db.GetClientRates(clientId)
	if len(rates) != 0 {
		t.Errorf("Expected 0 rates after deletion, got %d", len(rates))
	}
}

func TestGetEarnings(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add client
	client := db.Client{
		Name:     "Acme Corp",
		IsActive: true,
	}
	clientId, _ := db.AddClient(client)

	// Add rate
	rate := db.ClientRate{
		ClientId:      clientId,
		HourlyRate:    100.00,
		EffectiveDate: "2024-01-01",
	}
	db.AddClientRate(rate)

	// Add timesheet entries
	entry1 := db.TimesheetEntry{
		Date:         "2024-01-15",
		Client_name:  "Acme Corp",
		Client_hours: 8,
	}
	entry2 := db.TimesheetEntry{
		Date:         "2024-01-16",
		Client_name:  "Acme Corp",
		Client_hours: 6,
	}
	db.AddTimesheetEntry(entry1)
	db.AddTimesheetEntry(entry2)

	// Test yearly earnings
	req := httptest.NewRequest("GET", "/api/earnings?year=2024", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetEarnings(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify structure
	if result["year"] != float64(2024) {
		t.Errorf("Expected year 2024, got %v", result["year"])
	}
	if result["total_hours"] != float64(14) {
		t.Errorf("Expected total_hours 14, got %v", result["total_hours"])
	}

	// Verify Euro formatting
	totalEarnings, ok := result["total_earnings"].(string)
	if !ok {
		t.Fatalf("total_earnings is not a string: %v", result["total_earnings"])
	}
	if !strings.HasPrefix(totalEarnings, "€") {
		t.Errorf("Expected total_earnings to start with €, got %s", totalEarnings)
	}
	if !strings.Contains(totalEarnings, ",") {
		t.Errorf("Expected total_earnings to use comma separator, got %s", totalEarnings)
	}

	// Verify entries have proper Euro formatting
	entries, ok := result["entries"].([]interface{})
	if !ok || len(entries) == 0 {
		t.Fatalf("Expected entries array, got %v", result["entries"])
	}
	firstEntry := entries[0].(map[string]interface{})
	earnings, ok := firstEntry["earnings"].(string)
	if !ok {
		t.Fatalf("earnings is not a string: %v", firstEntry["earnings"])
	}
	if !strings.HasPrefix(earnings, "€") {
		t.Errorf("Expected earnings to start with €, got %s", earnings)
	}

	// Test monthly earnings
	req = httptest.NewRequest("GET", "/api/earnings?year=2024&month=1", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetEarnings(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["month"] != float64(1) {
		t.Errorf("Expected month 1, got %v", result["month"])
	}

	// Test invalid year
	req = httptest.NewRequest("GET", "/api/earnings?year=invalid", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetEarnings(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Test invalid month
	req = httptest.NewRequest("GET", "/api/earnings?year=2024&month=13", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetEarnings(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetEarningsDefaultYear(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Test without year parameter (should default to current year)
	req := httptest.NewRequest("GET", "/api/earnings", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetEarnings(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have a year field (current year)
	if _, ok := result["year"]; !ok {
		t.Error("Expected year field in response")
	}
}
