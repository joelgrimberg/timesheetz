package ui

import (
	"fmt"
	"time"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrainingBudgetModel represents the training budget view
type TrainingBudgetModel struct {
	table       table.Model
	currentYear int
}

// RefreshTrainingBudgetMsg is sent when the training budget should be refreshed
type RefreshTrainingBudgetMsg struct{}

// RefreshTrainingBudgetCmd returns a command that refreshes the training budget
func RefreshTrainingBudgetCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshTrainingBudgetMsg{}
	}
}

// InitialTrainingBudgetModel creates a new training budget model
func InitialTrainingBudgetModel() TrainingBudgetModel {
	// Get current year
	currentYear := time.Now().Year()

	// Create columns for the table
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Training", Width: 30},
		{Title: "Hours", Width: 8},
		{Title: "Cost (€)", Width: 12},
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

	// Get training budget entries for the current year
	entries, err := db.GetTrainingBudgetEntriesForYear(currentYear)
	if err != nil {
		return TrainingBudgetModel{
			table:       t,
			currentYear: currentYear,
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	var totalHours int
	var totalCost float64
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			entry.Training_name,
			fmt.Sprintf("%d", entry.Hours),
			fmt.Sprintf("%.2f", entry.Cost_without_vat),
		})
		totalHours += entry.Hours
		totalCost += entry.Cost_without_vat
	}

	// Add total row
	rows = append(rows, table.Row{
		"Total",
		"",
		fmt.Sprintf("%d", totalHours),
		fmt.Sprintf("%.2f", totalCost),
	})

	t.SetRows(rows)

	return TrainingBudgetModel{
		table:       t,
		currentYear: currentYear,
	}
}

func (m TrainingBudgetModel) Init() tea.Cmd {
	return RefreshTrainingBudgetCmd()
}

func (m TrainingBudgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "r":
			// Add refresh on 'r' key
			return m, RefreshTrainingBudgetCmd()
		case "c":
			// Delete the selected entry
			row := m.table.SelectedRow()
			if len(row) > 0 {
				// The first column is the date, but we need the ID
				// We'll need to fetch the ID from the database based on the row data
				// For now, let's assume the table includes the ID as a hidden field (not shown)
				// If not, you may need to adjust the table to include the ID
				// For now, let's fetch by date, training name, and cost
				date := row[0]
				trainingName := row[1]
				cost := row[3]
				// Find the entry in the DB
				entries, err := db.GetTrainingBudgetEntriesForYear(m.currentYear)
				if err == nil {
					for _, entry := range entries {
						if entry.Date == date && entry.Training_name == trainingName && fmt.Sprintf("%.2f", entry.Cost_without_vat) == cost {
							_ = db.DeleteTrainingBudgetEntry(entry.Id)
							break
						}
					}
				}
				// Refresh the model
				return InitialTrainingBudgetModel(), nil
			}
		}
	case RefreshTrainingBudgetMsg:
		// Reinitialize the model to refresh data
		newModel := InitialTrainingBudgetModel()
		return newModel, nil
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TrainingBudgetModel) View() string {
	return fmt.Sprintf(
		"\nTraining Budget Overview for %d\n\n%s\n\n%s",
		m.currentYear,
		m.table.View(),
		helpStyle.Render("↑/↓: Navigate • <: Prev tab • >: Next tab • $: Add entry • r: Refresh • q: Quit"),
	)
}
