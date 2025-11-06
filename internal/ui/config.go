package ui

import (
	"fmt"
	"strconv"
	"timesheet/internal/config"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigKeyMap defines the keybindings for the config view
type ConfigKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

// DefaultConfigKeyMap returns the default keybindings
func DefaultConfigKeyMap() ConfigKeyMap {
	return ConfigKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
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
func (k ConfigKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k ConfigKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.HelpKey,
			k.Quit,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// ConfigModel represents the configuration view
type ConfigModel struct {
	table    table.Model
	keys     ConfigKeyMap
	help     help.Model
	showHelp bool
}

// InitialConfigModel creates a new config model
func InitialConfigModel() ConfigModel {
	// Create columns for the table
	columns := []table.Column{
		{Title: "Field", Width: 30},
		{Title: "Value", Width: 50},
	}

	// Create the table
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(25),
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

	// Get config
	cfg, err := config.GetConfig()
	if err != nil {
		// Return empty model if config can't be loaded
		return ConfigModel{
			table:    t,
			keys:     DefaultConfigKeyMap(),
			help:     help.New(),
			showHelp: false,
		}
	}

	// Convert config to table rows
	var rows []table.Row

	// User Information
	rows = append(rows, table.Row{"User Information", ""})
	rows = append(rows, table.Row{"  Name", cfg.Name})
	rows = append(rows, table.Row{"  Company Name", cfg.CompanyName})
	rows = append(rows, table.Row{"  Free Speech", cfg.FreeSpeech})

	// API Server Configuration
	rows = append(rows, table.Row{"API Server", ""})
	rows = append(rows, table.Row{"  Start API Server", fmt.Sprintf("%v", cfg.StartAPIServer)})
	rows = append(rows, table.Row{"  API Port", strconv.Itoa(cfg.APIPort)})

	// API Client Configuration
	rows = append(rows, table.Row{"API Client", ""})
	if cfg.APIMode == "" {
		rows = append(rows, table.Row{"  API Mode", "local (default)"})
	} else {
		rows = append(rows, table.Row{"  API Mode", cfg.APIMode})
	}
	if cfg.APIBaseURL == "" {
		rows = append(rows, table.Row{"  API Base URL", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  API Base URL", cfg.APIBaseURL})
	}

	// Database Location
	rows = append(rows, table.Row{"Database", ""})
	if cfg.DBLocation == "" {
		rows = append(rows, table.Row{"  DB Location", "(default)"})
	} else {
		rows = append(rows, table.Row{"  DB Location", cfg.DBLocation})
	}

	// Development Settings
	rows = append(rows, table.Row{"Development", ""})
	rows = append(rows, table.Row{"  Development Mode", fmt.Sprintf("%v", cfg.DevelopmentMode)})

	// Document Settings
	rows = append(rows, table.Row{"Document", ""})
	rows = append(rows, table.Row{"  Send Document Type", cfg.SendDocumentType})

	// Email Configuration
	rows = append(rows, table.Row{"Email", ""})
	rows = append(rows, table.Row{"  Send To Others", fmt.Sprintf("%v", cfg.SendToOthers)})
	rows = append(rows, table.Row{"  Recipient Email", cfg.RecipientEmail})
	rows = append(rows, table.Row{"  Sender Email", cfg.SenderEmail})
	rows = append(rows, table.Row{"  Reply To Email", cfg.ReplyToEmail})
	if cfg.ResendAPIKey != "" {
		// Mask API key for security
		maskedKey := maskAPIKey(cfg.ResendAPIKey)
		rows = append(rows, table.Row{"  Resend API Key", maskedKey})
	} else {
		rows = append(rows, table.Row{"  Resend API Key", "(not set)"})
	}

	// Training Hours Configuration
	rows = append(rows, table.Row{"Training Hours", ""})
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.TrainingHours.YearlyTarget)})
	rows = append(rows, table.Row{"  Category", cfg.TrainingHours.Category})

	// Vacation Hours Configuration
	rows = append(rows, table.Row{"Vacation Hours", ""})
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.VacationHours.YearlyTarget)})
	rows = append(rows, table.Row{"  Category", cfg.VacationHours.Category})

	t.SetRows(rows)

	// Select the first row by default
	if len(rows) > 0 {
		t.SetCursor(0)
	}

	return ConfigModel{
		table:    t,
		keys:     DefaultConfigKeyMap(),
		help:     help.New(),
		showHelp: false,
	}
}

// maskAPIKey masks the API key showing only first few and last few characters
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Down):
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Up):
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	// Handle other table updates
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m ConfigModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ↑/↓, k/j: Move up/down\n  ?: Toggle help\n  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("↑/↓: Navigate • ?: Help • q: Quit • </>: Tabs")
	}

	return fmt.Sprintf(
		"%s\n%s\n%s%s",
		titleStyle.Render("Configuration"),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(m.table.View()),
		helpView,
		m.help.View(m.keys),
	)
}

