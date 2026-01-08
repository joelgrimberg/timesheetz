package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Form field constants
const (
	dateField = iota
	clientField
	clientHoursField
	trainingHoursField
	vacationHoursField
	idleHoursField
	holidayHoursField
	sickHoursField
)

// Add to your message types
type EditEntryMsg struct {
	Date string
}

type errMsg error

// FormModel for timesheet entry
type FormModel struct {
	inputs            []textinput.Model
	focused           int
	error             string
	success           string
	isEditing         bool
	quitAfterSubmit   bool
	activeClients     []db.Client
	currentSuggestion string
}

// Create a new form with initial values
func InitialFormModel() FormModel {
	// Default to today's date
	today := time.Now().Format("2006-01-02")
	return InitialFormModelWithDate(today)
}

// Create a new form with a specific date
func InitialFormModelWithDate(date string) FormModel {
	var inputs []textinput.Model

	// Date field
	dateInput := textinput.New()
	dateInput.Placeholder = "YYYY-MM-DD"
	dateInput.CharLimit = 10
	dateInput.Width = 12
	dateInput.SetValue(date)
	dateInput.Focus()
	inputs = append(inputs, dateInput)

	// Client field
	clientInput := textinput.New()
	clientInput.Placeholder = "Client name"
	clientInput.CharLimit = 50
	clientInput.Width = 30
	inputs = append(inputs, clientInput)

	// Hours fields (client, training, vacation, idle)
	for _, label := range []string{"Client hours", "Training hours", "Vacation hours", "Idle hours", "Holiday hours", "Sick hours"} {
		i := textinput.New()
		i.Placeholder = label
		i.CharLimit = 5
		i.Width = 5
		inputs = append(inputs, i)
	}

	// Load active clients for autocomplete
	dataLayer := datalayer.GetDataLayer()
	activeClients, err := dataLayer.GetActiveClients()
	if err != nil {
		activeClients = []db.Client{} // Empty list on error
	}

	return FormModel{
		inputs:            inputs,
		focused:           0,
		isEditing:         false,
		quitAfterSubmit:   false,
		activeClients:     activeClients,
		currentSuggestion: "",
	}
}

// Prefill the form with existing entry data
func (m *FormModel) prefillFromEntry(entry db.TimesheetEntry) {
	m.inputs[clientField].SetValue(entry.Client_name)
	m.inputs[clientHoursField].SetValue(strconv.Itoa(entry.Client_hours))
	m.inputs[trainingHoursField].SetValue(strconv.Itoa(entry.Training_hours))
	m.inputs[vacationHoursField].SetValue(strconv.Itoa(entry.Vacation_hours))
	m.inputs[idleHoursField].SetValue(strconv.Itoa(entry.Idle_hours))
	m.inputs[holidayHoursField].SetValue(strconv.Itoa(entry.Holiday_hours))
	m.inputs[sickHoursField].SetValue(strconv.Itoa(entry.Sick_hours))
}

// Clear all form fields except the date
func (m *FormModel) clearForm() {
	m.inputs[clientField].SetValue("")
	m.inputs[clientHoursField].SetValue("")
	m.inputs[trainingHoursField].SetValue("")
	m.inputs[vacationHoursField].SetValue("")
	m.inputs[idleHoursField].SetValue("")
	m.inputs[holidayHoursField].SetValue("")
	m.inputs[sickHoursField].SetValue("")
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			// Return to timesheet view
			return m, ReturnToTimesheet()

		case tea.KeyEnter:
			// Submit the form when Enter is pressed on any field
			return m, m.handleSubmit()

		case tea.KeyTab:
			// If in client field with a suggestion, autocomplete it
			if m.focused == clientField && m.currentSuggestion != "" {
				m.inputs[clientField].SetValue(m.currentSuggestion)
				m.currentSuggestion = ""
				return m, nil
			}
			// Otherwise, handle as normal navigation (continue below)
			fallthrough

		case tea.KeyShiftTab, tea.KeyUp, tea.KeyDown:
			// If leaving the date field, check if entry exists for that date
			if m.focused == dateField {
				date := m.inputs[dateField].Value()
				if isValidDate(date) {
					// Try to load existing entry for this date
					dataLayer := datalayer.GetDataLayer()
					entry, err := dataLayer.GetTimesheetEntryByDate(date)
					if err == nil {
						// Entry exists, populate the form
						m.prefillFromEntry(entry)
						m.isEditing = true
					} else {
						// No entry exists, clear the form
						m.clearForm()
						m.isEditing = false
					}
				}
			}

			// Handle navigation between fields
			// Change focus
			if msg.Type == tea.KeyUp || msg.Type == tea.KeyShiftTab {
				m.focused--
				if m.focused < 0 {
					m.focused = len(m.inputs) - 1
				}
			} else {
				m.focused++
				if m.focused >= len(m.inputs) {
					m.focused = 0
				}
			}

			for i := range m.inputs {
				if i == m.focused {
					cmds = append(cmds, m.inputs[i].Focus())
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle field updates
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Only update the focused input
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)

	// Update autocomplete suggestion if we're in the client field
	if m.focused == clientField {
		m.updateAutocompleteSuggestion()
	}

	return cmd
}

func (m FormModel) View() string {
	var s string

	// Add title
	if m.isEditing {
		s += titleStyle.Render("Edit Timesheet Entry") + "\n\n"
	} else {
		s += titleStyle.Render("New Timesheet Entry") + "\n\n"
	}

	// Render input fields
	for i, input := range m.inputs {
		s += inputStyle.Render(fieldLabel(i)) + "\n"

		// Special handling for client field with autocomplete suggestion
		if i == clientField && m.focused == clientField && m.currentSuggestion != "" {
			typedText := m.inputs[clientField].Value()
			suggestionRemaining := m.currentSuggestion[len(typedText):]

			// Manually construct the view with inline suggestion
			greyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

			// Render: prompt + typed text + cursor + grey suggestion
			prompt := input.PromptStyle.Render("> ")
			styledTypedText := input.TextStyle.Render(typedText)
			cursor := input.Cursor.View()
			greySuggestion := greyStyle.Render(suggestionRemaining)

			s += prompt + styledTypedText + cursor + greySuggestion + "\n\n"
		} else {
			s += input.View() + "\n\n"
		}
	}

	// Show validation errors or success messages
	if m.error != "" {
		s += errorStyle.Render(m.error) + "\n\n"
	}

	if m.success != "" {
		s += successStyle.Render(m.success) + "\n\n"
	}

	// Add help text
	s += helpStyle.Render("Tab/Shift+Tab: Navigate • Enter: Submit • Esc: Cancel") + "\n"

	return baseStyle.Render(s)
}

func (m FormModel) handleSubmit() tea.Cmd {
	// Reset messages
	m.error = ""
	m.success = ""

	// Validate input fields
	date := m.inputs[dateField].Value()
	if !isValidDate(date) {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid date format, must be YYYY-MM-DD"))
		}
	}

	clientName := m.inputs[clientField].Value()

	// Validate and parse hours
	clientHours, err := parseHours(m.inputs[clientHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid client hours: %v", err))
		}
	}

	trainingHours, err := parseHours(m.inputs[trainingHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid training hours: %v", err))
		}
	}

	vacationHours, err := parseHours(m.inputs[vacationHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid vacation hours: %v", err))
		}
	}

	idleHours, err := parseHours(m.inputs[idleHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid idle hours: %v", err))
		}
	}

	holidayHours, err := parseHours(m.inputs[holidayHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid holiday hours: %v", err))
		}
	}

	sickHours, err := parseHours(m.inputs[sickHoursField].Value())
	if err != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("invalid sick hours: %v", err))
		}
	}

	// Calculate total hours
	totalHours := clientHours + trainingHours + vacationHours + idleHours + holidayHours + sickHours

	// Validate that at least some hours are entered
	if totalHours == 0 {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("at least one hour field must be filled"))
		}
	}

	// If client hours are specified but no client name, require client name
	if clientHours > 0 && clientName == "" {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("client name is required when client hours are specified"))
		}
	}

	// Use default client name for non-client hours entries
	if clientName == "" {
		clientName = "-"
	}

	// Save to database
	entry := db.TimesheetEntry{
		Date:           date,
		Client_name:    clientName,
		Client_hours:   clientHours,
		Training_hours: trainingHours,
		Vacation_hours: vacationHours,
		Idle_hours:     idleHours,
		Holiday_hours:  holidayHours,
		Sick_hours:     sickHours,
		Total_hours:    totalHours,
	}

	var saveErr error
	if m.isEditing {
		dataLayer := datalayer.GetDataLayer()
		saveErr = dataLayer.UpdateTimesheetEntry(entry)
	} else {
		dataLayer := datalayer.GetDataLayer()
		saveErr = dataLayer.AddTimesheetEntry(entry)
	}

	if saveErr != nil {
		return func() tea.Msg {
			return errMsg(fmt.Errorf("failed to save entry: %v", saveErr))
		}
	}

	// If quitAfterSubmit is true, quit the app
	if m.quitAfterSubmit {
		return tea.Quit
	}

	// Otherwise return to timesheet view
	return ReturnToTimesheet(entry.Date)
}

// Helper functions

func fieldLabel(i int) string {
	labels := []string{
		"Date (YYYY-MM-DD):",
		"Client Name:",
		"Client Hours:",
		"Training Hours:",
		"Vacation Hours:",
		"Idle Hours:",
		"Holiday Hours:",
		"Sick Hours:",
	}
	return labels[i]
}

func isValidDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

func parseHours(input string) (int, error) {
	if input == "" {
		return 0, nil
	}

	hours, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("must be a number")
	}

	if hours < 0 {
		return 0, fmt.Errorf("cannot be negative")
	}

	return hours, nil
}

// updateAutocompleteSuggestion finds and updates the autocomplete suggestion
func (m *FormModel) updateAutocompleteSuggestion() {
	typedText := m.inputs[clientField].Value()

	// Clear suggestion if nothing is typed
	if typedText == "" {
		m.currentSuggestion = ""
		return
	}

	// Find first matching client (case-insensitive)
	typedLower := strings.ToLower(typedText)

	for _, client := range m.activeClients {
		clientNameLower := strings.ToLower(client.Name)
		if strings.HasPrefix(clientNameLower, typedLower) {
			// Found a match - store the full client name as suggestion
			m.currentSuggestion = client.Name
			return
		}
	}

	// No match found
	m.currentSuggestion = ""
}
