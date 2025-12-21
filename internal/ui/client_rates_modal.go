package ui

import (
	"fmt"
	"strconv"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"
	"timesheet/internal/utils"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ClientRatesModalKeyMap defines the keybindings for the client rates modal
type ClientRatesModalKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Quit    key.Binding
	Add     key.Binding
	Delete  key.Binding
	HelpKey key.Binding
}

// DefaultClientRatesModalKeyMap returns the default keybindings
func DefaultClientRatesModalKeyMap() ClientRatesModalKeyMap {
	return ClientRatesModalKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "close"),
		),
		Add: key.NewBinding(
			key.WithKeys("a", "n"),
			key.WithHelp("a/n", "add rate"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		HelpKey: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k ClientRatesModalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Add, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k ClientRatesModalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Add, k.Delete},
		{k.HelpKey, k.Quit},
	}
}

type ClientRatesViewMode int

const (
	RatesViewMode ClientRatesViewMode = iota
	RatesAddMode
)

type ClientRatesModalModel struct {
	client     db.Client
	rates      []db.ClientRate
	table      table.Model
	keys       ClientRatesModalKeyMap
	help       help.Model
	showHelp   bool
	mode       ClientRatesViewMode
	inputs     []textinput.Model
	focusIndex int
	err        error
}

func InitialClientRatesModalModel(clientId int) ClientRatesModalModel {
	dataLayer := datalayer.GetDataLayer()
	client, _ := dataLayer.GetClientById(clientId)
	rates, _ := dataLayer.GetClientRates(clientId)

	columns := []table.Column{
		{Title: "Effective Date", Width: 15},
		{Title: "Hourly Rate", Width: 15},
		{Title: "Notes", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
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

	// Create inputs for adding rates
	inputs := make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "YYYY-MM-DD"
	inputs[0].CharLimit = 10
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "100.00"
	inputs[1].CharLimit = 10

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Optional notes"
	inputs[2].CharLimit = 100

	model := ClientRatesModalModel{
		client:   client,
		rates:    rates,
		table:    t,
		keys:     DefaultClientRatesModalKeyMap(),
		help:     help.New(),
		showHelp: false,
		mode:     RatesViewMode,
		inputs:   inputs,
	}

	model.loadRates()

	return model
}

func (m *ClientRatesModalModel) loadRates() {
	dataLayer := datalayer.GetDataLayer()
	rates, _ := dataLayer.GetClientRates(m.client.Id)
	m.rates = rates

	var rows []table.Row
	for _, rate := range rates {
		rows = append(rows, table.Row{
			rate.EffectiveDate,
			utils.FormatEuro(rate.HourlyRate),
			rate.Notes,
		})
	}

	m.table.SetRows(rows)
	if len(rows) > 0 {
		m.table.SetCursor(0)
	}
}

func (m ClientRatesModalModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ClientRatesModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.mode == RatesAddMode {
		return m.updateAddMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, func() tea.Msg {
				return CloseClientRatesModalMsg{}
			}
		case key.Matches(msg, m.keys.Add):
			m.mode = RatesAddMode
			m.focusIndex = 0
			// Clear inputs
			for i := range m.inputs {
				m.inputs[i].SetValue("")
			}
			m.inputs[0].Focus()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Delete):
			if len(m.rates) > 0 && m.table.Cursor() < len(m.rates) {
				rate := m.rates[m.table.Cursor()]
				dataLayer := datalayer.GetDataLayer()
				if err := dataLayer.DeleteClientRate(rate.Id); err != nil {
					m.err = err
				} else {
					m.loadRates()
				}
			}
		case key.Matches(msg, m.keys.Up):
			if m.table.Cursor() == 0 && len(m.table.Rows()) > 0 {
				m.table.SetCursor(len(m.table.Rows()) - 1)
			} else {
				m.table.MoveUp(1)
			}
		case key.Matches(msg, m.keys.Down):
			if m.table.Cursor() == len(m.table.Rows())-1 && len(m.table.Rows()) > 0 {
				m.table.SetCursor(0)
			} else {
				m.table.MoveDown(1)
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ClientRatesModalModel) updateAddMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.mode = RatesViewMode
			m.err = nil
			return m, nil

		case "enter":
			if m.focusIndex == len(m.inputs)-1 {
				// Submit the form
				effectiveDate := m.inputs[0].Value()
				rateStr := m.inputs[1].Value()
				notes := m.inputs[2].Value()

				if effectiveDate == "" || rateStr == "" {
					m.err = fmt.Errorf("effective date and rate are required")
					return m, nil
				}

				rate, err := strconv.ParseFloat(rateStr, 64)
				if err != nil {
					m.err = fmt.Errorf("invalid rate value")
					return m, nil
				}

				dataLayer := datalayer.GetDataLayer()
				clientRate := db.ClientRate{
					ClientId:      m.client.Id,
					HourlyRate:    rate,
					EffectiveDate: effectiveDate,
					Notes:         notes,
				}

				if err := dataLayer.AddClientRate(clientRate); err != nil {
					m.err = err
					return m, nil
				}

				m.loadRates()
				m.mode = RatesViewMode
				m.err = nil
				return m, nil
			}

			// Move to next input
			m.focusIndex++
			for i := range m.inputs {
				if i == m.focusIndex {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

		case "tab":
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			for i := range m.inputs {
				if i == m.focusIndex {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

		case "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			for i := range m.inputs {
				if i == m.focusIndex {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
		}
	}

	// Update all inputs
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ClientRatesModalModel) View() string {
	if m.mode == RatesAddMode {
		return m.viewAddMode()
	}

	var s string

	s += titleStyle.Render(fmt.Sprintf("Rates for %s", m.client.Name)) + "\n\n"
	s += m.table.View() + "\n\n"

	if m.err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) + "\n\n"
	}

	if m.showHelp {
		s += m.help.FullHelpView(m.keys.FullHelp())
	} else {
		s += helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return baseStyle.Render(s)
}

func (m ClientRatesModalModel) viewAddMode() string {
	var s string

	s += titleStyle.Render(fmt.Sprintf("Add Rate for %s", m.client.Name)) + "\n\n"

	labels := []string{"Effective Date:", "Hourly Rate:", "Notes:"}
	for i, input := range m.inputs {
		s += labels[i] + "\n"
		s += input.View() + "\n\n"
	}

	if m.err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) + "\n\n"
	}

	s += helpStyle.Render("Enter: Save (when on last field) • Tab: Next field • Esc: Cancel") + "\n"

	return baseStyle.Render(s)
}

// CloseClientRatesModalMsg signals to close the client rates modal
type CloseClientRatesModalMsg struct{}
