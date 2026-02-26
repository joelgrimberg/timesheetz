package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"timesheet/internal/config"
	"timesheet/internal/db"

	"github.com/gin-gonic/gin"
)

func setupHandlerTest(t *testing.T) string {
	// Use in-memory database for testing
	dbPath := ":memory:"

	// Initialize (which also opens the connection for in-memory)
	if err := db.InitializeDatabase(dbPath); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Redirect config to a temp directory so we never touch the real config file
	tmpDir := t.TempDir()
	tmpConfigPath := filepath.Join(tmpDir, "config.json")
	config.SetConfigPathOverride(tmpConfigPath)

	// Create test config
	testConfig := config.Config{
		TrainingHours: config.TrainingHours{YearlyTarget: 36},
		VacationHours: config.VacationHours{YearlyTarget: 20},
	}
	if err := config.SaveConfig(testConfig); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	return dbPath
}

func teardownHandlerTest(t *testing.T, dbPath string) {
	db.Close()
	config.SetConfigPathOverride("")
}

func TestGetTimesheet(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add test entry
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	// Create request
	req := httptest.NewRequest("GET", "/api/timesheet", nil)
	w := httptest.NewRecorder()

	// Create Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	GetTimesheet(c)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var entries []db.TimesheetEntry
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestCreateTimesheet(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}

	body, _ := json.Marshal(entry)
	req := httptest.NewRequest("POST", "/api/timesheet", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	CreateTimesheet(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result db.TimesheetEntry
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result.Client_name != "Client A" {
		t.Errorf("Expected Client A, got %s", result.Client_name)
	}
}

func TestUpdateTimesheet(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entry first
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	// Get entry to get ID
	result, _ := db.GetTimesheetEntryByDate("2024-01-15")
	entry.Id = result.Id
	entry.Client_hours = 6
	entry.Client_name = result.Client_name // Keep original client name
	body, _ := json.Marshal(entry)
	idStr := strconv.Itoa(result.Id)
	req := httptest.NewRequest("PUT", "/api/timesheet/"+idStr, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: idStr}}

	UpdateTimesheet(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTimesheet(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entry first
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	req := httptest.NewRequest("DELETE", "/api/timesheet/1", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

	DeleteTimesheet(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deletion
	_, err := db.GetTimesheetEntryByDate("2024-01-15")
	if err == nil {
		t.Error("Entry should be deleted")
	}
}

func TestGetLastClientName(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entry
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   8,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	req := httptest.NewRequest("GET", "/api/last-client", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetLastClientName(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["client_name"] != "Client A" {
		t.Errorf("Expected Client A, got %s", result["client_name"])
	}
}

func TestGetTrainingBudget(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add training budget entry
	entry := db.TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}
	db.AddTrainingBudgetEntry(entry)

	req := httptest.NewRequest("GET", "/api/training-budget?year=2024", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetTrainingBudget(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var entries []db.TrainingBudgetEntry
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Test missing year parameter
	req = httptest.NewRequest("GET", "/api/training-budget", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetTrainingBudget(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateTrainingBudget(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	entry := db.TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}

	body, _ := json.Marshal(entry)
	req := httptest.NewRequest("POST", "/api/training-budget", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	CreateTrainingBudget(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestUpdateTrainingBudget(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entry first
	entry := db.TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}
	db.AddTrainingBudgetEntry(entry)

	// Get entry to get ID
	result, _ := db.GetTrainingBudgetEntryByDate("2024-01-15")
	entry.Id = result.Id
	entry.Training_name = "Training B"

	body, _ := json.Marshal(entry)
	req := httptest.NewRequest("PUT", "/api/training-budget", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	UpdateTrainingBudget(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDeleteTrainingBudget(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entry first
	entry := db.TrainingBudgetEntry{
		Date:             "2024-01-15",
		Training_name:    "Training A",
		Hours:            8,
		Cost_without_vat: 100.0,
	}
	db.AddTrainingBudgetEntry(entry)

	req := httptest.NewRequest("DELETE", "/api/training-budget?id=1", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	DeleteTrainingBudget(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test missing ID
	req = httptest.NewRequest("DELETE", "/api/training-budget", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	DeleteTrainingBudget(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetTrainingHours(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add training entry
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 4,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	req := httptest.NewRequest("GET", "/api/training-hours?year=2024", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetTrainingHours(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if w.Code == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if usedHours, ok := result["used_hours"].(float64); ok {
			if int(usedHours) != 4 {
				t.Errorf("Expected 4 used hours, got %v", usedHours)
			}
		} else {
			t.Errorf("used_hours is not a number: %v", result["used_hours"])
		}
	}
}

func TestGetVacationHours(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add vacation entry
	entry := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 8,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry)

	req := httptest.NewRequest("GET", "/api/vacation-hours?year=2024", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetVacationHours(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Code == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if usedHours, ok := result["used_hours"].(float64); ok {
			if int(usedHours) != 8 {
				t.Errorf("Expected 8 used hours, got %v", usedHours)
			}
		} else {
			t.Errorf("used_hours is not a number: %v", result["used_hours"])
		}
	}
}

func TestGetOverview(t *testing.T) {
	dbPath := setupHandlerTest(t)
	defer teardownHandlerTest(t, dbPath)

	// Add entries
	entry1 := db.TimesheetEntry{
		Date:           "2024-01-15",
		Client_name:    "Client A",
		Client_hours:   0,
		Vacation_hours: 0,
		Idle_hours:     0,
		Training_hours: 4,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	entry2 := db.TimesheetEntry{
		Date:           "2024-02-15",
		Client_name:    "Client B",
		Client_hours:   0,
		Vacation_hours: 8,
		Idle_hours:     0,
		Training_hours: 0,
		Sick_hours:     0,
		Holiday_hours:  0,
	}
	db.AddTimesheetEntry(entry1)
	db.AddTimesheetEntry(entry2)

	// Test without year (defaults to current year)
	req := httptest.NewRequest("GET", "/api/overview", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	GetOverview(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with year
	req = httptest.NewRequest("GET", "/api/overview?year=2024", nil)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	GetOverview(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Code == http.StatusOK {
		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if year, ok := result["year"].(float64); ok {
			if year != 2024 {
				t.Errorf("Expected year 2024, got %v", year)
			}
		} else {
			t.Errorf("year is not a number: %v", result["year"])
		}
	}
}

func TestExportPDF(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/export/pdf", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	ExportPDF(c)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}
}

func TestExportExcel(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/export/excel", nil)
	w := httptest.NewRecorder()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	ExportExcel(c)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("Expected status 501, got %d", w.Code)
	}
}
