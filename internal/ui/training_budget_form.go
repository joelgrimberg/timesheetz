package ui

import (
	"fmt"

	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TrainingBudgetFormModel struct {
	inputs       []textinput.Model
	focusIndex   int
	date         string
	trainingName string
	cost         string
	err          error
	isEditing    bool
	entryID      int
}

func InitialTrainingBudgetFormModel() TrainingBudgetFormModel {
	m := TrainingBudgetFormModel{
		inputs:    make([]textinput.Model, 3),
		isEditing: false,
		entryID:   0,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Date (YYYY-MM-DD)"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		case 1:
			t.Placeholder = "Training Name"
		case 2:
			t.Placeholder = "Cost (without VAT)"
		}

		m.inputs[i] = t
	}

	return m
}

// InitialTrainingBudgetFormModelForEdit creates a form pre-filled with entry data for editing
func InitialTrainingBudgetFormModelForEdit(entry db.TrainingBudgetEntry) TrainingBudgetFormModel {
	m := InitialTrainingBudgetFormModel()
	m.isEditing = true
	m.entryID = entry.Id

	// Pre-fill the form fields
	m.inputs[0].SetValue(entry.Date)
	m.inputs[1].SetValue(entry.Training_name)
	m.inputs[2].SetValue(fmt.Sprintf("%.2f", entry.Cost_without_vat))

	return m
}

func (m TrainingBudgetFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TrainingBudgetFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			// Submit the form
			entry := db.TrainingBudgetEntry{
				Date:             m.inputs[0].Value(),
				Training_name:    m.inputs[1].Value(),
				Hours:            0,
				Cost_without_vat: parseTrainingCost(m.inputs[2].Value()),
			}

			dataLayer := datalayer.GetDataLayer()

			if m.isEditing {
				// Update existing entry
				entry.Id = m.entryID
				if err := dataLayer.UpdateTrainingBudgetEntry(entry); err != nil {
					m.err = err
					return m, nil
				}
			} else {
				// Add new entry
				if err := dataLayer.AddTrainingBudgetEntry(entry); err != nil {
					m.err = err
					return m, nil
				}
			}

			// Return to training budget view
			return m, ReturnToTrainingBudget()
		case "tab":
			// Move to next input
			m.nextInput()
		case "shift+tab":
			// Move to previous input
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			for i := range m.inputs {
				if i == m.focusIndex {
					m.inputs[i].Focus()
					m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
					m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = lipgloss.NewStyle()
					m.inputs[i].TextStyle = lipgloss.NewStyle()
				}
			}
		case "esc":
			// Return to training budget view
			return m, ReturnToTrainingBudget()
		}

		// Handle other key presses
		for i := range m.inputs {
			if i == m.focusIndex {
				var cmd tea.Cmd
				m.inputs[i], cmd = m.inputs[i].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m TrainingBudgetFormModel) View() string {
	var s string

	title := "Add Training Budget Entry"
	if m.isEditing {
		title = "Edit Training Budget Entry"
	}
	s += titleStyle.Render(title) + "\n\n"

	for i := range m.inputs {
		s += inputStyle.Render(m.inputs[i].Placeholder) + "\n"
		s += m.inputs[i].View() + "\n\n"
	}

	s += "\n\n"
	if m.err != nil {
		s += errorStyle.Render(m.err.Error()) + "\n"
	}

	s += helpStyle.Render("Press Enter to submit • Ctrl+C or q to quit")

	return baseStyle.Render(s)
}

func (m *TrainingBudgetFormModel) nextInput() {
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
			m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
			m.inputs[i].TextStyle = lipgloss.NewStyle()
		}
	}
}

func parseTrainingCost(s string) float64 {
	var cost float64
	fmt.Sscanf(s, "%f", &cost)
	return cost
}
