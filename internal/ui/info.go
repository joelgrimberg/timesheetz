package ui

import (
	"fmt"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InfoKeyMap defines the keybindings for the info view
type InfoKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
	Add     key.Binding
}

// DefaultInfoKeyMap returns the default keybindings
func DefaultInfoKeyMap() InfoKeyMap {
	return InfoKeyMap{
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
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add training budget entry"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k InfoKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Left,
		k.Right,
		k.HelpKey,
		k.Quit,
		k.Add,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k InfoKeyMap) FullHelp() [][]key.Binding {
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
			k.Add,
		},
	}
}

// InfoModel represents the combined info view (Training, Vacation, Training Budget)
type InfoModel struct {
	// Training table
	trainingTable        table.Model
	trainingYearlyTarget int
	trainingCurrentYear  int

	// Vacation table
	vacationTable        table.Model
	vacationYearlyTarget int
	vacationCurrentYear  int
	vacationEntries      []db.TimesheetEntry
	vacationTotalHours   int
	vacationRemaining    int

	// Training Budget table (only this one can be selected)
	trainingBudgetTable       table.Model
	trainingBudgetCurrentYear int

	// Common fields
	currentYear int
	keys        InfoKeyMap
	help        help.Model
	showHelp    bool
	ready       bool

	// Data loading tracking
	dataLoadedFlags map[string]bool
}

// ChangeInfoYearMsg is used to change the year
type ChangeInfoYearMsg struct {
	Year int
}

// Command to change the year
func ChangeInfoYear(year int) tea.Cmd {
	return func() tea.Msg {
		return ChangeInfoYearMsg{Year: year}
	}
}

// InitialInfoModel creates a new info model
func InitialInfoModel() InfoModel {
	// Get current year
	currentYear := time.Now().Year()

	// Get yearly targets from config
	configFile, err := config.GetConfig()
	if err != nil {
		// Default values if config is not available
		return InfoModel{
			trainingYearlyTarget:      36,
			vacationYearlyTarget:      0,
			trainingCurrentYear:       currentYear,
			vacationCurrentYear:       currentYear,
			trainingBudgetCurrentYear: currentYear,
			currentYear:               currentYear,
			keys:                      DefaultInfoKeyMap(),
			help:                      help.New(),
			showHelp:                  false,
			ready:                     false,
			dataLoadedFlags:           make(map[string]bool),
		}
	}

	// Create training table
	trainingColumns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Hours", Width: 8},
	}
	trainingTable := table.New(
		table.WithColumns(trainingColumns),
		table.WithFocused(false), // Not selectable
		table.WithHeight(8),
	)

	// Create vacation table
	vacationColumns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Hours", Width: 8},
	}
	vacationTable := table.New(
		table.WithColumns(vacationColumns),
		table.WithFocused(false), // Not selectable
		table.WithHeight(8),
	)

	// Create training budget table
	trainingBudgetColumns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Training", Width: 34},
		{Title: "Cost (€)", Width: 16},
	}
	trainingBudgetTable := table.New(
		table.WithColumns(trainingBudgetColumns),
		table.WithFocused(true), // Only this one is selectable
		table.WithHeight(8),
	)

	// Set styles for all tables
	tableStyles := table.DefaultStyles()
	tableStyles.Header = tableStyles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyles.Selected = tableStyles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tableStyles.Cell = tableStyles.Cell.
		Foreground(lipgloss.Color("252"))

	trainingTable.SetStyles(tableStyles)
	vacationTable.SetStyles(tableStyles)
	trainingBudgetTable.SetStyles(tableStyles)

	return InfoModel{
		trainingTable:             trainingTable,
		vacationTable:             vacationTable,
		trainingBudgetTable:       trainingBudgetTable,
		trainingYearlyTarget:      configFile.TrainingHours.YearlyTarget,
		vacationYearlyTarget:      configFile.VacationHours.YearlyTarget,
		trainingCurrentYear:       currentYear,
		vacationCurrentYear:       currentYear,
		trainingBudgetCurrentYear: currentYear,
		currentYear:               currentYear,
		keys:                      DefaultInfoKeyMap(),
		help:                      help.New(),
		showHelp:                  false,
		ready:                     false,
		dataLoadedFlags:           make(map[string]bool),
	}
}

func (m *InfoModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadTrainingData,
		m.loadVacationData,
		m.loadTrainingBudgetData,
	)
}

func (m *InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeInfoYearMsg:
		// Update all years
		m.trainingCurrentYear = msg.Year
		m.vacationCurrentYear = msg.Year
		m.trainingBudgetCurrentYear = msg.Year
		m.currentYear = msg.Year
		m.ready = false                           // Reset ready state while loading
		m.dataLoadedFlags = make(map[string]bool) // Reset data loaded flags

		return m, tea.Batch(
			m.loadTrainingData,
			m.loadVacationData,
			m.loadTrainingBudgetData,
		)

	case trainingDataLoadedMsg:
		// Training data loaded
		m.trainingTable.SetRows(msg.rows)
		m.dataLoadedFlags["training"] = true
		if m.checkAllDataLoaded() {
			m.ready = true
		}
		return m, nil
	case vacationDataLoadedMsg:
		// Vacation data loaded
		m.vacationTable.SetRows(msg.rows)
		m.vacationEntries = msg.entries
		m.vacationTotalHours = msg.totalHours
		m.vacationRemaining = msg.remaining
		m.dataLoadedFlags["vacation"] = true
		if m.checkAllDataLoaded() {
			m.ready = true
		}
		return m, nil
	case trainingBudgetDataLoadedMsg:
		// Training budget data loaded
		m.trainingBudgetTable.SetRows(msg.rows)

		// Select the first row by default (if there are any entries)
		if len(msg.entries) > 0 {
			m.trainingBudgetTable.SetCursor(0)
		} else {
			// If no entries, select the total row
			m.trainingBudgetTable.SetCursor(len(msg.rows) - 1)
		}

		m.dataLoadedFlags["trainingBudget"] = true
		if m.checkAllDataLoaded() {
			m.ready = true
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
			return m, ChangeInfoYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			return m, ChangeInfoYear(m.currentYear + 1)
		case key.Matches(msg, m.keys.Add):
			// Switch to training budget form mode
			return m, func() tea.Msg {
				return SwitchToTrainingBudgetFormMsg{}
			}
		}
	}

	// Only update the training budget table (the only selectable one)
	m.trainingBudgetTable, cmd = m.trainingBudgetTable.Update(msg)
	return m, cmd
}

func (m *InfoModel) View() string {
	if !m.ready {
		return "Loading info data..."
	}

	var s string

	// Show the year as title
	yearTitle := fmt.Sprintf("Info %d", m.currentYear)
	s += titleStyle.Render(yearTitle) + "\n\n"

	// Training section
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Training") + "\n"
	s += baseStyle.Render(m.trainingTable.View()) + "\n\n"

	// Vacation section
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Vacation") + "\n"
	s += baseStyle.Render(m.vacationTable.View()) + "\n\n"

	// Training Budget section
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Training Budget") + "\n"
	s += baseStyle.Render(m.trainingBudgetTable.View()) + "\n\n"

	// Help text
	if m.showHelp {
		s += m.help.FullHelpView(m.keys.FullHelp())
	} else {
		s += helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return s
}

// SwitchToTrainingBudgetFormMsg is sent when switching to training budget form
type SwitchToTrainingBudgetFormMsg struct{}

// checkAllDataLoaded checks if all data sources have been loaded
func (m *InfoModel) checkAllDataLoaded() bool {
	return m.dataLoadedFlags["training"] &&
		m.dataLoadedFlags["vacation"] &&
		m.dataLoadedFlags["trainingBudget"]
}

// loadTrainingData loads training data for the current year
func (m *InfoModel) loadTrainingData() tea.Msg {
	entries, err := db.GetTrainingEntriesForYear(m.trainingCurrentYear)
	if err != nil {
		// If database query fails, return empty data instead of error
		// This allows the InfoModel to become ready even if there are database issues
		return trainingDataLoadedMsg{rows: []table.Row{}}
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
		fmt.Sprintf("%d/%d", totalHours, m.trainingYearlyTarget),
	})

	return trainingDataLoadedMsg{rows: rows}
}

// loadVacationData loads vacation data for the current year from timesheet table only
func (m *InfoModel) loadVacationData() tea.Msg {
	entries, err := db.GetVacationEntriesForYear(m.vacationCurrentYear)
	if err != nil {
		// If database query fails, return empty data instead of error
		// This allows the InfoModel to become ready even if there are database issues
		return vacationDataLoadedMsg{
			rows:       []table.Row{},
			entries:    nil,
			totalHours: 0,
			remaining:  0,
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	var totalHours int
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			fmt.Sprintf("%d", entry.Vacation_hours),
		})
		totalHours += entry.Vacation_hours
	}

	// Add total row
	rows = append(rows, table.Row{
		"Total",
		fmt.Sprintf("%d/%d", totalHours, m.vacationYearlyTarget),
	})

	return vacationDataLoadedMsg{
		rows:       rows,
		entries:    nil,
		totalHours: totalHours,
		remaining:  m.vacationYearlyTarget - totalHours,
	}
}

// loadTrainingBudgetData loads training budget data for the current year
func (m *InfoModel) loadTrainingBudgetData() tea.Msg {
	entries, err := db.GetTrainingBudgetEntriesForYear(m.trainingBudgetCurrentYear)
	if err != nil {
		// If database query fails, return empty data instead of error
		// This allows the InfoModel to become ready even if there are database issues
		return trainingBudgetDataLoadedMsg{
			rows:    []table.Row{},
			entries: []db.TrainingBudgetEntry{},
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	var totalCost float64
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			entry.Training_name,
			fmt.Sprintf("%.2f", entry.Cost_without_vat),
		})
		totalCost += entry.Cost_without_vat
	}

	// Add total row
	rows = append(rows, table.Row{
		"Total",
		"",
		fmt.Sprintf("%.2f", totalCost),
	})

	return trainingBudgetDataLoadedMsg{
		rows:    rows,
		entries: entries,
	}
}

// Messages for data loading
type trainingDataLoadedMsg struct {
	rows []table.Row
}
type vacationDataLoadedMsg struct {
	rows       []table.Row
	entries    []db.TimesheetEntry
	totalHours int
	remaining  int
}
type trainingBudgetDataLoadedMsg struct {
	rows    []table.Row
	entries []db.TrainingBudgetEntry
}
