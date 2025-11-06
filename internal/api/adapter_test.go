package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"timesheet/internal/db"
)

func TestNewClientAdapter(t *testing.T) {
	client := NewClient("http://localhost:8080")
	adapter := NewClientAdapter(client)
	if adapter == nil {
		t.Fatal("NewClientAdapter returned nil")
	}
	if adapter.client != client {
		t.Error("Adapter client mismatch")
	}
}

func TestClientAdapter_AllMethods(t *testing.T) {
	entries := []db.TimesheetEntry{
		{Id: 1, Date: "2024-01-15", Client_name: "Client A", Client_hours: 8},
	}
	trainingEntries := []db.TrainingBudgetEntry{
		{Id: 1, Date: "2024-01-15", Training_name: "Training A", Hours: 8, Cost_without_vat: 100.0},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/timesheet":
			json.NewEncoder(w).Encode(entries)
		case "/api/last-client":
			json.NewEncoder(w).Encode(map[string]string{"client_name": "Client A"})
		case "/api/training-budget":
			json.NewEncoder(w).Encode(trainingEntries)
		case "/health":
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	adapter := NewClientAdapter(client)

	// Test all adapter methods delegate to client
	_, err := adapter.GetAllTimesheetEntries(0, 0)
	if err != nil {
		t.Errorf("GetAllTimesheetEntries failed: %v", err)
	}

	_, err = adapter.GetTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Errorf("GetTimesheetEntryByDate failed: %v", err)
	}

	err = adapter.AddTimesheetEntry(db.TimesheetEntry{Date: "2024-01-16", Client_name: "Client B"})
	if err != nil {
		t.Errorf("AddTimesheetEntry failed: %v", err)
	}

	err = adapter.UpdateTimesheetEntry(db.TimesheetEntry{Id: 1, Date: "2024-01-15"})
	if err != nil {
		t.Errorf("UpdateTimesheetEntry failed: %v", err)
	}

	err = adapter.DeleteTimesheetEntryByDate("2024-01-15")
	if err != nil {
		t.Errorf("DeleteTimesheetEntryByDate failed: %v", err)
	}

	err = adapter.DeleteTimesheetEntry("1")
	if err != nil {
		t.Errorf("DeleteTimesheetEntry failed: %v", err)
	}

	_, err = adapter.GetLastClientName()
	if err != nil {
		t.Errorf("GetLastClientName failed: %v", err)
	}

	_, err = adapter.GetTrainingEntriesForYear(2024)
	if err != nil {
		t.Errorf("GetTrainingEntriesForYear failed: %v", err)
	}

	_, err = adapter.GetVacationEntriesForYear(2024)
	if err != nil {
		t.Errorf("GetVacationEntriesForYear failed: %v", err)
	}

	_, err = adapter.GetVacationHoursForYear(2024)
	if err != nil {
		t.Errorf("GetVacationHoursForYear failed: %v", err)
	}

	_, err = adapter.GetTrainingBudgetEntriesForYear(2024)
	if err != nil {
		t.Errorf("GetTrainingBudgetEntriesForYear failed: %v", err)
	}

	err = adapter.AddTrainingBudgetEntry(db.TrainingBudgetEntry{Date: "2024-01-16", Training_name: "Training B"})
	if err != nil {
		t.Errorf("AddTrainingBudgetEntry failed: %v", err)
	}

	err = adapter.UpdateTrainingBudgetEntry(db.TrainingBudgetEntry{Id: 1, Date: "2024-01-15"})
	if err != nil {
		t.Errorf("UpdateTrainingBudgetEntry failed: %v", err)
	}

	err = adapter.DeleteTrainingBudgetEntry(1)
	if err != nil {
		t.Errorf("DeleteTrainingBudgetEntry failed: %v", err)
	}

	_, err = adapter.GetTrainingBudgetEntry(1)
	if err != nil {
		t.Errorf("GetTrainingBudgetEntry failed: %v", err)
	}

	_, err = adapter.GetTrainingBudgetEntryByDate("2024-01-15")
	if err != nil {
		t.Errorf("GetTrainingBudgetEntryByDate failed: %v", err)
	}

	err = adapter.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

