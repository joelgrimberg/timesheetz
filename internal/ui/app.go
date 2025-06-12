package ui

import (
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application modes
type AppMode int

const (
	TimesheetMode AppMode = iota
	TrainingMode
	TrainingBudgetMode
	FormMode
	TrainingBudgetFormMode
)

// RefreshMsg is sent when the database is updated
type RefreshMsg struct{}

// AppModel is the top-level model that contains both timesheet and form models
type AppModel struct {
	TimesheetModel          TimesheetModel
	TrainingModel           TrainingModel
	TrainingBudgetModel     TrainingBudgetModel
	FormModel               FormModel
	TrainingBudgetFormModel TrainingBudgetFormModel
	ActiveMode              AppMode
	Help                    help.Model
	refreshChan             chan RefreshMsg
}

func NewAppModel(addMode bool) AppModel {
	model := AppModel{
		TimesheetModel:          InitialTimesheetModel(),
		TrainingModel:           InitialTrainingModel(),
		TrainingBudgetModel:     InitialTrainingBudgetModel(),
		FormModel:               InitialFormModel(),
		TrainingBudgetFormModel: InitialTrainingBudgetFormModel(),
		ActiveMode:              TimesheetMode,
		Help:                    help.New(),
		refreshChan:             make(chan RefreshMsg),
	}

	// If add mode is true, start in form mode for today
	if addMode {
		model.ActiveMode = FormMode
		model.FormModel = InitialFormModel()
	}

	return model
}

func (m AppModel) Init() tea.Cmd {
	// Initialize the current mode
	switch m.ActiveMode {
	case TimesheetMode:
		return m.TimesheetModel.Init()
	case FormMode:
		return m.FormModel.Init()
	case TrainingMode:
		return m.TrainingModel.Init()
	case TrainingBudgetMode:
		return m.TrainingBudgetModel.Init()
	case TrainingBudgetFormMode:
		return m.TrainingBudgetFormModel.Init()
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle global keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Global quit handler
		if keyMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		// Only handle special keys when not in form modes
		if m.ActiveMode != FormMode && m.ActiveMode != TrainingBudgetFormMode {
			// Handle tab switching
			switch keyMsg.String() {
			case "<":
				// Move to previous tab
				switch m.ActiveMode {
				case TrainingMode:
					m.ActiveMode = TimesheetMode
				case TrainingBudgetMode:
					m.ActiveMode = TrainingMode
				case TimesheetMode:
					// Wrap around to the last tab
					m.ActiveMode = TrainingBudgetMode
				}
			case ">":
				// Move to next tab
				switch m.ActiveMode {
				case TimesheetMode:
					m.ActiveMode = TrainingMode
				case TrainingMode:
					m.ActiveMode = TrainingBudgetMode
				case TrainingBudgetMode:
					// Wrap around to the first tab
					m.ActiveMode = TimesheetMode
				}
			case "$":
				// Switch to training budget view
				m.ActiveMode = TrainingBudgetMode
			case "r":
				// Refresh all views
				m.TimesheetModel = InitialTimesheetModel()
				m.TrainingModel = InitialTrainingModel()
				m.TrainingBudgetModel = InitialTrainingBudgetModel()
				return m, nil
			}
		}
	}

	// Handle refresh message
	if _, ok := msg.(RefreshMsg); ok {
		// Refresh all views
		m.TimesheetModel = InitialTimesheetModel()
		m.TrainingModel = InitialTrainingModel()
		m.TrainingBudgetModel = InitialTrainingBudgetModel()
		return m, nil
	}

	// Handle mode-specific updates
	switch m.ActiveMode {
	case TimesheetMode:
		// Special handling for switching to form mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "a" {
				m.ActiveMode = FormMode
				// Initialize a fresh form model
				m.FormModel = InitialFormModel()
				return m, m.FormModel.Init()
			}
		}

		// Handle edit entry message
		if editMsg, ok := msg.(EditEntryMsg); ok {
			// Switch to form mode for editing
			m.ActiveMode = FormMode

			// Initialize the form for editing
			date := editMsg.Date
			m.FormModel = InitialFormModelWithDate(date)

			// Try to load existing data
			entry, err := db.GetTimesheetEntryByDate(date)
			if err == nil {
				// Entry found, populate form fields
				m.FormModel.prefillFromEntry(entry)
				m.FormModel.isEditing = true
			}

			return m, m.FormModel.Init()
		}

		// Otherwise update timesheet view
		timesheetModel, cmd := m.TimesheetModel.Update(msg)
		m.TimesheetModel = timesheetModel.(TimesheetModel)

		// If the timesheet model sent a refresh message, refresh all views
		if _, ok := msg.(RefreshMsg); ok {
			m.TimesheetModel = InitialTimesheetModel()
			m.TrainingModel = InitialTrainingModel()
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
		}

		return m, cmd

	case FormMode:
		// Check for special message to return to timesheet mode
		if _, ok := msg.(ReturnToTimesheetMsg); ok {
			// If quitAfterSubmit is true, quit the app
			if m.FormModel.quitAfterSubmit {
				return m, tea.Quit
			}
			// Otherwise return to timesheet mode
			m.ActiveMode = TimesheetMode
			// Refresh all views
			m.TimesheetModel = InitialTimesheetModel()
			m.TrainingModel = InitialTrainingModel()
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
			return m, nil
		}

		// Otherwise update form view
		formModel, cmd := m.FormModel.Update(msg)
		m.FormModel = formModel.(FormModel)
		return m, cmd

	case TrainingMode:
		// Update training view
		trainingModel, cmd := m.TrainingModel.Update(msg)
		m.TrainingModel = trainingModel.(TrainingModel)

		// If the training model sent a refresh message, refresh all views
		if _, ok := msg.(RefreshMsg); ok {
			m.TimesheetModel = InitialTimesheetModel()
			m.TrainingModel = InitialTrainingModel()
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
		}

		return m, cmd

	case TrainingBudgetMode:
		// Special handling for switching to training budget form mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "a" {
				m.ActiveMode = TrainingBudgetFormMode
				// Initialize a fresh training budget form model
				m.TrainingBudgetFormModel = InitialTrainingBudgetFormModel()
				return m, m.TrainingBudgetFormModel.Init()
			}
		}

		// Update training budget view
		trainingBudgetModel, cmd := m.TrainingBudgetModel.Update(msg)
		m.TrainingBudgetModel = trainingBudgetModel.(TrainingBudgetModel)

		// If the training budget model sent a refresh message, refresh all views
		if _, ok := msg.(RefreshMsg); ok {
			m.TimesheetModel = InitialTimesheetModel()
			m.TrainingModel = InitialTrainingModel()
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
		}

		return m, cmd

	case TrainingBudgetFormMode:
		// Check for special message to return to training budget mode
		if _, ok := msg.(ReturnToTimesheetMsg); ok {
			// Return to training budget mode
			m.ActiveMode = TrainingBudgetMode
			// Refresh all views
			m.TimesheetModel = InitialTimesheetModel()
			m.TrainingModel = InitialTrainingModel()
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
			return m, nil
		}

		// Update training budget form view
		formModel, cmd := m.TrainingBudgetFormModel.Update(msg)
		m.TrainingBudgetFormModel = formModel.(TrainingBudgetFormModel)
		return m, cmd
	}

	return m, cmd
}

func (m AppModel) View() string {
	// Render tabs
	var renderedTabs []string
	for i, t := range []string{"Timesheet", "Training", "Training Budget"} {
		var style lipgloss.Style
		if i == int(m.ActiveMode) {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	// Join tabs horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Render the current view
	var content string
	switch m.ActiveMode {
	case TimesheetMode:
		content = m.TimesheetModel.View()
	case TrainingMode:
		content = m.TrainingModel.View()
	case TrainingBudgetMode:
		content = m.TrainingBudgetModel.View()
	case FormMode:
		content = m.FormModel.View()
	case TrainingBudgetFormMode:
		content = m.TrainingBudgetFormModel.View()
	}

	// Combine tabs and content
	return lipgloss.JoinVertical(lipgloss.Left, row, content)
}

// GetRefreshChan returns the refresh channel
func (m AppModel) GetRefreshChan() chan RefreshMsg {
	return m.refreshChan
}

// Message to return to timesheet mode
type ReturnToTimesheetMsg struct{}

func ReturnToTimesheet() tea.Cmd {
	return func() tea.Msg {
		return ReturnToTimesheetMsg{}
	}
}

// Tab styles
var (
	activeTabStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1)

	windowStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 0)
)
