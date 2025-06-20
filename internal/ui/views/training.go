package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

// TrainingView represents the training view
type TrainingView struct {
	ready bool
}

// NewTrainingView creates a new training view
func NewTrainingView() *TrainingView {
	return &TrainingView{}
}

// Init initializes the view
func (v TrainingView) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the view
func (v TrainingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v, nil
}

// View renders the view
func (v TrainingView) View() string {
	if !v.ready {
		return "Initializing..."
	}
	return "Training View"
} 