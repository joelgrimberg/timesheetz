package ui

import (
	"time"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application modes
type AppMode int

const (
	TimesheetMode AppMode = iota
	OverviewMode
	TrainingMode
	TrainingBudgetMode
	VacationMode
	FormMode
	TrainingBudgetFormMode
)

// RefreshMsg is sent when the database is updated
type RefreshMsg struct{}

// AppModel is the top-level model that contains both timesheet and form models
type AppModel struct {
	OverviewModel           OverviewModel
	TimesheetModel          TimesheetModel
	TrainingModel           TrainingModel
	TrainingBudgetModel     TrainingBudgetModel
	VacationModel           VacationModel
	FormModel               FormModel
	TrainingBudgetFormModel TrainingBudgetFormModel
	ActiveMode              AppMode
	Help                    help.Model
	refreshChan             chan RefreshMsg
}

func NewAppModel(addMode bool) AppModel {
	model := AppModel{
		OverviewModel:           InitialOverviewModel(),
		TimesheetModel:          InitialTimesheetModel(),
		TrainingModel:           InitialTrainingModel(),
		TrainingBudgetModel:     InitialTrainingBudgetModel(),
		VacationModel:           InitialVacationModel(),
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
	case OverviewMode:
		return m.OverviewModel.Init()
	case FormMode:
		return m.FormModel.Init()
	case TrainingMode:
		return m.TrainingModel.Init()
	case TrainingBudgetMode:
		return m.TrainingBudgetModel.Init()
	case TrainingBudgetFormMode:
		return m.TrainingBudgetFormModel.Init()
	case VacationMode:
		return m.VacationModel.Init()
	}
	return nil
}

// ReturnToTimesheetMsg is sent when returning to the timesheet view
type ReturnToTimesheetMsg struct {
	Date string // Optional: the date to select when returning
}

// ReturnToTrainingBudgetMsg is sent when returning to the training budget view
type ReturnToTrainingBudgetMsg struct{}

func ReturnToTimesheet(date ...string) tea.Cmd {
	return func() tea.Msg {
		d := ""
		if len(date) > 0 {
			d = date[0]
		}
		return ReturnToTimesheetMsg{Date: d}
	}
}

func ReturnToTrainingBudget() tea.Cmd {
	return func() tea.Msg {
		return ReturnToTrainingBudgetMsg{}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				prevMode := m.ActiveMode
				switch m.ActiveMode {
				case TimesheetMode:
					// Wrap around to the last tab
					m.ActiveMode = VacationMode
				case OverviewMode:
					m.ActiveMode = TimesheetMode
				case TrainingMode:
					m.ActiveMode = OverviewMode
				case TrainingBudgetMode:
					m.ActiveMode = TrainingMode
				case VacationMode:
					m.ActiveMode = TrainingBudgetMode
				}
				// Refresh models when switching to them
				if m.ActiveMode == TimesheetMode && prevMode != TimesheetMode {
					m.TimesheetModel = InitialTimesheetModel()
				} else if m.ActiveMode == OverviewMode && prevMode != OverviewMode {
					m.OverviewModel = InitialOverviewModel()
				} else if m.ActiveMode == TrainingMode && prevMode != TrainingMode {
					m.TrainingModel = InitialTrainingModel()
				}
			case ">":
				// Move to next tab
				prevMode := m.ActiveMode
				switch m.ActiveMode {
				case TimesheetMode:
					m.ActiveMode = OverviewMode
				case OverviewMode:
					m.ActiveMode = TrainingMode
				case TrainingMode:
					m.ActiveMode = TrainingBudgetMode
				case TrainingBudgetMode:
					m.ActiveMode = VacationMode
				case VacationMode:
					// Wrap around to the first tab
					m.ActiveMode = TimesheetMode
				}
				// Refresh models when switching to them
				if m.ActiveMode == TimesheetMode && prevMode != TimesheetMode {
					m.TimesheetModel = InitialTimesheetModel()
				} else if m.ActiveMode == OverviewMode && prevMode != OverviewMode {
					m.OverviewModel = InitialOverviewModel()
				} else if m.ActiveMode == TrainingMode && prevMode != TrainingMode {
					m.TrainingModel = InitialTrainingModel()
				}
			case "$":
				// Switch to training budget view
				m.ActiveMode = TrainingBudgetMode
			case "v":
				// Switch to vacation view
				m.ActiveMode = VacationMode
			case "r":
				// Refresh all views
				m.OverviewModel = InitialOverviewModel()
				m.TimesheetModel = InitialTimesheetModel()
				m.TrainingModel = InitialTrainingModel()
				m.TrainingBudgetModel = InitialTrainingBudgetModel()
				m.VacationModel = InitialVacationModel()
				return m, nil
			}
		}
	}

	// Handle refresh message
	if _, ok := msg.(RefreshMsg); ok {
		// Refresh all views
		m.OverviewModel = InitialOverviewModel()
		m.TimesheetModel = InitialTimesheetModel()
		m.TrainingModel = InitialTrainingModel()
		m.TrainingBudgetModel = InitialTrainingBudgetModel()
		m.VacationModel = InitialVacationModel()
		return m, nil
	}

	// Handle mode-specific updates
	switch m.ActiveMode {
	case TimesheetMode:
		// Special handling for switching to form mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "a" {
				m.ActiveMode = FormMode
				// Use the selected row's date for the form
				selectedDate := m.TimesheetModel.GetSelectedDate()
				m.FormModel = InitialFormModelWithDate(selectedDate)
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
		return m, cmd

	case FormMode:
		// Check for special message to return to timesheet mode
		if rttMsg, ok := msg.(ReturnToTimesheetMsg); ok {
			// If quitAfterSubmit is true, quit the app
			if m.FormModel.quitAfterSubmit {
				return m, tea.Quit
			}
			// Otherwise return to timesheet mode, on the correct month
			if rttMsg.Date != "" {
				// Parse year and month from date
				t, err := time.Parse("2006-01-02", rttMsg.Date)
				if err == nil {
					m.TimesheetModel = InitialTimesheetModelForMonth(t.Year(), t.Month(), rttMsg.Date)
				} else {
					m.TimesheetModel = InitialTimesheetModel()
				}
			} else {
				m.TimesheetModel = InitialTimesheetModel()
			}
			m.ActiveMode = TimesheetMode
			return m, nil
		}

		// Update form model
		formModel, cmd := m.FormModel.Update(msg)
		m.FormModel = formModel.(FormModel)
		return m, cmd

	case OverviewMode:
		// Update overview model
		overviewModel, cmd := m.OverviewModel.Update(msg)
		m.OverviewModel = overviewModel.(OverviewModel)
		return m, cmd

	case TrainingMode:
		// Update training model
		trainingModel, cmd := m.TrainingModel.Update(msg)
		m.TrainingModel = trainingModel.(TrainingModel)
		return m, cmd

	case TrainingBudgetMode:
		// Handle add entry message
		if _, ok := msg.(AddTrainingBudgetMsg); ok {
			m.ActiveMode = TrainingBudgetFormMode
			m.TrainingBudgetFormModel = InitialTrainingBudgetFormModel()
			return m, m.TrainingBudgetFormModel.Init()
		}

		// Update training budget model
		trainingBudgetModel, cmd := m.TrainingBudgetModel.Update(msg)
		m.TrainingBudgetModel = trainingBudgetModel.(TrainingBudgetModel)
		return m, cmd

	case TrainingBudgetFormMode:
		// Check for special message to return to training budget mode
		if _, ok := msg.(ReturnToTrainingBudgetMsg); ok {
			m.ActiveMode = TrainingBudgetMode
			return m, nil
		}

		// Update training budget form model
		trainingBudgetFormModel, cmd := m.TrainingBudgetFormModel.Update(msg)
		m.TrainingBudgetFormModel = trainingBudgetFormModel.(TrainingBudgetFormModel)
		return m, cmd

	case VacationMode:
		// Update vacation model
		vacationModel, cmd := m.VacationModel.Update(msg)
		m.VacationModel = vacationModel.(VacationModel)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	// Render tabs
	var renderedTabs []string
	for i, t := range []string{"Timesheet", "Overview", "Training", "Training Budget", "Vacation"} {
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
	case OverviewMode:
		content = m.OverviewModel.View()
	case TrainingMode:
		content = m.TrainingModel.View()
	case TrainingBudgetMode:
		content = m.TrainingBudgetModel.View()
	case VacationMode:
		content = m.VacationModel.View()
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
