package ui

import (
	"fmt"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrainingModel represents the training hours view
type TrainingModel struct {
	table        table.Model
	yearlyTarget int
	currentYear  int
}

// InitialTrainingModel creates a new training model
func InitialTrainingModel() TrainingModel {
	// Get current year
	currentYear := time.Now().Year()

	// Get yearly target from config
	configFile, err := config.GetConfig()
	if err != nil {
		// Default to 36 if config is not available
		return TrainingModel{
			yearlyTarget: 36,
			currentYear:  currentYear,
		}
	}

	// Create columns for the table
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Hours", Width: 8},
	}

	// Create the table
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Set styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Get training entries for the current year
	entries, err := db.GetTrainingEntriesForYear(currentYear)
	if err != nil {
		return TrainingModel{
			table:        t,
			yearlyTarget: configFile.TrainingHours.YearlyTarget,
			currentYear:  currentYear,
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	var totalHours int
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			fmt.Sprintf("%d", entry.Training_hours),
		})
		totalHours += entry.Training_hours
	}

	// Add total row
	rows = append(rows, table.Row{
		"Total",
		fmt.Sprintf("%d/%d", totalHours, configFile.TrainingHours.YearlyTarget),
	})

	t.SetRows(rows)

	return TrainingModel{
		table:        t,
		yearlyTarget: configFile.TrainingHours.YearlyTarget,
		currentYear:  currentYear,
	}
}

func (m TrainingModel) Init() tea.Cmd {
	return nil
}

func (m TrainingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TrainingModel) View() string {
	return fmt.Sprintf(
		"\nTraining Hours Overview for %d\n\n%s\n\n%s",
		m.currentYear,
		m.table.View(),
		helpStyle.Render("↑/↓: Navigate • <: Prev tab • >: Next tab • q: Quit"),
	)
}
