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
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "move right"),
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
			key.WithKeys("h"),
			key.WithHelp("h", "previous month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("l"),
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
	table    table.Model
	keys     KeyMap
	help     help.Model
	showHelp bool
	ShowAll  bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.GotoToday):
			// Find today's date in the table and select it
			today := time.Now().Format("2006-01-02")
			for i, row := range m.table.Rows() {
				if row[0] == today {
					m.table.SetCursor(i)
					break
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Enter):
			return m, tea.Printf("Selected: %s", m.table.SelectedRow()[0])

		case key.Matches(msg, m.keys.PrevMonth):
			// Functionality to be implemented later
			return m, tea.Printf("Previous month")

		case key.Matches(msg, m.keys.NextMonth):
			// Functionality to be implemented later
			return m, tea.Printf("Next month")
		}
	}

	// Handle table navigation
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string
	s = baseStyle.Render(m.table.View()) + "\n"

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

	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Day", Width: 10},
		{Title: "Client", Width: 20},
		{Title: "Hours", Width: 10},
		{Title: "Total", Width: 10},
	}

	// Fetch all timesheet entries
	entries, err := db.GetAllTimesheetEntries()
	if err != nil {
		log.Printf("Error fetching timesheet entries: %v", err)
		// Continue with empty table if there's an error
	}

	// Create a map of entries by date for faster lookup
	entriesByDate := make(map[string]db.TimesheetEntry)
	for _, entry := range entries {
		entriesByDate[entry.Date] = entry
	}

	// Generate all days in March 2025
	year, month := 2025, time.March
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

	// Initialize help model
	h := help.New()

	// Create model with table and keymap
	m := model{
		table:    t,
		keys:     DefaultKeyMap(),
		help:     h,
		showHelp: false,
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
