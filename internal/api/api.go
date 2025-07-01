package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"timesheet/internal/db"
)

// TrainingBudgetEntry represents a training budget entry
type TrainingBudgetEntry struct {
	Id               int     `json:"id"`
	Date             string  `json:"date"`
	Training_name    string  `json:"training_name"`
	Hours            int     `json:"hours"`
	Cost_without_vat float64 `json:"cost_without_vat"`
}

// GetTrainingBudgetEntries handles GET request for training budget entries
func GetTrainingBudgetEntries(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" {
		http.Error(w, "Year parameter is required", http.StatusBadRequest)
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		http.Error(w, "Invalid year parameter", http.StatusBadRequest)
		return
	}

	entries, err := db.GetTrainingBudgetEntriesForYear(yearInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching entries: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// AddTrainingBudgetEntry handles POST request to add a training budget entry
func AddTrainingBudgetEntry(w http.ResponseWriter, r *http.Request) {
	var entry db.TrainingBudgetEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := db.AddTrainingBudgetEntry(entry); err != nil {
		http.Error(w, fmt.Sprintf("Error adding entry: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdateTrainingBudgetEntry handles PUT request to update a training budget entry
func UpdateTrainingBudgetEntry(w http.ResponseWriter, r *http.Request) {
	var entry db.TrainingBudgetEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := db.UpdateTrainingBudgetEntry(entry); err != nil {
		http.Error(w, fmt.Sprintf("Error updating entry: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteTrainingBudgetEntry handles DELETE request to remove a training budget entry
func DeleteTrainingBudgetEntry(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTrainingBudgetEntry(idInt); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting entry: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetTrainingHours handles GET request for total training hours
func GetTrainingHours(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year")
	if year == "" {
		http.Error(w, "Year parameter is required", http.StatusBadRequest)
		return
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		http.Error(w, "Invalid year parameter", http.StatusBadRequest)
		return
	}

	entries, err := db.GetTrainingEntriesForYear(yearInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching entries: %v", err), http.StatusInternalServerError)
		return
	}

	var totalHours int
	for _, entry := range entries {
		totalHours += entry.Training_hours
	}

	response := struct {
		TotalHours int `json:"total_hours"`
	}{
		TotalHours: totalHours,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// ... existing routes ...

	// Training Budget routes
	http.HandleFunc("/api/training-budget", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetTrainingBudgetEntries(w, r)
		case http.MethodPost:
			AddTrainingBudgetEntry(w, r)
		case http.MethodPut:
			UpdateTrainingBudgetEntry(w, r)
		case http.MethodDelete:
			DeleteTrainingBudgetEntry(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Training Hours route
	http.HandleFunc("/api/training-hours", GetTrainingHours)

	// ... rest of main function ...
}
