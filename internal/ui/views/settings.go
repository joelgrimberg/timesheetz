package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsView represents the settings view
type SettingsView struct {
	ready bool
}

// NewSettingsView creates a new settings view
func NewSettingsView() *SettingsView {
	return &SettingsView{}
}

// Init initializes the view
func (v SettingsView) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the view
func (v SettingsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

// View renders the view
func (v SettingsView) View() string {
	if !v.ready {
		return "Initializing..."
	}
	return "Settings View"
} 