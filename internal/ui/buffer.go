package ui

import (
	"fmt"
	"time"
	"timesheet/internal/datalayer"
	"timesheet/internal/db"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BufferKeyMap defines the keybindings for the buffer view
type BufferKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

func DefaultBufferKeyMap() BufferKeyMap {
	return BufferKeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev year")),
		Right:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next year")),
		HelpKey: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add entry")),
		Edit:    key.NewBinding(key.WithKeys("e", "enter"), key.WithHelp("e/↵", "edit entry")),
		Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete entry")),
		PrevTab: key.NewBinding(key.WithKeys("<"), key.WithHelp("<", "prev tab")),
		NextTab: key.NewBinding(key.WithKeys(">"), key.WithHelp(">", "next tab")),
	}
}

func (k BufferKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Add, k.Edit, k.Delete, k.HelpKey, k.Quit}
}

func (k BufferKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.HelpKey, k.Quit},
		{k.Add, k.Edit, k.Delete},
		{k.PrevTab, k.NextTab},
	}
}

// BufferModel represents the Buffer (banked overtime) tab
type BufferModel struct {
	table       table.Model
	currentYear int
	entries     []db.BufferEntry
	totalHours  int
	keys        BufferKeyMap
	help        help.Model
	showHelp    bool
}

// ChangeBufferYearMsg signals a year change for the Buffer tab
type ChangeBufferYearMsg struct {
	Year int
}

func ChangeBufferYear(year int) tea.Cmd {
	return func() tea.Msg { return ChangeBufferYearMsg{Year: year} }
}

// AddBufferMsg is sent when the user wants to add a new entry
type AddBufferMsg struct{}

// EditBufferMsg is sent when the user wants to edit an existing entry
type EditBufferMsg struct {
	Entry db.BufferEntry
}

func monthName(month int) string {
	if month < 1 || month > 12 {
		return fmt.Sprintf("%d", month)
	}
	return time.Month(month).String()
}

func newBufferTable() table.Model {
	columns := []table.Column{
		{Title: "Month", Width: 12},
		{Title: "Hours", Width: 8},
		{Title: "Notes", Width: 40},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
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
	s.Cell = s.Cell.Foreground(lipgloss.Color("252"))
	t.SetStyles(s)
	return t
}

func bufferRows(entries []db.BufferEntry, total int) []table.Row {
	rows := make([]table.Row, 0, len(entries)+1)
	for _, e := range entries {
		rows = append(rows, table.Row{
			monthName(e.Month),
			fmt.Sprintf("%d", e.Hours),
			e.Notes,
		})
	}
	rows = append(rows, table.Row{"Total", fmt.Sprintf("%d", total), ""})
	return rows
}

func InitialBufferModel() BufferModel {
	currentYear := time.Now().Year()
	t := newBufferTable()

	m := BufferModel{
		table:       t,
		currentYear: currentYear,
		keys:        DefaultBufferKeyMap(),
		help:        help.New(),
	}
	m.reload(currentYear)
	return m
}

func (m *BufferModel) reload(year int) {
	dl := datalayer.GetDataLayer()
	entries, err := dl.GetBufferEntriesForYear(year)
	if err != nil {
		entries = nil
	}
	total, err := dl.GetBufferTotalForYear(year)
	if err != nil {
		total = 0
	}

	m.entries = entries
	m.totalHours = total
	m.table.SetRows(bufferRows(entries, total))

	if len(entries) > 0 {
		m.table.SetCursor(0)
	} else {
		// Only the total row exists; cursor at -1 so we never accidentally edit it.
		m.table.SetCursor(-1)
	}
}

func (m BufferModel) Init() tea.Cmd { return nil }

func (m BufferModel) lastSelectableRowIndex() int {
	if len(m.entries) == 0 {
		return -1
	}
	return len(m.entries) - 1
}

func (m BufferModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeBufferYearMsg:
		m.currentYear = msg.Year
		m.reload(msg.Year)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			return m, ChangeBufferYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			return m, ChangeBufferYear(m.currentYear + 1)
		case key.Matches(msg, m.keys.Add):
			return m, func() tea.Msg { return AddBufferMsg{} }
		case key.Matches(msg, m.keys.Edit):
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.entries) {
				entry := m.entries[cursor]
				return m, func() tea.Msg { return EditBufferMsg{Entry: entry} }
			}
		case key.Matches(msg, m.keys.Delete):
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.entries) {
				entry := m.entries[cursor]
				dl := datalayer.GetDataLayer()
				if err := dl.DeleteBufferEntry(entry.Year, entry.Month); err != nil {
					return m, tea.Printf("Error deleting buffer entry: %v", err)
				}
				m.reload(m.currentYear)
				return m, TriggerSync()
			}
		case key.Matches(msg, m.keys.Down):
			last := m.lastSelectableRowIndex()
			if last < 0 {
				return m, nil
			}
			if m.table.Cursor() < last {
				m.table, cmd = m.table.Update(msg)
			}
			return m, cmd
		case key.Matches(msg, m.keys.Up):
			if m.lastSelectableRowIndex() < 0 {
				return m, nil
			}
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	m.table, cmd = m.table.Update(msg)
	// Never let the cursor land on the total row.
	if last := m.lastSelectableRowIndex(); last >= 0 && m.table.Cursor() > last {
		m.table.SetCursor(last)
	}
	return m, cmd
}

func (m BufferModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ↑/↓, k/j: Move up/down\n  ←/→, h/l: Change year\n  a: Add entry  e/↵: Edit  d: Delete\n  ?: Toggle help  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + helpStyle.Render("↑/↓: Navigate • ←/→: Year • a: Add • e/↵: Edit • d: Delete • ?: Help • q: Quit • </>: Tabs")
	}

	summary := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Render(fmt.Sprintf(
			"%s\n  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Banked this year:"),
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")).Render(fmt.Sprintf("%d hours", m.totalHours)),
		))

	tableView := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(m.table.View())

	main := lipgloss.JoinHorizontal(lipgloss.Top, tableView, "  ", summary)
	return fmt.Sprintf("%s%s", main, helpView)
}
