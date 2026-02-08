package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppState represents persisted application state
type AppState struct {
	ActiveTab string `json:"activeTab"`
}

// getStatePath returns the path to the state file
func getStatePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "timesheetz", "state.json")
}

// LoadAppState loads the persisted app state from disk
func LoadAppState() AppState {
	state := AppState{
		ActiveTab: "timesheet", // default
	}

	statePath := getStatePath()
	if statePath == "" {
		return state
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return state
	}

	json.Unmarshal(data, &state)
	return state
}

// SaveAppState saves the app state to disk
func SaveAppState(state AppState) error {
	statePath := getStatePath()
	if statePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// AppModeToString converts AppMode to a string for persistence
func AppModeToString(mode AppMode) string {
	switch mode {
	case TimesheetMode:
		return "timesheet"
	case OverviewMode:
		return "overview"
	case TrainingMode:
		return "training"
	case TrainingBudgetMode:
		return "training_budget"
	case VacationMode:
		return "vacation"
	case ClientsMode:
		return "clients"
	case EarningsMode:
		return "earnings"
	case ConfigMode:
		return "config"
	default:
		return "timesheet"
	}
}

// StringToAppMode converts a string back to AppMode
func StringToAppMode(s string) AppMode {
	switch s {
	case "timesheet":
		return TimesheetMode
	case "overview":
		return OverviewMode
	case "training":
		return TrainingMode
	case "training_budget":
		return TrainingBudgetMode
	case "vacation":
		return VacationMode
	case "clients":
		return ClientsMode
	case "earnings":
		return EarningsMode
	case "config":
		return ConfigMode
	default:
		return TimesheetMode
	}
}
