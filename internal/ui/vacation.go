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

// VacationKeyMap defines the keybindings for the vacation view
type VacationKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

// DefaultVacationKeyMap returns the default keybindings
func DefaultVacationKeyMap() VacationKeyMap {
	return VacationKeyMap{
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
func (k VacationKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Left,
		k.Right,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k VacationKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
			k.HelpKey,
			k.Quit,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// VacationModel represents the vacation hours view
type VacationModel struct {
	table        table.Model
	yearlyTarget int
	currentYear  int
	summary      db.VacationSummary
	keys         VacationKeyMap
	help         help.Model
	showHelp     bool
}

// ChangeVacationYearMsg is used to change the year
type ChangeVacationYearMsg struct {
	Year int
}

// Command to change the year
func ChangeVacationYear(year int) tea.Cmd {
	return func() tea.Msg {
		return ChangeVacationYearMsg{Year: year}
	}
}

// InitialVacationModel creates a new vacation model
func InitialVacationModel() VacationModel {
	// Get current year
	currentYear := time.Now().Year()

	// Get yearly target from config
	configFile, err := config.GetConfig()
	if err != nil {
		// Default to 25 if config is not available
		return VacationModel{
			yearlyTarget: 25,
			currentYear:  currentYear,
			keys:         DefaultVacationKeyMap(),
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

	// Get vacation entries and summary for the current year
	dataLayer := datalayer.GetDataLayer()

	// Get comprehensive vacation summary (includes carryover)
	summary, err := dataLayer.GetVacationSummaryForYear(currentYear)
	if err != nil {
		return VacationModel{
			table:        t,
			yearlyTarget: configFile.VacationHours.YearlyTarget,
			currentYear:  currentYear,
			summary:      db.VacationSummary{},
			keys:         DefaultVacationKeyMap(),
			help:         help.New(),
			showHelp:     false,
		}
	}

	entries, err := dataLayer.GetVacationEntriesForYear(currentYear)
	if err != nil {
		return VacationModel{
			table:        t,
			yearlyTarget: configFile.VacationHours.YearlyTarget,
			currentYear:  currentYear,
			summary:      summary,
			keys:         DefaultVacationKeyMap(),
			help:         help.New(),
			showHelp:     false,
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			fmt.Sprintf("%d", entry.Vacation_hours),
		})
	}

	// Add total row showing used hours and total available
	rows = append(rows, table.Row{
		"Total",
		fmt.Sprintf("%d/%d", summary.UsedHours, summary.TotalAvailable),
	})

	t.SetRows(rows)

	// Select the first row by default (if there are any entries)
	// Never select the total row
	if len(entries) > 0 {
		t.SetCursor(0)
	} else {
		// If no entries, don't select anything (cursor will be at -1)
		t.SetCursor(-1)
	}

	return VacationModel{
		table:        t,
		yearlyTarget: configFile.VacationHours.YearlyTarget,
		currentYear:  currentYear,
		summary:      summary,
		keys:         DefaultVacationKeyMap(),
		help:         help.New(),
		showHelp:     false,
	}
}

func (m VacationModel) Init() tea.Cmd {
	return nil
}

// getLastSelectableRowIndex returns the index of the last row that can be selected
// (excludes the total row)
func (m VacationModel) getLastSelectableRowIndex() int {
	rows := m.table.Rows()
	if len(rows) <= 1 {
		return -1 // No selectable rows
	}
	return len(rows) - 2 // Last row before the total row
}

func (m VacationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeVacationYearMsg:
		// Update the current year in the model
		m.currentYear = msg.Year

		// Reload config to get the latest yearly target
		configFile, err := config.GetConfig()
		if err == nil {
			m.yearlyTarget = configFile.VacationHours.YearlyTarget
		}

		// Get vacation summary for the new year (includes carryover)
		dataLayer := datalayer.GetDataLayer()
		summary, err := dataLayer.GetVacationSummaryForYear(msg.Year)
		if err == nil {
			m.summary = summary
		}

		// Get vacation entries for the new year
		entries, err := dataLayer.GetVacationEntriesForYear(msg.Year)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
		}

		// Convert entries to table rows
		var rows []table.Row
		for _, entry := range entries {
			rows = append(rows, table.Row{
				entry.Date,
				fmt.Sprintf("%d", entry.Vacation_hours),
			})
		}

		// Add total row showing used hours and total available
		rows = append(rows, table.Row{
			"Total",
			fmt.Sprintf("%d/%d", m.summary.UsedHours, m.summary.TotalAvailable),
		})

		m.table.SetRows(rows)

		// Select the first row by default (if there are any entries)
		// Never select the total row
		if len(entries) > 0 {
			m.table.SetCursor(0)
		} else {
			// If no entries, don't select anything (cursor will be at -1)
			m.table.SetCursor(-1)
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
			return m, ChangeVacationYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			return m, ChangeVacationYear(m.currentYear + 1)
		case key.Matches(msg, m.keys.Down):
			// Handle down navigation to prevent landing on total row
			currentCursor := m.table.Cursor()
			lastSelectable := m.getLastSelectableRowIndex()

			if currentCursor < lastSelectable {
				// Normal navigation - let the table handle it
				m.table, cmd = m.table.Update(msg)
			} else if currentCursor == lastSelectable {
				// We're at the last selectable row, don't move down
				return m, nil
			} else {
				// We're somehow on the total row, move to last selectable
				m.table.SetCursor(lastSelectable)
			}
			return m, cmd
		case key.Matches(msg, m.keys.Up):
			// Handle up navigation normally, but ensure we don't end up on total row
			m.table, cmd = m.table.Update(msg)
			// Check if we landed on the total row and move up if needed
			if m.table.Cursor() == m.getLastSelectableRowIndex()+1 {
				m.table.SetCursor(m.getLastSelectableRowIndex())
			}
			return m, cmd
		}
	}

	// Handle other table updates
	m.table, cmd = m.table.Update(msg)

	// Ensure cursor never lands on the total row
	if m.table.Cursor() == m.getLastSelectableRowIndex()+1 {
		m.table.SetCursor(m.getLastSelectableRowIndex())
	}

	return m, cmd
}

func (m VacationModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ↑/↓, k/j: Move up/down\n  ←/→, h/l: Change year\n  ?: Toggle help\n  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("↑/↓: Navigate • ←/→: Change year • ?: Help • q: Quit • </>: Tabs")
	}

	// Create summary section showing carryover breakdown
	summaryContent := ""
	if m.summary.CarryoverHours > 0 {
		summaryContent = fmt.Sprintf(
			"%s\n  %s\n  %s\n\n%s\n  %s\n  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Available:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("Current Year (%d): %d hours", m.currentYear, m.summary.YearlyTarget)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("Carryover from %d: %d hours", m.summary.Year-1, m.summary.CarryoverHours)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Used:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("From Carryover: %d hours", m.summary.UsedFromCarryover)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("From Current Year: %d hours", m.summary.UsedFromCurrent)),
		)
	} else {
		summaryContent = fmt.Sprintf(
			"%s\n  %s\n\n%s\n  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Available:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("Current Year (%d): %d hours", m.currentYear, m.summary.YearlyTarget)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Used:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(fmt.Sprintf("Total: %d hours", m.summary.UsedHours)),
		)
	}

	summaryBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Render(summaryContent)

	return fmt.Sprintf(
		"%s\n%s\n\n%s\n\n%s%s",
		titleStyle.Render(fmt.Sprintf("Vacation %d", m.currentYear)),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(m.table.View()),
		summaryBox,
		helpStyle.Render("↑/↓: Navigate • <: Prev tab • >: Next tab • q: Quit"),
		helpView,
	)
}
