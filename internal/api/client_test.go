package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"timesheet/internal/db"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("Expected baseURL 'http://localhost:8080', got %s", client.baseURL)
	}
	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", client.httpClient.Timeout)
	}
}

func TestClient_makeRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           interface{}
		handler        http.HandlerFunc
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "GET success",
			method:   "GET",
			endpoint: "/test",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "POST with body",
			method:   "POST",
			endpoint: "/test",
			body:     map[string]string{"key": "value"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id":1}`))
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:     "Error status code",
			method:   "GET",
			endpoint: "/test",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"not found"}`))
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClient(server.URL)
			result, err := client.makeRequest(tt.method, tt.endpoint, tt.body)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestClient_GetAllTimesheetEntries(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Client_name: "Client A", Client_hours: 8},
		{Id: 2, Date: "2024-02-15", Client_name: "Client B", Client_hours: 6},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/timesheet" {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetAllTimesheetEntries(0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	// Test with year/month filter
	result, err = client.GetAllTimesheetEntries(2024, time.January)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 entry for Jan 2024, got %d", len(result))
	}
	if result[0].Date != "2024-01-15" {
		t.Errorf("Expected date 2024-01-15, got %s", result[0].Date)
	}
}

func TestClient_GetTimesheetEntryByDate(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Client_name: "Client A"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry, err := client.GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if entry.Date != "2024-01-15" {
		t.Errorf("Expected date 2024-01-15, got %s", entry.Date)
	}

	// Test not found
	_, err = client.GetTimesheetEntryByDate("2024-01-16")
	if err == nil {
		t.Error("Expected error for non-existent date")
	}
}

func TestClient_AddTimesheetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		var entry db.TimesheetEntry
		json.NewDecoder(r.Body).Decode(&entry)
		entry.Id = 1
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(entry)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry := db.TimesheetEntry{Date: "2024-01-15", Client_name: "Client A", Client_hours: 8}
	err := client.AddTimesheetEntry(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_UpdateTimesheetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(db.TimesheetEntry{Id: 1})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry := db.TimesheetEntry{Id: 1, Date: "2024-01-15", Client_name: "Client A"}
	err := client.UpdateTimesheetEntry(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test with zero ID
	entry.Id = 0
	err = client.UpdateTimesheetEntry(entry)
	if err == nil {
		t.Error("Expected error for zero ID")
	}
}

func TestClient_DeleteTimesheetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteTimesheetEntry("1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_DeleteTimesheetEntryByDate(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(entries)
		} else if r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_GetLastClientName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"client_name": "Client A"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	name, err := client.GetLastClientName()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if name != "Client A" {
		t.Errorf("Expected 'Client A', got %s", name)
	}
}

func TestClient_GetTrainingEntriesForYear(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Training_hours: 4},
		{Id: 2, Date: "2024-02-15", Training_hours: 0},
		{Id: 3, Date: "2023-01-15", Training_hours: 2},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetTrainingEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 training entry for 2024, got %d", len(result))
	}
}

func TestClient_GetVacationEntriesForYear(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Vacation_hours: 8},
		{Id: 2, Date: "2024-02-15", Vacation_hours: 0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetVacationEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 vacation entry for 2024, got %d", len(result))
	}
}

func TestClient_GetVacationHoursForYear(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Vacation_hours: 8},
		{Id: 2, Date: "2024-02-15", Vacation_hours: 4},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	total, err := client.GetVacationHoursForYear(2024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if total != 12 {
		t.Errorf("Expected 12 hours, got %d", total)
	}
}

func TestClient_GetTrainingBudgetEntriesForYear(t *testing.T) {
	entries := []db.TrainingBudgetEntry{
		{Id: 1, Date: "2024-01-15", Training_name: "Training A", Hours: 8, Cost_without_vat: 100.0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("year") != "2024" {
			t.Errorf("Expected year=2024, got %s", r.URL.Query().Get("year"))
		}
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetTrainingBudgetEntriesForYear(2024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(result))
	}
}

func TestClient_AddTrainingBudgetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(db.TrainingBudgetEntry{Id: 1})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry := db.TrainingBudgetEntry{Date: "2024-01-15", Training_name: "Training A", Hours: 8, Cost_without_vat: 100.0}
	err := client.AddTrainingBudgetEntry(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_UpdateTrainingBudgetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry := db.TrainingBudgetEntry{Id: 1, Date: "2024-01-15", Training_name: "Training A", Hours: 8, Cost_without_vat: 100.0}
	err := client.UpdateTrainingBudgetEntry(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_DeleteTrainingBudgetEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "1" {
			t.Errorf("Expected id=1, got %s", r.URL.Query().Get("id"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.DeleteTrainingBudgetEntry(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestClient_GetTrainingBudgetEntry(t *testing.T) {
	entries := []db.TrainingBudgetEntry{
		{Id: 1, Date: "2024-01-15", Training_name: "Training A"},
		{Id: 2, Date: "2024-02-15", Training_name: "Training B"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry, err := client.GetTrainingBudgetEntry(1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if entry.Id != 1 {
		t.Errorf("Expected ID 1, got %d", entry.Id)
	}

	// Test not found
	_, err = client.GetTrainingBudgetEntry(999)
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}
}

func TestClient_GetTrainingBudgetEntryByDate(t *testing.T) {
	entries := []db.TrainingBudgetEntry{
		{Id: 1, Date: "2024-01-15", Training_name: "Training A"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	entry, err := client.GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if entry.Date != "2024-01-15" {
		t.Errorf("Expected date 2024-01-15, got %s", entry.Date)
	}

	// Test invalid date format
	_, err = client.GetTrainingBudgetEntryByDate("24")
	if err == nil {
		t.Error("Expected error for invalid date format")
	}
}

func TestClient_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("Expected /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Ping()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetClient(t *testing.T) {
	// Test local mode
	os.Setenv("TIMESHEETZ_API_MODE", "local")
	defer os.Unsetenv("TIMESHEETZ_API_MODE")

	client, err := GetClient()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client != nil {
		t.Error("Expected nil client for local mode")
	}

	// Test remote mode with valid server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer server.Close()

	os.Setenv("TIMESHEETZ_API_MODE", "remote")
	os.Setenv("TIMESHEETZ_API_URL", server.URL)
	defer func() {
		os.Unsetenv("TIMESHEETZ_API_MODE")
		os.Unsetenv("TIMESHEETZ_API_URL")
	}()

	client, err = GetClient()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Error("Expected client for remote mode")
	}

	// Test remote mode with missing URL
	os.Setenv("TIMESHEETZ_API_MODE", "remote")
	os.Unsetenv("TIMESHEETZ_API_URL")
	client, err = GetClient()
	if err == nil {
		t.Error("Expected error for missing API URL")
	}
}
