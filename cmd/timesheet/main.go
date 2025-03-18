package main

import (
	"fmt"
	"log"
	"os"
	"timesheet/api/handler"
	"timesheet/internal/db"

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
		{Title: "Date", Width: 4},
		{Title: "Client", Width: 10},
		{Title: "Client Hours", Width: 10},
		{Title: "Training", Width: 10},
		{Title: "Vacation", Width: 10},
		{Title: "Idle", Width: 10},
	}

	rows := []table.Row{
		{"1", "TerraIndex", "9", "0", "0", "0"},
		{"1", "-", "0", "9", "0", "0"},
		{"1", "TerraIndex", "9", "0", "0", "0"},
		{"1", "TerraIndex", "9", "0", "0", "0"},
		{"1", "TerraIndex", "9", "0", "0", "0"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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
