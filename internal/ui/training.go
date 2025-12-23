package ui

import (
	"fmt"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrainingKeyMap defines the keybindings for the training view
type TrainingKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Enter   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

// DefaultTrainingKeyMap returns the default keybindings
func DefaultTrainingKeyMap() TrainingKeyMap {
	return TrainingKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev year"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next year"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "go to timesheet"),
		),
		HelpKey: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("<"),
			key.WithHelp("<", "prev tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys(">"),
			key.WithHelp(">", "next tab"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k TrainingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Left,
		k.Right,
		k.Enter,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k TrainingKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
			k.Enter,
			k.HelpKey,
			k.Quit,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// TrainingModel represents the training hours view
type TrainingModel struct {
	table        table.Model
	yearlyTarget int
	currentYear  int
	keys         TrainingKeyMap
	help         help.Model
	showHelp     bool
}

// ChangeTrainingYearMsg is used to change the year
type ChangeTrainingYearMsg struct {
	Year int
}

// Command to change the year
func ChangeTrainingYear(year int) tea.Cmd {
	return func() tea.Msg {
		return ChangeTrainingYearMsg{Year: year}
	}
}

// NavigateToTimesheetMsg is sent when user wants to navigate to timesheet for a specific date
type NavigateToTimesheetMsg struct {
	Date string // YYYY-MM-DD format
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
			keys:         DefaultTrainingKeyMap(),
			help:         help.New(),
			showHelp:     false,
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
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("252"))
	t.SetStyles(s)

	// Get training entries for the current year
	dataLayer := datalayer.GetDataLayer()
	entries, err := dataLayer.GetTrainingEntriesForYear(currentYear)
	if err != nil {
		return TrainingModel{
			table:        t,
			yearlyTarget: configFile.TrainingHours.YearlyTarget,
			currentYear:  currentYear,
			keys:         DefaultTrainingKeyMap(),
			help:         help.New(),
			showHelp:     false,
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

	// Select the first row by default (if there are any entries)
	if len(entries) > 0 {
		t.SetCursor(0)
	} else {
		// If no entries, select the total row
		t.SetCursor(len(rows) - 1)
	}

	return TrainingModel{
		table:        t,
		yearlyTarget: configFile.TrainingHours.YearlyTarget,
		currentYear:  currentYear,
		keys:         DefaultTrainingKeyMap(),
		help:         help.New(),
		showHelp:     false,
	}
}

func (m TrainingModel) Init() tea.Cmd {
	return nil
}

func (m TrainingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeTrainingYearMsg:
		// Update the current year in the model
		m.currentYear = msg.Year

		// Get training entries for the new year
		entries, err := db.GetTrainingEntriesForYear(msg.Year)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
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
			fmt.Sprintf("%d/%d", totalHours, m.yearlyTarget),
		})

		m.table.SetRows(rows)

		// Select the first row by default (if there are any entries)
		if len(entries) > 0 {
			m.table.SetCursor(0)
		} else {
			// If no entries, select the total row
			m.table.SetCursor(len(rows) - 1)
		}

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			// Move to previous year
			return m, ChangeTrainingYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			return m, ChangeTrainingYear(m.currentYear + 1)
		case key.Matches(msg, m.keys.Enter):
			// Get the selected row
			cursorRow := m.table.Cursor()
			rows := m.table.Rows()

			// Don't navigate if on total row (last row)
			if cursorRow >= 0 && cursorRow < len(rows)-1 {
				selectedRow := rows[cursorRow]
				selectedDate := selectedRow[0] // First column is the date

				// Send navigation message
				return m, func() tea.Msg {
					return NavigateToTimesheetMsg{Date: selectedDate}
				}
			}
			return m, nil
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TrainingModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ↑/↓, k/j: Move up/down\n  ←/→, h/l: Change year\n  enter: Go to timesheet for selected date\n  ?: Toggle help\n  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("↑/↓: Navigate • ←/→: Change year • enter: Go to timesheet • ?: Help • q: Quit • </>: Tabs")
	}

	return fmt.Sprintf(
		"%s\n%s\n%s%s",
		titleStyle.Render(fmt.Sprintf("Training %d", m.currentYear)),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(m.table.View()),
		helpStyle.Render("↑/↓: Navigate • enter: Go to timesheet • <: Prev tab • >: Next tab • q: Quit"),
		helpView,
	)
}
