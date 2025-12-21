package ui

import (
	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ClientFormType int

const (
	ClientFormAdd ClientFormType = iota
	ClientFormEdit
)

type ClientFormModel struct {
	inputs     []textinput.Model
	focusIndex int
	mode       ClientFormType
	client     db.Client
	isActive   bool
	err        error
}

func InitialClientFormModel() ClientFormModel {
	m := ClientFormModel{
		inputs:   make([]textinput.Model, 1),
		isActive: true, // Default to active for new clients
	}

	t := textinput.New()
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CharLimit = 100
	t.Placeholder = "Client Name"
	t.Focus()
	t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m.inputs[0] = t

	return m
}

func (m ClientFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ClientFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Cancel and return to clients view
			return m, func() tea.Msg {
				return SwitchToClientsMsg{}
			}

		case "enter":
			// Submit the form
			clientName := m.inputs[0].Value()
			if clientName == "" {
				m.err = nil
				return m, nil
			}

			dataLayer := datalayer.GetDataLayer()

			if m.mode == ClientFormAdd {
				// Add new client
				client := db.Client{
					Name:     clientName,
					IsActive: m.isActive,
				}

				_, err := dataLayer.AddClient(client)
				if err != nil {
					m.err = err
					return m, nil
				}
			} else {
				// Edit existing client
				m.client.Name = clientName
				m.client.IsActive = m.isActive

				err := dataLayer.UpdateClient(m.client)
				if err != nil {
					m.err = err
					return m, nil
				}
			}

			// Return to clients view
			return m, func() tea.Msg {
				return SwitchToClientsMsg{}
			}

		case "tab":
			// Toggle active status
			m.isActive = !m.isActive
		}
	}

	// Update inputs
	cmd := m.updateInputs(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ClientFormModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return cmd
}

func (m ClientFormModel) View() string {
	var s string

	if m.mode == ClientFormAdd {
		s += titleStyle.Render("Add New Client") + "\n\n"
	} else {
		s += titleStyle.Render("Edit Client") + "\n\n"
	}

	s += m.inputs[0].View() + "\n\n"

	// Active status toggle
	activeStatus := "[ ] Active"
	if m.isActive {
		activeStatus = "[x] Active"
	}
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(activeStatus) + "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(Press Tab to toggle)") + "\n\n"

	if m.err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) + "\n\n"
	}

	s += helpStyle.Render("Enter: Save • Esc: Cancel") + "\n"

	return baseStyle.Render(s)
}

func (m *ClientFormModel) SetAddMode() {
	m.mode = ClientFormAdd
	m.isActive = true
	m.inputs[0].SetValue("")
	m.inputs[0].Focus()
	m.err = nil
}

func (m *ClientFormModel) SetEditMode(client db.Client) {
	m.mode = ClientFormEdit
	m.client = client
	m.isActive = client.IsActive
	m.inputs[0].SetValue(client.Name)
	m.inputs[0].Focus()
	m.err = nil
}

// SwitchToClientsMsg signals to return to clients view
type SwitchToClientsMsg struct{}
