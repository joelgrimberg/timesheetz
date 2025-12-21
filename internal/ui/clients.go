package ui

import (
	"fmt"
	"strconv"
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

// ClientsKeyMap defines the keybindings for the clients view
type ClientsKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	HelpKey     key.Binding
	Quit        key.Binding
	Refresh     key.Binding
	Add         key.Binding
	Edit        key.Binding
	Delete      key.Binding
	ViewRates   key.Binding
	AddRate     key.Binding
	PrevTab     key.Binding
	NextTab     key.Binding
	ToggleState key.Binding
}

// DefaultClientsKeyMap returns the default keybindings
func DefaultClientsKeyMap() ClientsKeyMap {
	return ClientsKeyMap{
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
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add client"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit client"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deactivate"),
		),
		ViewRates: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view rates"),
		),
		AddRate: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new rate"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("<"),
			key.WithHelp("<", "prev tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys(">"),
			key.WithHelp(">", "next tab"),
		),
		ToggleState: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle active"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k ClientsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k ClientsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.HelpKey,
			k.Quit,
		},
		{
			k.Refresh,
			k.Add,
			k.Edit,
			k.Delete,
			k.ToggleState,
		},
		{
			k.ViewRates,
			k.AddRate,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// ClientsModel represents the clients view
type ClientsModel struct {
	table      table.Model
	clients    []db.Client
	keys       ClientsKeyMap
	help       help.Model
	showHelp   bool
	showActive bool // Filter to show only active clients
}

// RefreshClientsMsg is sent when the clients should be refreshed
type RefreshClientsMsg struct{}

// AddClientMsg is sent when adding a new client
type AddClientMsg struct{}

// EditClientMsg is sent when editing a client
type EditClientMsg struct {
	Client db.Client
}

// ViewClientRatesMsg is sent when viewing rates for a client
type ViewClientRatesMsg struct {
	ClientId int
}

// AddClientRateMsg is sent when adding a rate to a client
type AddClientRateMsg struct {
	ClientId int
}

// RefreshClientsCmd returns a command that refreshes the clients
func RefreshClientsCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshClientsMsg{}
	}
}

// InitialClientsModel creates a new clients model
func InitialClientsModel() ClientsModel {
	// Create columns for the table
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 30},
		{Title: "Current Rate", Width: 16},
		{Title: "Active", Width: 10},
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

	model := ClientsModel{
		table:      t,
		clients:    []db.Client{},
		keys:       DefaultClientsKeyMap(),
		help:       help.New(),
		showHelp:   false,
		showActive: false, // Show all clients by default
	}

	// Load initial data
	model.loadClients()

	return model
}

func (m *ClientsModel) loadClients() {
	dataLayer := datalayer.GetDataLayer()
	var clients []db.Client
	var err error

	if m.showActive {
		clients, err = dataLayer.GetActiveClients()
	} else {
		clients, err = dataLayer.GetAllClients()
	}

	if err != nil {
		m.clients = []db.Client{}
		return
	}

	m.clients = clients

	// Convert clients to table rows
	var rows []table.Row
	for _, client := range clients {
		// Get current rate for this client
		currentRate := "-"
		rates, err := dataLayer.GetClientRates(client.Id)
		if err == nil && len(rates) > 0 {
			// Find the most recent rate (highest effective date that's <= today)
			var latestRate *db.ClientRate
			today := time.Now().Format("2006-01-02")
			for i := range rates {
				if rates[i].EffectiveDate <= today {
					if latestRate == nil || rates[i].EffectiveDate > latestRate.EffectiveDate {
						latestRate = &rates[i]
					}
				}
			}
			if latestRate != nil {
				currentRate = utils.FormatEuro(latestRate.HourlyRate)
			}
		}

		activeStr := "No"
		if client.IsActive {
			activeStr = "Yes"
		}

		rows = append(rows, table.Row{
			strconv.Itoa(client.Id),
			client.Name,
			currentRate,
			activeStr,
		})
	}

	m.table.SetRows(rows)

	// Select the first row by default
	if len(clients) > 0 {
		m.table.SetCursor(0)
	}
}

func (m ClientsModel) Init() tea.Cmd {
	return RefreshClientsCmd()
}

func (m ClientsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case RefreshClientsMsg:
		m.loadClients()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.loadClients()
			return m, nil
		case key.Matches(msg, m.keys.Add):
			return m, func() tea.Msg {
				return AddClientMsg{}
			}
		case key.Matches(msg, m.keys.Edit):
			if len(m.clients) > 0 && m.table.Cursor() < len(m.clients) {
				client := m.clients[m.table.Cursor()]
				return m, func() tea.Msg {
					return EditClientMsg{Client: client}
				}
			}
		case key.Matches(msg, m.keys.Delete):
			if len(m.clients) > 0 && m.table.Cursor() < len(m.clients) {
				client := m.clients[m.table.Cursor()]
				dataLayer := datalayer.GetDataLayer()
				if err := dataLayer.DeactivateClient(client.Id); err != nil {
					return m, tea.Printf("Error deactivating client: %v", err)
				}
				m.loadClients()
				return m, nil
			}
		case key.Matches(msg, m.keys.ToggleState):
			if len(m.clients) > 0 && m.table.Cursor() < len(m.clients) {
				client := m.clients[m.table.Cursor()]
				client.IsActive = !client.IsActive
				dataLayer := datalayer.GetDataLayer()
				if err := dataLayer.UpdateClient(client); err != nil {
					return m, tea.Printf("Error updating client: %v", err)
				}
				m.loadClients()
				return m, nil
			}
		case key.Matches(msg, m.keys.ViewRates):
			if len(m.clients) > 0 && m.table.Cursor() < len(m.clients) {
				client := m.clients[m.table.Cursor()]
				return m, func() tea.Msg {
					return ViewClientRatesMsg{ClientId: client.Id}
				}
			}
		case key.Matches(msg, m.keys.AddRate):
			if len(m.clients) > 0 && m.table.Cursor() < len(m.clients) {
				client := m.clients[m.table.Cursor()]
				return m, func() tea.Msg {
					return AddClientRateMsg{ClientId: client.Id}
				}
			}
		case key.Matches(msg, m.keys.Up):
			if m.table.Cursor() == 0 {
				// If at first row, go to last row
				m.table.SetCursor(len(m.table.Rows()) - 1)
				return m, nil
			}
			m.table.MoveUp(0)
		case key.Matches(msg, m.keys.Down):
			if m.table.Cursor() == len(m.table.Rows())-1 {
				// If at last row, go to first row
				m.table.SetCursor(0)
				return m, nil
			}
			m.table.MoveDown(0)
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ClientsModel) View() string {
	var s string

	// Title
	filterText := "All Clients"
	if m.showActive {
		filterText = "Active Clients"
	}
	title := fmt.Sprintf("Client Management - %s", filterText)
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

func (k ClientsKeyMap) Help() []key.Binding {
	return k.ShortHelp()
}
