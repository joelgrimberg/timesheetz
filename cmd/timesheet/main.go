package main

import (
	"fmt"
	"log"
	"os"
	"timesheet/api/handler"
	"timesheet/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
)

var (
	keywordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	ips        []string
	suspending bool
	quitting   bool
	altscreen  bool
}

func main() {
	dbUser, dbPassword := GetDBCredentials()
	err := db.Connect(dbUser, dbPassword)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	p := tea.NewProgram(model{})
	go handler.StartServer(p)

	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+z":
			m.suspending = true
			return m, tea.Suspend
		case " ":
			var cmd tea.Cmd
			if m.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}
			m.altscreen = !m.altscreen
			return m, cmd
		}
	case handler.ApiMsg:
		m.ips = append(m.ips, msg.IP)
	}
	return m, nil
}

func (m model) View() string {
	if m.suspending {
		return ""
	}

	if m.quitting {
		return "Bye!\n"
	}

	const (
		altscreenMode = " altscreen mode "
		inlineMode    = " inline mode "
	)

	var mode string
	if m.altscreen {
		mode = altscreenMode
	} else {
		mode = inlineMode
	}

	namesView := ""
	for _, name := range m.ips {
		namesView += fmt.Sprintf("{ Name: \"%s\" }\n", name)
	}

	return fmt.Sprintf("\n\n  You're in %s\n\n\n", keywordStyle.Render(mode)) +
		helpStyle.Render("  space: switch modes • ctrl-z: suspend • q: exit\n") +
		fmt.Sprintf("\nAPI Data:\n%s", namesView)
}
