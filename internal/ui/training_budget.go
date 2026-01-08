package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TrainingBudgetKeyMap defines the keybindings for the training budget view
type TrainingBudgetKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	Refresh key.Binding
	Add     key.Binding
	Clear   key.Binding
	Yank    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

// DefaultTrainingBudgetKeyMap returns the default keybindings
func DefaultTrainingBudgetKeyMap() TrainingBudgetKeyMap {
	return TrainingBudgetKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "prev year"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next year"),
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
			key.WithHelp("a", "add entry"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear entry"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank entry"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("<"),
			key.WithHelp("<", "prev tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys(">"),
			key.WithHelp(">", "next tab"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k TrainingBudgetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Left,
		k.Right,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k TrainingBudgetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
			k.HelpKey,
			k.Quit,
		},
		{
			k.Refresh,
			k.Add,
			k.Clear,
			k.Yank,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// TrainingBudgetModel represents the training budget view
type TrainingBudgetModel struct {
	table       table.Model
	currentYear int
	keys        TrainingBudgetKeyMap
	help        help.Model
	showHelp    bool
	entries     []db.TrainingBudgetEntry // Store entries to access IDs
}

// RefreshTrainingBudgetMsg is sent when the training budget should be refreshed
type RefreshTrainingBudgetMsg struct{}

// YankTrainingBudgetMsg is sent when an entry is yanked
type YankTrainingBudgetMsg struct {
	Entry db.TrainingBudgetEntry
}

// AddTrainingBudgetMsg is sent when a new entry is added
type AddTrainingBudgetMsg struct{}

// ChangeYearMsg is used to change the year
type ChangeYearMsg struct {
	Year int
}

// Command to change the year
func ChangeYear(year int) tea.Cmd {
	return func() tea.Msg {
		return ChangeYearMsg{Year: year}
	}
}

// RefreshTrainingBudgetCmd returns a command that refreshes the training budget
func RefreshTrainingBudgetCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshTrainingBudgetMsg{}
	}
}

// YankTrainingBudgetCmd returns a command that yanks an entry
func YankTrainingBudgetCmd(entry db.TrainingBudgetEntry) tea.Cmd {
	return func() tea.Msg {
		return YankTrainingBudgetMsg{Entry: entry}
	}
}

// InitialTrainingBudgetModel creates a new training budget model
func InitialTrainingBudgetModel() TrainingBudgetModel {
	// Get current year
	currentYear := time.Now().Year()

	// Create columns for the table
	columns := []table.Column{
		{Title: "Date", Width: 12},
		{Title: "Training", Width: 34},
		{Title: "Cost (€)", Width: 16},
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

	// Get training budget entries for the current year
	dataLayer := datalayer.GetDataLayer()
	entries, err := dataLayer.GetTrainingBudgetEntriesForYear(currentYear)
	if err != nil {
		return TrainingBudgetModel{
			table:       t,
			currentYear: currentYear,
			keys:        DefaultTrainingBudgetKeyMap(),
			help:        help.New(),
			showHelp:    false,
		}
	}

	// Convert entries to table rows
	var rows []table.Row
	var totalCost float64
	for _, entry := range entries {
		rows = append(rows, table.Row{
			entry.Date,
			entry.Training_name,
			fmt.Sprintf("%.2f", entry.Cost_without_vat),
		})
		totalCost += entry.Cost_without_vat
	}

	// Add total row
	rows = append(rows, table.Row{
		"Total",
		"",
		fmt.Sprintf("%.2f", totalCost),
	})

	t.SetRows(rows)

	// Select the first row by default (if there are any entries)
	if len(entries) > 0 {
		t.SetCursor(0)
	} else {
		// If no entries, select the total row
		t.SetCursor(len(rows) - 1)
	}

	return TrainingBudgetModel{
		table:       t,
		currentYear: currentYear,
		keys:        DefaultTrainingBudgetKeyMap(),
		help:        help.New(),
		showHelp:    false,
		entries:     entries,
	}
}

func (m TrainingBudgetModel) Init() tea.Cmd {
	return RefreshTrainingBudgetCmd()
}

func (m TrainingBudgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeYearMsg:
		// Update the current year in the model
		m.currentYear = msg.Year

		// Get training budget entries for the new year
		entries, err := db.GetTrainingBudgetEntriesForYear(msg.Year)
		if err != nil {
			return m, tea.Printf("Error: %v", err)
		}

		// Store entries in model
		m.entries = entries

		// Convert entries to table rows
		var rows []table.Row
		var totalCost float64
		for _, entry := range entries {
			rows = append(rows, table.Row{
				entry.Date,
				entry.Training_name,
				fmt.Sprintf("%.2f", entry.Cost_without_vat),
			})
			totalCost += entry.Cost_without_vat
		}

		// Add total row
		rows = append(rows, table.Row{
			"Total",
			"",
			fmt.Sprintf("%.2f", totalCost),
		})

		m.table.SetRows(rows)

		// Select the first row by default (if there are any entries)
		if len(entries) > 0 {
			m.table.SetCursor(0)
		} else {
			// If no entries, select the total row
			m.table.SetCursor(len(rows) - 1)
		}

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refreshCmd()
		case key.Matches(msg, m.keys.Add):
			return m, m.addEntryCmd()
		case key.Matches(msg, m.keys.Clear):
			cursorPos := m.table.Cursor()
			if cursorPos < len(m.table.Rows())-1 { // Don't allow clearing the total row
				// Use cursor position to get the entry ID from stored entries
				if cursorPos >= 0 && cursorPos < len(m.entries) {
					entryID := m.entries[cursorPos].Id

					// Delete the entry using its ID
					dataLayer := datalayer.GetDataLayer()
					if err := dataLayer.DeleteTrainingBudgetEntry(entryID); err != nil {
						return m, tea.Printf("Error deleting entry: %v", err)
					}

					// Get all entries for the current year
					entries, err := dataLayer.GetTrainingBudgetEntriesForYear(m.currentYear)
					if err != nil {
						return m, tea.Printf("Error refreshing entries: %v", err)
					}

					// Store updated entries in model
					m.entries = entries

					// Convert entries to table rows
					var rows []table.Row
					var totalCost float64
					for _, entry := range entries {
						rows = append(rows, table.Row{
							entry.Date,
							entry.Training_name,
							fmt.Sprintf("%.2f", entry.Cost_without_vat),
						})
						totalCost += entry.Cost_without_vat
					}

					// Add total row
					rows = append(rows, table.Row{
						"Total",
						"",
						fmt.Sprintf("%.2f", totalCost),
					})

					// Update the table with new rows
					m.table.SetRows(rows)

					// Adjust cursor position
					if len(entries) > 0 {
						// If we deleted the last entry, move cursor up
						if m.table.Cursor() >= len(rows)-1 {
							m.table.SetCursor(len(rows) - 2)
						}
					} else {
						// If no entries left, select the total row
						m.table.SetCursor(len(rows) - 1)
					}

					return m, nil
				}
			}
		case key.Matches(msg, m.keys.Yank):
			cursorPos := m.table.Cursor()
			if cursorPos < len(m.table.Rows())-1 { // Don't allow yanking the total row
				// Use cursor position to get the entry from stored entries
				if cursorPos >= 0 && cursorPos < len(m.entries) {
					entry := m.entries[cursorPos]

					// Create a custom struct for yanking that excludes hours and id
					type yankData struct {
						Date           string  `json:"date"`
						TrainingName   string  `json:"training_name"`
						CostWithoutVat float64 `json:"cost_without_vat"`
					}

					data := yankData{
						Date:           entry.Date,
						TrainingName:   entry.Training_name,
						CostWithoutVat: entry.Cost_without_vat,
					}

					// Convert to JSON
					jsonData, err := json.MarshalIndent(data, "", "  ")
					if err != nil {
						// Handle error
						return m, nil
					}

					// Copy to clipboard using pbcopy on macOS
					cmd := exec.Command("pbcopy")
					cmd.Stdin = strings.NewReader(string(jsonData))
					if err := cmd.Run(); err != nil {
						// Handle error
						return m, nil
					}

					// Show a message that the entry was yanked
					return m, tea.Printf("Yanked entry to clipboard")
				}
			}
		case key.Matches(msg, m.keys.Up):
			if m.table.Cursor() == 0 {
				// If at first row, go to last data row (excluding total)
				m.table.SetCursor(len(m.table.Rows()) - 2)
				return m, nil
			}
			m.table.MoveUp(0)
		case key.Matches(msg, m.keys.Down):
			if m.table.Cursor() == len(m.table.Rows())-2 { // If at last data row
				// Go to first row
				m.table.SetCursor(0)
				return m, nil
			}
			m.table.MoveDown(0)
		case key.Matches(msg, m.keys.Left):
			// Move to previous year
			return m, ChangeYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			return m, ChangeYear(m.currentYear + 1)
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m TrainingBudgetModel) View() string {
	var s string

	// Show the year as title
	yearTitle := fmt.Sprintf("Training Budget %d", m.currentYear)
	s += titleStyle.Render(yearTitle) + "\n"

	// Get the table view
	tableView := m.table.View()

	// Render the table with baseStyle
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

func (k TrainingBudgetKeyMap) Help() []key.Binding {
	return k.ShortHelp()
}

func (m TrainingBudgetModel) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		return RefreshTrainingBudgetMsg{}
	}
}

func (m TrainingBudgetModel) addEntryCmd() tea.Cmd {
	return func() tea.Msg {
		return AddTrainingBudgetMsg{}
	}
}
