package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var (
	keywordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	GotoToday key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	PrevMonth key.Binding
	NextMonth key.Binding
}

// DefaultKeyMap returns a set of default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),

		GotoToday: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "go to today"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select entry"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h", "previous month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l", "next month"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.GotoToday, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.PrevMonth, k.NextMonth},      // second column - month navigation
		{k.GotoToday, k.Enter},          // third column
		{k.Help, k.Quit},                // fourth column
	}
}

type model struct {
	table        table.Model
	keys         KeyMap
	help         help.Model
	showHelp     bool
	ShowAll      bool
	currentYear  int
	currentMonth time.Month
}

// Custom message for changing the current month
type ChangeMonthMsg struct {
	Year  int
	Month time.Month
}

// Command to change the month
func changeMonth(year int, month time.Month) tea.Cmd {
	return func() tea.Msg {
		return ChangeMonthMsg{Year: year, Month: month}
	}
}

// Generate table for a specific month
func generateMonthTable(year int, month time.Month) (table.Model, error) {
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Day", Width: 10},
		{Title: "Client", Width: 20},
		{Title: "Hours", Width: 10},
		{Title: "Total", Width: 10},
	}

	// Fetch timesheet entries for the specified month
	entries, err := db.GetAllTimesheetEntries(year, month)
	if err != nil {
		return table.Model{}, fmt.Errorf("error fetching timesheet entries: %v", err)
	}

	// Create a map of entries by date for faster lookup
	entriesByDate := make(map[string]db.TimesheetEntry)
	for _, entry := range entries {
		entriesByDate[entry.Date] = entry
	}

	// Generate all days in the specified month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	// Create table rows for each day of the month
	rows := []table.Row{}
	for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
		dateStr := day.Format("2006-01-02")
		weekday := day.Weekday().String()

		// Default values for days without entries
		clientName := "-"
		clientHours := "-"
		totalHours := "-"

		// If we have an entry for this date, use its data
		if entry, exists := entriesByDate[dateStr]; exists {
			clientName = entry.Client_name
			clientHours = fmt.Sprintf("%d", entry.Client_hours)
			totalHours = fmt.Sprintf("%d", entry.Total_hours)
		}

		// Weekend styling - make them visually distinct
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			weekday = "💤 " + weekday // Add emoji for weekends
		}

		row := table.Row{
			dateStr,
			weekday,
			clientName,
			clientHours,
			totalHours,
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(31), // Height to show entire month
	)

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
	t.SetStyles(s)

	return t, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeMonthMsg:
		// Update the current year and month in the model
		m.currentYear = msg.Year
		m.currentMonth = msg.Month

		// Generate a new table for the selected month
		newTable, err := generateMonthTable(msg.Year, msg.Month)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
		}

		m.table = newTable
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.GotoToday):
			// Get today's date
			now := time.Now()

			// If we're already in the current month, just highlight today's row
			if now.Year() == m.currentYear && now.Month() == m.currentMonth {
				today := now.Format("2006-01-02")
				for i, row := range m.table.Rows() {
					if row[0] == today {
						m.table.SetCursor(i)
						break
					}
				}
				return m, nil
			}

			// Otherwise, change to the current month
			return m, changeMonth(now.Year(), now.Month())

		case key.Matches(msg, m.keys.Enter):
			return m, tea.Printf("Selected: %s", m.table.SelectedRow()[0])

		case key.Matches(msg, m.keys.PrevMonth):
			// Calculate the previous month
			prevYear, prevMonth := m.currentYear, m.currentMonth-1
			if prevMonth < time.January {
				prevMonth = time.December
				prevYear--
			}
			return m, changeMonth(prevYear, prevMonth)

		case key.Matches(msg, m.keys.NextMonth):
			// Don't allow navigating past the current month
			now := time.Now()

			// If we're already at the current month or beyond, don't go further
			if (m.currentYear > now.Year()) ||
				(m.currentYear == now.Year() && m.currentMonth >= now.Month()) {
				return m, nil
			}

			// Calculate the next month
			nextYear, nextMonth := m.currentYear, m.currentMonth+1
			if nextMonth > time.December {
				nextMonth = time.January
				nextYear++
			}

			// Only proceed if we're not going past the current month
			if (nextYear < now.Year()) ||
				(nextYear == now.Year() && nextMonth <= now.Month()) {
				return m, changeMonth(nextYear, nextMonth)
			}

			return m, nil
		}
	}

	// Handle table navigation
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string
	// s = fmt.Sprintf("%s %d\n", m.currentMonth.String(), m.currentYear) // Display current month and year
	s += baseStyle.Render(m.table.View()) + "\n"

	if m.showHelp {
		// Full help view
		s += m.help.FullHelpView(m.keys.FullHelp())
	} else {
		// Short help view
		s += helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return s
}

func initDatabase() {
	dbUser, dbPassword := GetDBCredentials()
	if dbUser == "" || dbPassword == "" {
		fmt.Println("Error: Database username or password is empty")
		os.Exit(1)
	}

	err := db.Connect(dbUser, dbPassword)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
}

func main() {
	initDatabase()
	defer db.Close()

	// Start with the current month
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()

	// Generate initial table for the current month
	t, err := generateMonthTable(currentYear, currentMonth)
	if err != nil {
		log.Fatalf("Error generating table: %v", err)
	}

	// Initialize help model
	h := help.New()

	// Create model with table and keymap
	m := model{
		table:        t,
		keys:         DefaultKeyMap(),
		help:         h,
		showHelp:     false,
		currentYear:  currentYear,
		currentMonth: currentMonth,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// Note: This code will never be reached because the tea.NewProgram().Run() above
	// will block until the program exits
	p := tea.NewProgram(model{})
	go handler.StartServer(p)
}
