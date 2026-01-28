package ui

import (
	"time"
	"timesheet/internal/config"
	"timesheet/internal/datalayer"

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
	ClientsMode
	EarningsMode
	ConfigMode
	FormMode
	TrainingBudgetFormMode
	ClientFormMode
	ClientRatesModalMode
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
	ClientsModel            ClientsModel
	EarningsModel           EarningsModel
	ConfigModel             ConfigModel
	FormModel               FormModel
	TrainingBudgetFormModel TrainingBudgetFormModel
	ClientFormModel         ClientFormModel
	ClientRatesModalModel   ClientRatesModalModel
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
		ClientsModel:            InitialClientsModel(),
		EarningsModel:           InitialEarningsModel(),
		ConfigModel:             InitialConfigModel(),
		FormModel:               InitialFormModel(),
		TrainingBudgetFormModel: InitialTrainingBudgetFormModel(),
		ClientFormModel:         InitialClientFormModel(),
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
	case ClientsMode:
		return m.ClientsModel.Init()
	case ClientFormMode:
		return m.ClientFormModel.Init()
	case ClientRatesModalMode:
		return m.ClientRatesModalModel.Init()
	case EarningsMode:
		return m.EarningsModel.Init()
	case ConfigMode:
		return m.ConfigModel.Init()
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

		// Only handle special keys when not in form modes or client form/modal or config editing
		configEditing := m.ActiveMode == ConfigMode && m.ConfigModel.IsEditing()
		if m.ActiveMode != FormMode && m.ActiveMode != TrainingBudgetFormMode && m.ActiveMode != ClientFormMode && m.ActiveMode != ClientRatesModalMode && !configEditing {
			// Handle tab switching
			switch keyMsg.String() {
			case "<":
				// Move to previous tab
				prevMode := m.ActiveMode
				switch m.ActiveMode {
				case TimesheetMode:
					// Wrap around to the last tab
					m.ActiveMode = ConfigMode
				case OverviewMode:
					m.ActiveMode = TimesheetMode
				case TrainingMode:
					m.ActiveMode = OverviewMode
				case TrainingBudgetMode:
					m.ActiveMode = TrainingMode
				case VacationMode:
					m.ActiveMode = TrainingBudgetMode
				case ClientsMode:
					m.ActiveMode = VacationMode
				case EarningsMode:
					m.ActiveMode = ClientsMode
				case ConfigMode:
					m.ActiveMode = EarningsMode
				}
				// Refresh models when switching to them
				if m.ActiveMode == TimesheetMode && prevMode != TimesheetMode {
					m.TimesheetModel = InitialTimesheetModel()
				} else if m.ActiveMode == OverviewMode && prevMode != OverviewMode {
					m.OverviewModel = InitialOverviewModel()
				} else if m.ActiveMode == TrainingMode && prevMode != TrainingMode {
					m.TrainingModel = InitialTrainingModel()
				} else if m.ActiveMode == ConfigMode && prevMode != ConfigMode {
					m.ConfigModel = InitialConfigModel()
					return m, m.ConfigModel.Init()
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
					m.ActiveMode = ClientsMode
				case ClientsMode:
					m.ActiveMode = EarningsMode
				case EarningsMode:
					m.ActiveMode = ConfigMode
				case ConfigMode:
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
				} else if m.ActiveMode == ConfigMode && prevMode != ConfigMode {
					m.ConfigModel = InitialConfigModel()
					return m, m.ConfigModel.Init()
				}
			case "$":
				// Switch to training budget view
				m.ActiveMode = TrainingBudgetMode
			case "v":
				// Switch to vacation view (but not when in ClientsMode, where 'v' views rates)
				if m.ActiveMode != ClientsMode {
					m.ActiveMode = VacationMode
				}
			case "r":
				// Refresh all views
				m.OverviewModel = InitialOverviewModel()
				m.TimesheetModel = InitialTimesheetModel()
				m.TrainingModel = InitialTrainingModel()
				m.TrainingBudgetModel = InitialTrainingBudgetModel()
				m.VacationModel = InitialVacationModel()
				m.ClientsModel = InitialClientsModel()
				m.EarningsModel = InitialEarningsModel()
				m.ConfigModel = InitialConfigModel()
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
		m.ClientsModel = InitialClientsModel()
		m.EarningsModel = InitialEarningsModel()
		m.ConfigModel = InitialConfigModel()
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
			dataLayer := datalayer.GetDataLayer()
			entry, err := dataLayer.GetTimesheetEntryByDate(date)
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
		// Check for navigation message
		if navMsg, ok := msg.(NavigateToTimesheetMsg); ok {
			// Parse the date to extract year and month
			date, err := time.Parse("2006-01-02", navMsg.Date)
			if err == nil {
				// Switch to timesheet mode
				m.ActiveMode = TimesheetMode

				// Initialize timesheet with the specific date selected
				m.TimesheetModel = InitialTimesheetModel()

				// Send command to change to that month and select the date
				return m, ChangeMonth(date.Year(), date.Month(), navMsg.Date)
			}
		}

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
			// Refresh the training budget model to show the new entry
			m.TrainingBudgetModel = InitialTrainingBudgetModel()
			return m, m.TrainingBudgetModel.Init()
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

	case ClientsMode:
		// Handle add/edit client messages
		if _, ok := msg.(AddClientMsg); ok {
			m.ActiveMode = ClientFormMode
			m.ClientFormModel.SetAddMode()
			return m, m.ClientFormModel.Init()
		}
		if editMsg, ok := msg.(EditClientMsg); ok {
			m.ActiveMode = ClientFormMode
			m.ClientFormModel.SetEditMode(editMsg.Client)
			return m, m.ClientFormModel.Init()
		}
		// Handle view/add client rates messages
		if viewMsg, ok := msg.(ViewClientRatesMsg); ok {
			m.ActiveMode = ClientRatesModalMode
			m.ClientRatesModalModel = InitialClientRatesModalModel(viewMsg.ClientId)
			return m, m.ClientRatesModalModel.Init()
		}
		if addMsg, ok := msg.(AddClientRateMsg); ok {
			m.ActiveMode = ClientRatesModalMode
			m.ClientRatesModalModel = InitialClientRatesModalModel(addMsg.ClientId)
			// Switch to add mode immediately
			m.ClientRatesModalModel.mode = RatesAddMode
			return m, m.ClientRatesModalModel.Init()
		}

		// Update clients model
		clientsModel, cmd := m.ClientsModel.Update(msg)
		m.ClientsModel = clientsModel.(ClientsModel)
		return m, cmd

	case ClientFormMode:
		// Check for special message to return to clients mode
		if _, ok := msg.(SwitchToClientsMsg); ok {
			m.ActiveMode = ClientsMode
			m.ClientsModel = InitialClientsModel()
			return m, nil
		}

		// Update client form model
		clientFormModel, cmd := m.ClientFormModel.Update(msg)
		m.ClientFormModel = clientFormModel.(ClientFormModel)
		return m, cmd

	case ClientRatesModalMode:
		// Check for special message to close modal
		if _, ok := msg.(CloseClientRatesModalMsg); ok {
			m.ActiveMode = ClientsMode
			m.ClientsModel = InitialClientsModel()
			return m, nil
		}

		// Update client rates modal model
		clientRatesModalModel, cmd := m.ClientRatesModalModel.Update(msg)
		m.ClientRatesModalModel = clientRatesModalModel.(ClientRatesModalModel)
		return m, cmd

	case EarningsMode:
		// Update earnings model
		earningsModel, cmd := m.EarningsModel.Update(msg)
		m.EarningsModel = earningsModel.(EarningsModel)
		return m, cmd

	case ConfigMode:
		// Handle mode selection messages from config modal
		switch msg := msg.(type) {
		case ModeSelectedMsg:
			// Save selected mode
			cfg, err := config.GetConfig()
			if err == nil {
				cfg.APIMode = msg.Mode
				config.SaveConfig(cfg)
			}
			// Refresh config model and ensure cursor is on API Mode row
			m.ConfigModel = InitialConfigModel()
			// Set cursor to API Mode row
			if m.ConfigModel.apiModeRowIdx < len(m.ConfigModel.table.Rows()) {
				m.ConfigModel.table.SetCursor(m.ConfigModel.apiModeRowIdx)
			}
			return m, nil
		case ModeCancelledMsg:
			// Just refresh config model to close modal and ensure cursor is on API Mode row
			m.ConfigModel = InitialConfigModel()
			// Set cursor to API Mode row
			if m.ConfigModel.apiModeRowIdx < len(m.ConfigModel.table.Rows()) {
				m.ConfigModel.table.SetCursor(m.ConfigModel.apiModeRowIdx)
			}
			return m, nil
		case LanguageSelectedMsg:
			cfg, err := config.GetConfig()
			if err == nil {
				cfg.ExportLanguage = msg.Language
				config.SaveConfig(cfg)
			}
			m.ConfigModel = InitialConfigModel()
			if m.ConfigModel.exportLangRowIdx < len(m.ConfigModel.table.Rows()) {
				m.ConfigModel.table.SetCursor(m.ConfigModel.exportLangRowIdx)
			}
			return m, nil
		case LanguageCancelledMsg:
			m.ConfigModel = InitialConfigModel()
			if m.ConfigModel.exportLangRowIdx < len(m.ConfigModel.table.Rows()) {
				m.ConfigModel.table.SetCursor(m.ConfigModel.exportLangRowIdx)
			}
			return m, nil
		}
		// Update config model
		configModel, cmd := m.ConfigModel.Update(msg)
		m.ConfigModel = configModel.(ConfigModel)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	// Render tabs
	var renderedTabs []string
	tabs := []string{"Timesheet", "Overview", "Training", "Training Budget", "Vacation", "Clients", "Earnings", "Config"}
	// Map tab names to their corresponding modes
	tabModes := []AppMode{TimesheetMode, OverviewMode, TrainingMode, TrainingBudgetMode, VacationMode, ClientsMode, EarningsMode, ConfigMode}
	
	for i, t := range tabs {
		var style lipgloss.Style
		if tabModes[i] == m.ActiveMode {
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
	case ClientsMode:
		content = m.ClientsModel.View()
	case ClientFormMode:
		content = m.ClientFormModel.View()
	case ClientRatesModalMode:
		content = m.ClientRatesModalModel.View()
	case EarningsMode:
		content = m.EarningsModel.View()
	case ConfigMode:
		content = m.ConfigModel.View()
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
