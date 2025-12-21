package ui

import (
	"fmt"
	"time"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"
	"timesheet/internal/utils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EarningsKeyMap defines the keybindings for the earnings view
type EarningsKeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	HelpKey       key.Binding
	Quit          key.Binding
	Refresh       key.Binding
	ToggleView    key.Binding
	ToggleSummary key.Binding
	MonthUp       key.Binding
	MonthDown     key.Binding
	PrevTab       key.Binding
	NextTab       key.Binding
}

// DefaultEarningsKeyMap returns the default keybindings
func DefaultEarningsKeyMap() EarningsKeyMap {
	return EarningsKeyMap{
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
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "toggle monthly/yearly"),
		),
		ToggleSummary: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle summary"),
		),
		MonthUp: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "prev month"),
		),
		MonthDown: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "next month"),
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
func (k EarningsKeyMap) ShortHelp() []key.Binding {
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
func (k EarningsKeyMap) FullHelp() [][]key.Binding {
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
			k.Refresh,
			k.ToggleView,
			k.ToggleSummary,
			k.MonthUp,
			k.MonthDown,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// EarningsModel represents the earnings overview view
type EarningsModel struct {
	table        table.Model
	currentYear  int
	currentMonth int // 0 for yearly view, 1-12 for monthly
	monthlyView  bool
	summaryMode  bool // true = summary grouped by client/rate, false = detailed by date
	keys         EarningsKeyMap
	help         help.Model
	showHelp     bool
}

// RefreshEarningsMsg is sent when the earnings should be refreshed
type RefreshEarningsMsg struct{}

// RefreshEarningsCmd returns a command that refreshes the earnings
func RefreshEarningsCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshEarningsMsg{}
	}
}

// InitialEarningsModel creates a new earnings model
func InitialEarningsModel() EarningsModel {
	// Get current year and month
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Create columns for the table (start with summary mode)
	columns := []table.Column{
		{Title: "Client", Width: 30},
		{Title: "Rate", Width: 14},
		{Title: "Hours", Width: 10},
		{Title: "Earnings", Width: 16},
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
	s.Cell = s.Cell

	// Set padding for all cells
	s.Header = s.Header.PaddingLeft(0).PaddingRight(0)
	s.Selected = s.Selected.PaddingLeft(0).PaddingRight(0)
	s.Cell = s.Cell.PaddingLeft(0).PaddingRight(0)

	t.SetStyles(s)

	model := EarningsModel{
		table:        t,
		currentYear:  currentYear,
		currentMonth: currentMonth,
		monthlyView:  false, // Start with yearly view
		summaryMode:  true,  // Start with summary view (grouped by client/rate)
		keys:         DefaultEarningsKeyMap(),
		help:         help.New(),
		showHelp:     false,
	}

	// Load initial data
	model.loadEarnings()

	return model
}

func (m *EarningsModel) loadEarnings() {
	dataLayer := datalayer.GetDataLayer()
	var overview db.EarningsOverview
	var err error

	if m.monthlyView {
		overview, err = dataLayer.CalculateEarningsForMonth(m.currentYear, m.currentMonth)
	} else {
		// Yearly view - use summary or detailed based on summaryMode
		if m.summaryMode {
			overview, err = dataLayer.CalculateEarningsSummaryForYear(m.currentYear)
		} else {
			overview, err = dataLayer.CalculateEarningsForYear(m.currentYear)
		}
	}

	if err != nil {
		m.table.SetRows([]table.Row{})
		return
	}

	// Convert entries to table rows
	var rows []table.Row
	for _, entry := range overview.Entries {
		if m.summaryMode && !m.monthlyView {
			// Summary mode: no date column
			rows = append(rows, table.Row{
				entry.ClientName,
				utils.FormatEuro(entry.HourlyRate),
				fmt.Sprintf("%d", entry.ClientHours),
				utils.FormatEuro(entry.Earnings),
			})
		} else {
			// Detailed mode: include date
			rows = append(rows, table.Row{
				entry.Date,
				entry.ClientName,
				fmt.Sprintf("%d", entry.ClientHours),
				utils.FormatEuro(entry.HourlyRate),
				utils.FormatEuro(entry.Earnings),
			})
		}
	}

	// Add total row
	if m.summaryMode && !m.monthlyView {
		rows = append(rows, table.Row{
			"TOTAL",
			"",
			fmt.Sprintf("%d", overview.TotalHours),
			utils.FormatEuro(overview.TotalEarnings),
		})
	} else {
		rows = append(rows, table.Row{
			"TOTAL",
			"",
			fmt.Sprintf("%d", overview.TotalHours),
			"",
			utils.FormatEuro(overview.TotalEarnings),
		})
	}

	m.table.SetRows(rows)

	// Select the first row by default
	if len(overview.Entries) > 0 {
		m.table.SetCursor(0)
	} else {
		// If no entries, select the total row
		m.table.SetCursor(len(rows) - 1)
	}
}

func (m EarningsModel) Init() tea.Cmd {
	return RefreshEarningsCmd()
}

func (m EarningsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case RefreshEarningsMsg:
		m.loadEarnings()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.loadEarnings()
			return m, nil
		case key.Matches(msg, m.keys.ToggleView):
			m.monthlyView = !m.monthlyView
			m.loadEarnings()
			return m, nil
		case key.Matches(msg, m.keys.ToggleSummary):
			// Only toggle summary in yearly view
			if !m.monthlyView {
				m.summaryMode = !m.summaryMode
				// Clear rows before changing columns to avoid index out of range
				m.table.SetRows([]table.Row{})
				// Update table columns based on mode
				if m.summaryMode {
					m.table.SetColumns([]table.Column{
						{Title: "Client", Width: 30},
						{Title: "Rate", Width: 14},
						{Title: "Hours", Width: 10},
						{Title: "Earnings", Width: 16},
					})
				} else {
					m.table.SetColumns([]table.Column{
						{Title: "Date", Width: 12},
						{Title: "Client", Width: 25},
						{Title: "Hours", Width: 8},
						{Title: "Rate", Width: 14},
						{Title: "Earnings", Width: 14},
					})
				}
				m.loadEarnings()
			}
			return m, nil
		case key.Matches(msg, m.keys.Left):
			// Move to previous year
			m.currentYear--
			m.loadEarnings()
			return m, nil
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			m.currentYear++
			m.loadEarnings()
			return m, nil
		case key.Matches(msg, m.keys.MonthUp):
			// Only in monthly view
			if m.monthlyView {
				m.currentMonth--
				if m.currentMonth < 1 {
					m.currentMonth = 12
					m.currentYear--
				}
				m.loadEarnings()
			}
			return m, nil
		case key.Matches(msg, m.keys.MonthDown):
			// Only in monthly view
			if m.monthlyView {
				m.currentMonth++
				if m.currentMonth > 12 {
					m.currentMonth = 1
					m.currentYear++
				}
				m.loadEarnings()
			}
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if len(m.table.Rows()) > 0 {
				if m.table.Cursor() == 0 {
					// If at first row, go to last data row (excluding total)
					m.table.SetCursor(len(m.table.Rows()) - 2)
					return m, nil
				}
				m.table.MoveUp(0)
			}
		case key.Matches(msg, m.keys.Down):
			if len(m.table.Rows()) > 0 {
				if m.table.Cursor() == len(m.table.Rows())-2 { // If at last data row
					// Go to first row
					m.table.SetCursor(0)
					return m, nil
				}
				m.table.MoveDown(0)
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m EarningsModel) View() string {
	var s string

	// Title
	var title string
	if m.monthlyView {
		monthName := time.Month(m.currentMonth).String()
		title = fmt.Sprintf("Earnings Overview - %s %d", monthName, m.currentYear)
	} else {
		if m.summaryMode {
			title = fmt.Sprintf("Earnings Overview - %d (Summary)", m.currentYear)
		} else {
			title = fmt.Sprintf("Earnings Overview - %d (Detailed)", m.currentYear)
		}
	}
	s += titleStyle.Render(title) + "\n"

	// Table view
	tableView := m.table.View()
	s += baseStyle.Render(tableView) + "\n"

	if m.showHelp {
		// Full help view
		s += m.help.FullHelpView(m.keys.FullHelp())
	} else {
		// Short help view
		s += helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return s
}

func (k EarningsKeyMap) Help() []key.Binding {
	return k.ShortHelp()
}
