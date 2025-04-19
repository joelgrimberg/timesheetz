// Package ui provides a terminal-based user interface for the timesheet application,
// built using the Charm libraries (Bubble Tea, Bubbles, and Lip Gloss). It implements
// a monthly calendar view that allows users to navigate through time periods, view
// daily timesheet entries, and manage work hours across different categories.
//
// The main component is a TimesheetModel, which displays a table of days for a month
// with various hour categories (client work, training, vacation, etc.) that can be
// edited, copied, and pasted between days.
//
// Key features:
//   - Monthly calendar view with visual distinction for weekends
//   - Navigation between months with shortcuts
//   - Adding, editing, and deleting timesheet entries
//   - Copy/paste functionality for entries between days
//   - Column totals for different hour categories
//   - PDF export and email capabilities
//   - Vim-inspired keybindings for efficient navigation
//
// Example usage:
//
//	p := tea.NewProgram(ui.InitialTimesheetModel())
//	if _, err := p.Run(); err != nil {
//	    log.Fatal("Error running program:", err)
//	}

package ui

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/db"
	printExcel "timesheet/internal/print-excel"
	printPDF "timesheet/internal/print-pdf"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Key bindings
type TimesheetKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	GotoToday   key.Binding
	Help        key.Binding
	Quit        key.Binding
	Enter       key.Binding
	PrevMonth   key.Binding
	NextMonth   key.Binding
	AddEntry    key.Binding
	JumpUp      key.Binding
	JumpDown    key.Binding
	ClearEntry  key.Binding
	YankEntry   key.Binding
	PasteEntry  key.Binding
	Print       key.Binding
	SendAsEmail key.Binding
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
		JumpUp: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "jump up")),
		JumpDown: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "jump down")),
		ClearEntry: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear entry")),
		YankEntry: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank entry")),
		PasteEntry: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "paste entry")),
		Print: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "print timesheet")),
		SendAsEmail: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "email timesheet")),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k TimesheetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.GotoToday, k.AddEntry, k.ClearEntry, k.YankEntry, k.PasteEntry, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k TimesheetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.JumpUp, k.JumpDown},               // first column
		{k.PrevMonth, k.NextMonth},                                          // second column - month navigation
		{k.GotoToday, k.Enter, k.AddEntry, k.ClearEntry},                    // third column
		{k.YankEntry, k.PasteEntry, k.Print, k.SendAsEmail, k.Help, k.Quit}, // fourth column
	}
}

// YankedEntry stores the copied entry data
type YankedEntry struct {
	ClientName    string
	ClientHours   int
	TrainingHours int
	VacationHours int
	IdleHours     int
	HolidayHours  int
	SickHours     int
}

// TimesheetModel represents the timesheet view
type TimesheetModel struct {
	table        table.Model
	keys         TimesheetKeyMap
	help         help.Model
	showHelp     bool
	currentYear  int
	currentMonth time.Month
	cursorRow    int            // Track the current cursor position
	columnTotals map[string]int // Store column sums
	yankedEntry  *YankedEntry   // Store yanked entry data
}

// ChangeMonthMsg is used to change the month
type ChangeMonthMsg struct {
	Year       int
	Month      time.Month
	SelectDate string // Optional date to select after changing month
	CursorRow  int    // Track cursor position
	Preserve   bool   // Whether to preserve cursor position
}

// Command to change the month with optional date selection
func ChangeMonth(year int, month time.Month, selectDate string) tea.Cmd {
	return func() tea.Msg {
		return ChangeMonthMsg{Year: year, Month: month, SelectDate: selectDate, Preserve: false}
	}
}

// Command to refresh while preserving cursor position
func RefreshPreservingCursor(year int, month time.Month, cursorRow int) tea.Cmd {
	return func() tea.Msg {
		return ChangeMonthMsg{Year: year, Month: month, CursorRow: cursorRow, Preserve: true}
	}
}

// Create the initial timesheet model
func InitialTimesheetModel() TimesheetModel {
	// Start with the current month
	now := time.Now()
	currentYear, currentMonth := now.Year(), now.Month()

	// Generate initial table and column totals
	t, totals, err := generateMonthTable(currentYear, currentMonth)
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
		cursorRow:    0,
		columnTotals: totals,
		yankedEntry:  nil,
	}
}

func (m TimesheetModel) Init() tea.Cmd {
	return nil
}

// RefreshCmd refreshes the timesheet data
func (m TimesheetModel) RefreshCmd() tea.Cmd {
	// Get current cursor position
	cursorRow := m.table.Cursor()

	// Preserve cursor position when refreshing
	return RefreshPreservingCursor(m.currentYear, m.currentMonth, cursorRow)
}

// Helper function to parse an int from a string with default value of 0
func parseIntWithDefault(s string) int {
	if s == "-" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// Helper function to check if the row has any data to yank
func hasYankableData(row []string) bool {
	// Check if there's actual data in any hours column (3-9)
	for i := 3; i <= 9; i++ {
		if row[i] != "-" && row[i] != "0" {
			return true
		}
	}

	// Also check if there's a client name
	return row[2] != "-"
}

func sendDocument(content string, sendAsEmail bool, year int, month time.Month) (string, error) {
	format := config.GetDocumentType()
	fmt.Println(format)
	fmt.Println("printing to ...")

	if format == "excel" {
		// Fetch timesheet entries
		entries, err := db.GetAllTimesheetEntries(year, month)
		if err != nil {
			return "", fmt.Errorf("error fetching timesheet entries: %v", err)
		}

		// Debug: Output entries information to a log file
		logFile, err := os.OpenFile("entries_debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("Error creating debug log: %v\n", err)
		} else {
			defer logFile.Close()
			fmt.Fprintf(logFile, "Year: %d, Month: %s\n", year, month)
			fmt.Fprintf(logFile, "Total entries found: %d\n", len(entries))
			fmt.Fprintf(logFile, "------------------------------------------\n")

			for i, entry := range entries {
				fmt.Fprintf(logFile, "Entry %d:\n", i+1)
				fmt.Fprintf(logFile, "  Date: %s\n", entry.Date)
				fmt.Fprintf(logFile, "  Client: %s\n", entry.Client_name)
				fmt.Fprintf(logFile, "  Client Hours: %d\n", entry.Client_hours)
				fmt.Fprintf(logFile, "  Training Hours: %d\n", entry.Training_hours)
				fmt.Fprintf(logFile, "  Vacation Hours: %d\n", entry.Vacation_hours)
				fmt.Fprintf(logFile, "  Idle Hours: %d\n", entry.Idle_hours)
				fmt.Fprintf(logFile, "  Holiday Hours: %d\n", entry.Holiday_hours)
				fmt.Fprintf(logFile, "  Sick Hours: %d\n", entry.Sick_hours)
				fmt.Fprintf(logFile, "  Total Hours: %d\n", entry.Total_hours)
				fmt.Fprintf(logFile, "------------------------------------------\n")
			}
		}

		// Convert database entries to TimesheetRow objects
		var timesheetRows []printExcel.TimesheetRow
		for _, entry := range entries {
			// Debug: Print each entry that you're converting
			fmt.Printf("Converting entry: Date=%s, Client=%s, Hours=%d\n",
				entry.Date, entry.Client_name, entry.Client_hours)

			row := printExcel.TimesheetRow{
				Date:          entry.Date,
				ClientName:    entry.Client_name,
				ClientHours:   float64(entry.Client_hours),
				TrainingHours: float64(entry.Training_hours),
				VacationHours: float64(entry.Vacation_hours),
				IdleHours:     float64(entry.Idle_hours),
				HolidayHours:  float64(entry.Holiday_hours),
				SickHours:     float64(entry.Sick_hours),
			}
			timesheetRows = append(timesheetRows, row)
		}

		fmt.Printf("Total timesheet rows created: %d\n", len(timesheetRows))

		// Export to Excel
		if err := printExcel.TimesheetToExcel(timesheetRows); err != nil {
			return "", err
		}

		return "Timesheet.xlsx", nil
	} else {
		fmt.Println("printing to pdf...")
		return printPDF.TimesheetToPDF(content, sendAsEmail)
	}
}

func (m TimesheetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeMonthMsg:
		// Update the current year and month in the model
		m.currentYear = msg.Year
		m.currentMonth = msg.Month

		// Generate a new table for the selected month and get column totals
		newTable, totals, err := generateMonthTable(msg.Year, msg.Month)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
		}

		m.table = newTable
		m.columnTotals = totals

		// If a specific date was requested, try to select it
		if msg.SelectDate != "" {
			for i, row := range m.table.Rows() {
				if row[0] == msg.SelectDate {
					m.table.SetCursor(i)
					m.cursorRow = i
					break
				}
			}
		} else if msg.Preserve {
			// If preserving cursor position
			rowCount := len(m.table.Rows())
			if rowCount > 0 {
				// Make sure cursor position is valid
				if msg.CursorRow >= 0 && msg.CursorRow < rowCount {
					m.table.SetCursor(msg.CursorRow)
					m.cursorRow = msg.CursorRow
				}
			}
		}

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.SendAsEmail):
			// Send as email (PDF or Excel based on configuration)
			sendAsEmail := true
			filename, err := sendDocument(m.View(), sendAsEmail, m.currentYear, m.currentMonth)
			if err != nil {
				return m, tea.Printf("Error sending timesheet: %v", err)
			}
			return m, tea.Printf("Timesheet saved to %s and sent as email", filename)

		case key.Matches(msg, m.keys.Print):
			// Print without emailing (PDF or Excel based on configuration)
			sendAsEmail := false
			filename, err := sendDocument(m.View(), sendAsEmail, m.currentYear, m.currentMonth)
			if err != nil {
				return m, tea.Printf("Error printing timesheet: %v", err)
			}
			return m, tea.Printf("Timesheet saved to %s", filename)

		case key.Matches(msg, m.keys.YankEntry):
			// Get the selected row data
			row := m.table.SelectedRow()

			// Check if there's any data to yank
			if !hasYankableData(row) {
				return m, tea.Printf("No entry to yank")
			}

			// Store the data in the yankedEntry
			clientHours := parseIntWithDefault(row[3])
			trainingHours := parseIntWithDefault(row[4])
			vacationHours := parseIntWithDefault(row[5])
			idleHours := parseIntWithDefault(row[6])
			holidayHours := parseIntWithDefault(row[7])
			sickHours := parseIntWithDefault(row[8])

			m.yankedEntry = &YankedEntry{
				ClientName:    row[2],
				ClientHours:   clientHours,
				TrainingHours: trainingHours,
				VacationHours: vacationHours,
				IdleHours:     idleHours,
				HolidayHours:  holidayHours,
				SickHours:     sickHours,
			}

			return m, tea.Printf("Entry yanked: %s", row[2])

		case key.Matches(msg, m.keys.PasteEntry):
			// Check if we have any yanked data
			if m.yankedEntry == nil {
				return m, tea.Printf("No entry to paste")
			}

			// Get the date from the selected row
			selectedDate := m.table.SelectedRow()[0]
			cursorRow := m.table.Cursor()

			// Calculate total hours
			totalHours := m.yankedEntry.ClientHours +
				m.yankedEntry.TrainingHours +
				m.yankedEntry.VacationHours +
				m.yankedEntry.IdleHours +
				m.yankedEntry.HolidayHours +
				m.yankedEntry.SickHours

			// Create entry object
			entry := db.TimesheetEntry{
				Date:           selectedDate,
				Client_name:    m.yankedEntry.ClientName,
				Client_hours:   m.yankedEntry.ClientHours,
				Training_hours: m.yankedEntry.TrainingHours,
				Vacation_hours: m.yankedEntry.VacationHours,
				Idle_hours:     m.yankedEntry.IdleHours,
				Holiday_hours:  m.yankedEntry.HolidayHours,
				Sick_hours:     m.yankedEntry.SickHours,
				Total_hours:    totalHours,
			}

			// Check if an entry already exists for this date
			existingEntry, err := db.GetTimesheetEntryByDate(selectedDate)
			if err == nil {
				// Entry exists, update it
				entry.Id = existingEntry.Id // Keep the same ID
				err = db.UpdateTimesheetEntry(entry)
			} else {
				// Entry doesn't exist, add a new one
				err = db.AddTimesheetEntry(entry)
			}

			if err != nil {
				return m, tea.Printf("Error saving entry: %v", err)
			}

			// Refresh the table but maintain cursor position
			return m, RefreshPreservingCursor(m.currentYear, m.currentMonth, cursorRow)

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.GotoToday):
			// Get today's date
			now := time.Now()
			today := now.Format("2006-01-02")

			// Always go to today's date, regardless of which month we're in
			return m, ChangeMonth(now.Year(), now.Month(), today)

		case key.Matches(msg, m.keys.Enter):
			// Get the date from the selected row
			selectedDate := m.table.SelectedRow()[0]
			return m, func() tea.Msg {
				return EditEntryMsg{Date: selectedDate}
			}

		case key.Matches(msg, m.keys.ClearEntry):
			// Get the date from the selected row
			selectedDate := m.table.SelectedRow()[0]
			// Remember current cursor position
			cursorRow := m.table.Cursor()
			// Delete the entry
			err := db.DeleteTimesheetEntryByDate(selectedDate)
			if err != nil {
				return m, tea.Printf("Error clearing entry: %v", err)
			}
			// Refresh the table but maintain cursor position
			return m, RefreshPreservingCursor(m.currentYear, m.currentMonth, cursorRow)

		case key.Matches(msg, m.keys.PrevMonth):
			// Calculate the previous month
			prevYear, prevMonth := m.currentYear, m.currentMonth-1
			if prevMonth < time.January {
				prevMonth = time.December
				prevYear--
			}
			return m, ChangeMonth(prevYear, prevMonth, "")

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
				return m, ChangeMonth(nextYear, nextMonth, "")
			}

			return m, nil
		}

		// Handle table navigation
		m.table, cmd = m.table.Update(msg)
		// Store the current cursor position
		m.cursorRow = m.table.Cursor()
		return m, cmd
	}

	return m, cmd
}

func (m TimesheetModel) View() string {
	var s string

	// Show the month and year as title
	monthYearTitle := fmt.Sprintf("%s %d", m.currentMonth.String(), m.currentYear)
	s += titleStyle.Render(monthYearTitle) + "\n"

	// Render the table
	s += baseStyle.Render(m.table.View()) + "\n"

	// Render the footer with totals
	footerContent := fmt.Sprintf("%-12s %-10s %-20s", "Total:", "", "")
	footerContent += fmt.Sprintf("           %d", m.columnTotals["clientHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["trainingHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["vacationHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["idleHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["holidayHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["sickHours"])
	footerContent += fmt.Sprintf("           %d", m.columnTotals["totalHours"])

	s += footerStyle.Render(footerContent) + "\n\n"

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
func generateMonthTable(year int, month time.Month) (table.Model, map[string]int, error) {
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Day", Width: 15},
		{Title: "Client", Width: 20},
		{Title: "Hours", Width: 10},
		{Title: "Training", Width: 10},
		{Title: "Vacation", Width: 10},
		{Title: "Idle", Width: 10},
		{Title: "Holiday", Width: 10},
		{Title: "Sick", Width: 10},
		{Title: "Total", Width: 10},
	}

	// Initialize column totals
	columnTotals := map[string]int{
		"clientHours":   0,
		"trainingHours": 0,
		"vacationHours": 0,
		"idleHours":     0,
		"holidayHours":  0,
		"sickHours":     0,
		"totalHours":    0,
	}

	// Fetch timesheet entries for the specified month
	entries, err := db.GetAllTimesheetEntries(year, month)
	if err != nil {
		return table.Model{}, columnTotals, fmt.Errorf("error fetching timesheet entries: %v", err)
	}

	// Create a map of entries by date for faster lookup
	entriesByDate := make(map[string]db.TimesheetEntry)
	for _, entry := range entries {
		entriesByDate[entry.Date] = entry

		// Add to totals
		columnTotals["clientHours"] += entry.Client_hours
		columnTotals["trainingHours"] += entry.Training_hours
		columnTotals["vacationHours"] += entry.Vacation_hours
		columnTotals["idleHours"] += entry.Idle_hours
		columnTotals["holidayHours"] += entry.Holiday_hours
		columnTotals["sickHours"] += entry.Sick_hours
		columnTotals["totalHours"] += entry.Total_hours
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
		training := "-"
		vacation := "-"
		idle := "-"
		holiday := "-"
		sick := "-"
		totalHours := "-"

		// If we have an entry for this date, use its data
		if entry, exists := entriesByDate[dateStr]; exists {
			clientName = entry.Client_name
			clientHours = fmt.Sprintf("%d", entry.Client_hours)
			training = fmt.Sprintf("%d", entry.Training_hours)
			vacation = fmt.Sprintf("%d", entry.Vacation_hours)
			idle = fmt.Sprintf("%d", entry.Idle_hours)
			holiday = fmt.Sprintf("%d", entry.Holiday_hours)
			sick = fmt.Sprintf("%d", entry.Sick_hours)
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
			training,
			vacation,
			idle,
			holiday,
			sick,
			totalHours,
		}
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(32), // Reduced height slightly to make room for footer
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

	return t, columnTotals, nil
}
