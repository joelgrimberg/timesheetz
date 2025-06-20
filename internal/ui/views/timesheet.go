package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TimesheetView represents the timesheet view
type TimesheetView struct {
	ready bool
}

// NewTimesheetView creates a new timesheet view
func NewTimesheetView() *TimesheetView {
	return &TimesheetView{}
}

// Init initializes the view
func (v TimesheetView) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the view
func (v TimesheetView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

// View renders the view
func (v TimesheetView) View() string {
	if !v.ready {
		return "Initializing..."
	}
	return "Timesheet View"
} 