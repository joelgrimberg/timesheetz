package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
)

type KeyMap struct {
	Up   key.Binding
	Down key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),        // actual keybindings
		key.WithHelp("↑/k", "move up"), // corresponding help text
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var (
	keywordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	table table.Model
	// ips   []string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
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

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model{})
	go handler.StartServer(p)
}
