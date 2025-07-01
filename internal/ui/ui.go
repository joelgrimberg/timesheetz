package ui

import (
	"timesheet/internal/ui/views"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InitializeUI sets up the main UI components
func InitializeUI() *tea.Program {
	// Create views
	timesheetView := views.NewTimesheetView()
	trainingView := views.NewTrainingView()
	settingsView := views.NewSettingsView()

	// Create the main model
	model := NewModel(timesheetView, trainingView, settingsView)

	// Create and return the program
	return tea.NewProgram(model, tea.WithAltScreen())
}

// Model represents the main UI model
type Model struct {
	currentView int
	views       []tea.Model
	viewport    viewport.Model
	ready       bool
}

// NewModel creates a new UI model
func NewModel(views ...tea.Model) Model {
	return Model{
		views: views,
		ready: true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.currentView = (m.currentView + 1) % len(m.views)
		case "shift+tab":
			m.currentView = (m.currentView - 1 + len(m.views)) % len(m.views)
		}
	}

	// Update the current view
	var cmd tea.Cmd
	m.views[m.currentView], cmd = m.views[m.currentView].Update(msg)
	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Get the current view's content
	content := m.views[m.currentView].View()

	// Create the tab bar
	tabs := []string{"Timesheet", "Training", "Settings"}
	tabBar := ""
	for i, tab := range tabs {
		if i == m.currentView {
			tabBar += lipgloss.NewStyle().Bold(true).Render(tab)
		} else {
			tabBar += tab
		}
		if i < len(tabs)-1 {
			tabBar += " | "
		}
	}

	// Combine the tab bar and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
	)
}
