package ui

import (
	"fmt"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/datalayer"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverviewKeyMap defines the keybindings for the overview view
type OverviewKeyMap struct {
	Left    key.Binding
	Right   key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
}

// DefaultOverviewKeyMap returns the default keybindings
func DefaultOverviewKeyMap() OverviewKeyMap {
	return OverviewKeyMap{
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
func (k OverviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Left,
		k.Right,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k OverviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Left,
			k.Right,
			k.HelpKey,
			k.Quit,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// OverviewModel represents the overview view
type OverviewModel struct {
	trainingHoursLeft int
	vacationHoursLeft int
	currentYear       int
	keys              OverviewKeyMap
	help              help.Model
	showHelp          bool
}

// ChangeOverviewYearMsg is used to change the year
type ChangeOverviewYearMsg struct {
	Year int
}

// Command to change the year
func ChangeOverviewYear(year int) tea.Cmd {
	return func() tea.Msg {
		return ChangeOverviewYearMsg{Year: year}
	}
}

// InitialOverviewModel creates a new overview model
func InitialOverviewModel() OverviewModel {
	currentYear := time.Now().Year()

	// Get config
	configFile, err := config.GetConfig()
	if err != nil {
		return OverviewModel{
			trainingHoursLeft: 0,
			vacationHoursLeft: 0,
			currentYear:       currentYear,
			keys:              DefaultOverviewKeyMap(),
			help:              help.New(),
			showHelp:          false,
		}
	}

	// Calculate training hours left
	dataLayer := datalayer.GetDataLayer()
	trainingEntries, err := dataLayer.GetTrainingEntriesForYear(currentYear)
	var totalTrainingHours int
	if err == nil {
		for _, entry := range trainingEntries {
			totalTrainingHours += entry.Training_hours
		}
	}
	trainingHoursLeft := configFile.TrainingHours.YearlyTarget - totalTrainingHours

	// Calculate vacation hours left (includes carryover)
	vacationSummary, err := dataLayer.GetVacationSummaryForYear(currentYear)
	var vacationHoursLeft int
	if err == nil {
		vacationHoursLeft = vacationSummary.RemainingTotal
	} else {
		vacationHoursLeft = 0
	}

	return OverviewModel{
		trainingHoursLeft: trainingHoursLeft,
		vacationHoursLeft: vacationHoursLeft,
		currentYear:       currentYear,
		keys:              DefaultOverviewKeyMap(),
		help:              help.New(),
		showHelp:          false,
	}
}

func (m OverviewModel) Init() tea.Cmd {
	return nil
}

func (m OverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ChangeOverviewYearMsg:
		// Update the current year in the model
		m.currentYear = msg.Year

		// Reload config to get the latest yearly targets
		configFile, err := config.GetConfig()
		if err != nil {
			return m, nil
		}

		// Calculate training hours left
		dataLayer := datalayer.GetDataLayer()
		trainingEntries, err := dataLayer.GetTrainingEntriesForYear(msg.Year)
		var totalTrainingHours int
		if err == nil {
			for _, entry := range trainingEntries {
				totalTrainingHours += entry.Training_hours
			}
		}
		trainingHoursLeft := configFile.TrainingHours.YearlyTarget - totalTrainingHours
		m.trainingHoursLeft = trainingHoursLeft

		// Calculate vacation hours left (includes carryover)
		vacationSummary, err := dataLayer.GetVacationSummaryForYear(msg.Year)
		if err == nil {
			m.vacationHoursLeft = vacationSummary.RemainingTotal
		} else {
			m.vacationHoursLeft = 0
		}

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Left):
			// Move to previous year
			return m, ChangeOverviewYear(m.currentYear - 1)
		case key.Matches(msg, m.keys.Right):
			// Move to next year
			return m, ChangeOverviewYear(m.currentYear + 1)
		}
	}

	return m, cmd
}

func (m OverviewModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ←/→, h/l: Change year\n  ?: Toggle help\n  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("←/→: Change year • ?: Help • q: Quit • </>: Tabs")
	}

	// Create the overview content
	content := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(2, 4).
		Render(
			fmt.Sprintf(
				"%s\n%s\n\n%s\n%s",
				lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Training Hours Remaining:"),
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")).Render(fmt.Sprintf("  %d hours", m.trainingHoursLeft)),
				lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("Vacation Hours Remaining:"),
				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")).Render(fmt.Sprintf("  %d hours", m.vacationHoursLeft)),
			),
		)

	return fmt.Sprintf(
		"%s\n%s%s",
		content,
		helpStyle.Render("←/→: Change year • <: Prev tab • >: Next tab • q: Quit"),
		helpView,
	)
}
