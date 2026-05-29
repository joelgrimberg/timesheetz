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

// Focus indices for the buffer form
const (
	bufFocusMonth = iota
	bufFocusHours
	bufFocusNotes
)

// BufferFormModel is the add/edit form for a Buffer entry.
// Year is fixed by the tab; the form only edits month/hours/notes.
type BufferFormModel struct {
	hoursInput textinput.Model
	notesInput textinput.Model
	focusIndex int
	month      int // 1-12
	year       int
	isEditing  bool
	err        error
	// existingMonths tracks months already used in the year (for add-mode collision check)
	existingMonths map[int]bool
}

// ReturnToBufferMsg signals returning to the Buffer tab from the form
type ReturnToBufferMsg struct{}

func ReturnToBuffer() tea.Cmd {
	return func() tea.Msg { return ReturnToBufferMsg{} }
}

// InitialBufferFormModel builds a blank form ready for ADD mode.
func InitialBufferFormModel(year int, existing map[int]bool) BufferFormModel {
	h := textinput.New()
	h.Placeholder = "Hours (positive number)"
	h.CharLimit = 8

	n := textinput.New()
	n.Placeholder = "Notes (optional)"
	n.CharLimit = 200

	m := BufferFormModel{
		hoursInput:     h,
		notesInput:     n,
		focusIndex:     bufFocusMonth,
		month:          firstUnusedMonth(existing),
		year:           year,
		isEditing:      false,
		existingMonths: existing,
	}
	return m
}

// InitialBufferFormModelForEdit builds the form pre-filled with an existing entry.
func InitialBufferFormModelForEdit(entry db.BufferEntry) BufferFormModel {
	m := InitialBufferFormModel(entry.Year, nil)
	m.isEditing = true
	m.month = entry.Month
	m.hoursInput.SetValue(fmt.Sprintf("%d", entry.Hours))
	m.notesInput.SetValue(entry.Notes)
	return m
}

func firstUnusedMonth(existing map[int]bool) int {
	now := time.Now()
	candidate := int(now.Month())
	for offset := 0; offset < 12; offset++ {
		m := ((candidate - 1 + offset) % 12) + 1
		if !existing[m] {
			return m
		}
	}
	return int(now.Month())
}

func (m BufferFormModel) Init() tea.Cmd { return textinput.Blink }

func (m *BufferFormModel) focusInput(i int) {
	m.focusIndex = i
	if i == bufFocusHours {
		m.hoursInput.Focus()
		m.hoursInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.hoursInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	} else {
		m.hoursInput.Blur()
		m.hoursInput.PromptStyle = lipgloss.NewStyle()
		m.hoursInput.TextStyle = lipgloss.NewStyle()
	}
	if i == bufFocusNotes {
		m.notesInput.Focus()
		m.notesInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		m.notesInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	} else {
		m.notesInput.Blur()
		m.notesInput.PromptStyle = lipgloss.NewStyle()
		m.notesInput.TextStyle = lipgloss.NewStyle()
	}
}

func (m *BufferFormModel) nextField() {
	m.focusInput((m.focusIndex + 1) % 3)
}

func (m *BufferFormModel) prevField() {
	idx := m.focusIndex - 1
	if idx < 0 {
		idx = 2
	}
	m.focusInput(idx)
}

// cycleMonth moves the month forward (+1) or backward (-1), wrapping 1..12.
// In add-mode, it skips months that already have an entry; if every month is
// taken, it falls back to allowing them all (the upsert will then edit).
func (m *BufferFormModel) cycleMonth(delta int) {
	skip := !m.isEditing && len(m.existingMonths) > 0
	allTaken := skip && len(m.existingMonths) >= 12
	current := m.month
	for i := 0; i < 12; i++ {
		current += delta
		if current > 12 {
			current = 1
		}
		if current < 1 {
			current = 12
		}
		if !skip || allTaken || !m.existingMonths[current] {
			m.month = current
			return
		}
	}
}

func (m BufferFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, ReturnToBuffer()
		case "tab":
			m.nextField()
			return m, nil
		case "shift+tab":
			m.prevField()
			return m, nil
		case "left":
			if m.focusIndex == bufFocusMonth {
				m.cycleMonth(-1)
				return m, nil
			}
		case "right":
			if m.focusIndex == bufFocusMonth {
				m.cycleMonth(+1)
				return m, nil
			}
		case "enter":
			hours, err := strconv.Atoi(strings.TrimSpace(m.hoursInput.Value()))
			if err != nil {
				m.err = fmt.Errorf("hours must be a whole number")
				return m, nil
			}
			if hours < 0 {
				m.err = fmt.Errorf("hours must be 0 or greater")
				return m, nil
			}

			entry := db.BufferEntry{
				Year:  m.year,
				Month: m.month,
				Hours: hours,
				Notes: m.notesInput.Value(),
			}
			dl := datalayer.GetDataLayer()
			if err := dl.UpsertBufferEntry(entry); err != nil {
				m.err = err
				return m, nil
			}
			return m, tea.Batch(ReturnToBuffer(), TriggerSync())
		}
	}

	// Forward text input updates to the focused field
	var cmd tea.Cmd
	switch m.focusIndex {
	case bufFocusHours:
		m.hoursInput, cmd = m.hoursInput.Update(msg)
	case bufFocusNotes:
		m.notesInput, cmd = m.notesInput.Update(msg)
	}
	return m, cmd
}

func (m BufferFormModel) View() string {
	title := "Add Buffer Entry"
	if m.isEditing {
		title = "Edit Buffer Entry"
	}

	monthLabel := monthName(m.month)
	monthLine := fmt.Sprintf("  %s  ←/→ to change", monthLabel)
	if m.focusIndex == bufFocusMonth {
		monthLine = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("▶ "+monthLabel+"  ") +
			helpStyle.Render("←/→ to change")
	}

	var s string
	s += titleStyle.Render(title) + "\n\n"
	s += inputStyle.Render(fmt.Sprintf("Year: %d", m.year)) + "\n\n"

	s += inputStyle.Render("Month") + "\n"
	s += monthLine + "\n\n"

	s += inputStyle.Render("Hours") + "\n"
	s += m.hoursInput.View() + "\n\n"

	s += inputStyle.Render("Notes") + "\n"
	s += m.notesInput.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(m.err.Error()) + "\n\n"
	}

	s += helpStyle.Render("Tab: next field • ←/→ (on Month): change month • Enter: save • Esc: cancel")
	return baseStyle.Render(s)
}
