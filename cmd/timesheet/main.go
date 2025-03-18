package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"timesheet/api/handler"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
)

// Application modes
type AppMode int

const (
	TimesheetMode AppMode = iota
	FormMode
)

// Form field constants
const (
	dateField = iota
	clientField
	clientHoursField
)

// Styles
var (
	baseStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	keywordStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	inputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	buttonStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
)

// ==================== TOP LEVEL APPLICATION MODEL ====================

// AppModel is the top-level model that contains both timesheet and form models
type AppModel struct {
	Mode          AppMode
	TimesheetView TimesheetModel
	FormView      FormModel
}

func (m AppModel) Init() tea.Cmd {
	// Initialize the current mode
	if m.Mode == TimesheetMode {
		return m.TimesheetView.Init()
	}
	return m.FormView.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle global keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Global quit handler
		if keyMsg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	// Handle mode-specific updates
	switch m.Mode {
	case TimesheetMode:
		// Special handling for switching to form mode
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "a" {
				m.Mode = FormMode
				// Initialize a fresh form model
				m.FormView = InitialFormModel()
				return m, m.FormView.Init()
			}
		}

		// Otherwise update timesheet view
		timesheetModel, cmd := m.TimesheetView.Update(msg)
		m.TimesheetView = timesheetModel.(TimesheetModel)
		return m, cmd

	case FormMode:
		// Check for special message to return to timesheet mode
		if _, ok := msg.(ReturnToTimesheetMsg); ok {
			m.Mode = TimesheetMode
			// Refresh the timesheet data
			return m, m.TimesheetView.RefreshCmd()
		}

		// Otherwise update form view
		formModel, cmd := m.FormView.Update(msg)
		m.FormView = formModel.(FormModel)
		return m, cmd
	}

	return m, cmd
}

func (m AppModel) View() string {
	switch m.Mode {
	case TimesheetMode:
		return m.TimesheetView.View()
	case FormMode:
		return m.FormView.View()
	}
	return "Unknown mode"
}

// Message to return to timesheet mode
type ReturnToTimesheetMsg struct{}

func ReturnToTimesheet() tea.Cmd {
	return func() tea.Msg {
		return ReturnToTimesheetMsg{}
	}
}

// ==================== TIMESHEET MODEL ====================

// Key bindings
type TimesheetKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	GotoToday key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	PrevMonth key.Binding
	NextMonth key.Binding
	AddEntry  key.Binding
}

// Default keybindings for the timesheet view
func DefaultTimesheetKeyMap() TimesheetKeyMap {
	return TimesheetKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		GotoToday: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "go to today"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select entry"),
		),
		PrevMonth: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h", "previous month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l", "next month"),
		),
		AddEntry: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add entry"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k TimesheetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.GotoToday, k.AddEntry, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k TimesheetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},    // first column
		{k.PrevMonth, k.NextMonth},         // second column - month navigation
		{k.GotoToday, k.Enter, k.AddEntry}, // third column
		{k.Help, k.Quit},                   // fourth column
	}
}

// TimesheetModel represents the timesheet view
type TimesheetModel struct {
	table        table.Model
	keys         TimesheetKeyMap
	help         help.Model
	showHelp     bool
	currentYear  int
	currentMonth time.Month
}

// ChangeMonthMsg is used to change the month
type ChangeMonthMsg struct {
	Year  int
	Month time.Month
}

// Command to change the month
func ChangeMonth(year int, month time.Month) tea.Cmd {
	return func() tea.Msg {
		return ChangeMonthMsg{Year: year, Month: month}
	}
}

// Create the initial timesheet model
func InitialTimesheetModel() TimesheetModel {
	// Start with the current month
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()

	// Generate initial table
	t, err := generateMonthTable(currentYear, currentMonth)
	if err != nil {
		log.Fatalf("Error generating table: %v", err)
	}

	// Create model
	return TimesheetModel{
		table:        t,
		keys:         DefaultTimesheetKeyMap(),
		help:         help.New(),
		showHelp:     false,
		currentYear:  currentYear,
		currentMonth: currentMonth,
	}
}

func (m TimesheetModel) Init() tea.Cmd {
	return nil
}

// RefreshCmd refreshes the timesheet data
func (m TimesheetModel) RefreshCmd() tea.Cmd {
	return ChangeMonth(m.currentYear, m.currentMonth)
}

func (m TimesheetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeMonthMsg:
		// Update the current year and month in the model
		m.currentYear = msg.Year
		m.currentMonth = msg.Month

		// Generate a new table for the selected month
		newTable, err := generateMonthTable(msg.Year, msg.Month)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
		}

		m.table = newTable
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.GotoToday):
			// Get today's date
			now := time.Now()

			// If we're already in the current month, just highlight today's row
			if now.Year() == m.currentYear && now.Month() == m.currentMonth {
				today := now.Format("2006-01-02")
				for i, row := range m.table.Rows() {
					if row[0] == today {
						m.table.SetCursor(i)
						break
					}
				}
				return m, nil
			}

			// Otherwise, change to the current month
			return m, ChangeMonth(now.Year(), now.Month())

		case key.Matches(msg, m.keys.Enter):
			return m, tea.Printf("Selected: %s", m.table.SelectedRow()[0])

		case key.Matches(msg, m.keys.PrevMonth):
			// Calculate the previous month
			prevYear, prevMonth := m.currentYear, m.currentMonth-1
			if prevMonth < time.January {
				prevMonth = time.December
				prevYear--
			}
			return m, ChangeMonth(prevYear, prevMonth)

		case key.Matches(msg, m.keys.NextMonth):
			// Don't allow navigating past the current month
			now := time.Now()

			// If we're already at the current month or beyond, don't go further
			if (m.currentYear > now.Year()) ||
				(m.currentYear == now.Year() && m.currentMonth >= now.Month()) {
				return m, nil
			}

			// Calculate the next month
			nextYear, nextMonth := m.currentYear, m.currentMonth+1
			if nextMonth > time.December {
				nextMonth = time.January
				nextYear++
			}

			// Only proceed if we're not going past the current month
			if (nextYear < now.Year()) ||
				(nextYear == now.Year() && nextMonth <= now.Month()) {
				return m, ChangeMonth(nextYear, nextMonth)
			}

			return m, nil
		}

		// Handle table navigation
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, cmd
}

func (m TimesheetModel) View() string {
	var s string
	s += baseStyle.Render(m.table.View()) + "\n"

	if m.showHelp {
		// Full help view
		s += m.help.FullHelpView(m.keys.FullHelp())
	} else {
		// Short help view
		s += helpStyle.Render(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return s
}

// Generate table for a specific month
func generateMonthTable(year int, month time.Month) (table.Model, error) {
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Day", Width: 10},
		{Title: "Client", Width: 20},
		{Title: "Hours", Width: 10},
		{Title: "Total", Width: 10},
	}

	// Fetch timesheet entries for the specified month
	entries, err := db.GetAllTimesheetEntries(year, month)
	if err != nil {
		return table.Model{}, fmt.Errorf("error fetching timesheet entries: %v", err)
	}

	// Create a map of entries by date for faster lookup
	entriesByDate := make(map[string]db.TimesheetEntry)
	for _, entry := range entries {
		entriesByDate[entry.Date] = entry
	}

	// Generate all days in the specified month
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	// Create table rows for each day of the month
	rows := []table.Row{}
	for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
		dateStr := day.Format("2006-01-02")
		weekday := day.Weekday().String()

		// Default values for days without entries
		clientName := "-"
		clientHours := "-"
		totalHours := "-"

		// If we have an entry for this date, use its data
		if entry, exists := entriesByDate[dateStr]; exists {
			clientName = entry.Client_name
			clientHours = fmt.Sprintf("%d", entry.Client_hours)
			totalHours = fmt.Sprintf("%d", entry.Total_hours)
		}

		// Weekend styling - make them visually distinct
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			weekday = "💤 " + weekday // Add emoji for weekends
		}

		row := table.Row{
			dateStr,
			weekday,
			clientName,
			clientHours,
			totalHours,
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(31), // Height to show entire month
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

	return t, nil
}

// ==================== FORM MODEL ====================

// FormModel for timesheet entry
type FormModel struct {
	inputs     []textinput.Model
	focused    int
	err        error
	successMsg string
	submitted  bool
}

func InitialFormModel() FormModel {
	// Set today's date as the default
	today := time.Now().Format("2006-01-02")

	// Create the input fields
	inputs := make([]textinput.Model, 3)

	inputs[dateField] = textinput.New()
	inputs[dateField].Placeholder = "YYYY-MM-DD"
	inputs[dateField].Focus()
	inputs[dateField].CharLimit = 10
	inputs[dateField].Width = 30
	inputs[dateField].Prompt = ""
	inputs[dateField].SetValue(today)
	inputs[dateField].Validate = validateDate

	inputs[clientField] = textinput.New()
	inputs[clientField].Placeholder = "Client name"
	inputs[clientField].CharLimit = 50
	inputs[clientField].Width = 30
	inputs[clientField].Prompt = ""

	inputs[clientHoursField] = textinput.New()
	inputs[clientHoursField].Placeholder = "8"
	inputs[clientHoursField].CharLimit = 5
	inputs[clientHoursField].Width = 10
	inputs[clientHoursField].Prompt = ""
	inputs[clientHoursField].Validate = validateHours

	return FormModel{
		inputs:  inputs,
		focused: 0,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Return to timesheet view
			return m, ReturnToTimesheet()

		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				// Submit the form
				err := m.submitForm()
				if err != nil {
					m.err = err
					return m, nil
				}

				// Show success message
				m.submitted = true
				m.successMsg = "Entry added successfully!"

				// Return to timesheet view after a brief delay
				return m, tea.Sequence(
					tea.Tick(time.Second, func(_ time.Time) tea.Msg {
						return ReturnToTimesheetMsg{}
					}),
				)
			} else {
				// Move to next field
				m.focused++
				for i := range m.inputs {
					if i == m.focused {
						m.inputs[i].Focus()
					} else {
						m.inputs[i].Blur()
					}
				}
			}

		case tea.KeyTab:
			// Move to next field
			m.focused = (m.focused + 1) % len(m.inputs)
			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

		case tea.KeyShiftTab:
			// Move to previous field
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			for i := range m.inputs {
				if i == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
		}

	case errMsg:
		m.err = msg.(error)
		return m, nil
	}

	// Handle character input
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) View() string {
	var view string

	view += titleStyle.Render("Add Timesheet Entry") + "\n\n"

	view += inputStyle.Render("Date (YYYY-MM-DD):") + "\n"
	view += m.inputs[dateField].View() + "\n\n"

	view += inputStyle.Render("Client:") + "\n"
	view += m.inputs[clientField].View() + "\n\n"

	view += inputStyle.Render("Client Hours:") + "\n"
	view += m.inputs[clientHoursField].View() + "\n\n"

	if !m.submitted {
		view += buttonStyle.Render("Press Enter to submit • Esc to cancel") + "\n"
	}

	if m.err != nil {
		view += "\n" + errorStyle.Render(m.err.Error())
	}

	if m.submitted {
		view += "\n" + successStyle.Render(m.successMsg)
	}

	return view
}

func (m FormModel) submitForm() error {
	// Validate required fields
	if m.inputs[dateField].Value() == "" {
		return fmt.Errorf("date is required")
	}
	if m.inputs[clientField].Value() == "" {
		return fmt.Errorf("client is required")
	}
	if m.inputs[clientHoursField].Value() == "" {
		return fmt.Errorf("client hours are required")
	}

	// Parse values
	date := m.inputs[dateField].Value()
	client := m.inputs[clientField].Value()
	clientHoursStr := m.inputs[clientHoursField].Value()

	clientHours, err := strconv.Atoi(clientHoursStr)
	if err != nil {
		return fmt.Errorf("invalid client hours: %v", err)
	}

	// Create the timesheet entry
	entry := db.TimesheetEntry{
		Date:         date,
		Client_name:  client,
		Client_hours: clientHours,
		Total_hours:  clientHours, // Default total hours to client hours
	}

	// Save to database
	err = db.AddTimesheetEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to save entry: %v", err)
	}

	return nil
}

// Helper functions for form validation
func validateDate(s string) error {
	if s == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return nil
}

func validateHours(s string) error {
	if s == "" {
		return nil
	}
	hours, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("hours must be a number")
	}
	if hours < 0 {
		return fmt.Errorf("hours cannot be negative")
	}
	return nil
}

type errMsg error

// ==================== MAIN FUNCTION ====================

func initDatabase() {
	dbUser, dbPassword := GetDBCredentials()
	if dbUser == "" || dbPassword == "" {
		fmt.Println("Error: Database username or password is empty")
		os.Exit(1)
	}

	err := db.Connect(dbUser, dbPassword)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
}

func main() {
	initDatabase()
	defer db.Close()

	// Initialize the app with timesheet as the default view
	app := AppModel{
		Mode:          TimesheetMode,
		TimesheetView: InitialTimesheetModel(),
	}

	// Run the program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// API server (this won't be reached during normal operation)
	apiP := tea.NewProgram(AppModel{})
	go handler.StartServer(apiP)
}
