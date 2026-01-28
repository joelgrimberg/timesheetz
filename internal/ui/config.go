package ui

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"timesheet/internal/config"
	"timesheet/internal/updater"
	"timesheet/internal/version"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rmhubbert/bubbletea-overlay"
)

// Package-level cache for update check results (1-hour cache to avoid rate limiting)
var (
	lastCheckTime time.Time
	cachedResult  updateCheckResultMsg
	checkMutex    sync.Mutex
)

// updateCheckResultMsg contains the result of checking for updates
type updateCheckResultMsg struct {
	latestVersion   string
	updateAvailable bool
	err             error
}

// ConfigKeyMap defines the keybindings for the config view
type ConfigKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	HelpKey key.Binding
	Quit    key.Binding
	PrevTab key.Binding
	NextTab key.Binding
	Enter   key.Binding
	Escape  key.Binding
}

// DefaultConfigKeyMap returns the default keybindings
func DefaultConfigKeyMap() ConfigKeyMap {
	return ConfigKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
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
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k ConfigKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.HelpKey,
		k.Quit,
	}
}

// FullHelp returns keybindings for the expanded help view
func (k ConfigKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.HelpKey,
			k.Quit,
		},
		{
			k.PrevTab,
			k.NextTab,
		},
	}
}

// ModeModalModel represents the modal for selecting API mode
type ModeModalModel struct {
	cursor int
	keys   ConfigKeyMap
}

// TextInputModal represents a modal for editing text fields
type TextInputModal struct {
	textInput textinput.Model
	fieldName string
	keys      ConfigKeyMap
}

// InitialTextInputModal creates a new text input modal
func InitialTextInputModal(fieldName, currentValue string) *TextInputModal {
	ti := textinput.New()
	ti.SetValue(currentValue)
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50
	ti.Prompt = "> "
	return &TextInputModal{
		textInput: ti,
		fieldName: fieldName,
		keys:      DefaultConfigKeyMap(),
	}
}

func (m TextInputModal) Init() tea.Cmd {
	return textinput.Blink
}

// TextInputSavedMsg is sent when text input is confirmed
type TextInputSavedMsg struct {
	FieldName string
	Value     string
}

// TextInputCancelledMsg is sent when text input modal is cancelled
type TextInputCancelledMsg struct{}

func (m TextInputModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg {
				return TextInputCancelledMsg{}
			}
		case tea.KeyEnter:
			return m, func() tea.Msg {
				return TextInputSavedMsg{
					FieldName: m.fieldName,
					Value:     m.textInput.Value(),
				}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputModal) View() string {
	var modalRows []string
	modalRows = append(modalRows, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Edit %s:", m.fieldName)))
	modalRows = append(modalRows, "")
	modalRows = append(modalRows, m.textInput.View())
	modalRows = append(modalRows, "")
	modalRows = append(modalRows, lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Enter: Save • Esc: Cancel"))

	modalContent := lipgloss.JoinVertical(lipgloss.Left, modalRows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		Render(modalContent)
}

// ConfigModel represents the configuration view
type ConfigModel struct {
	table         table.Model
	keys          ConfigKeyMap
	help          help.Model
	showHelp      bool
	showModeModal bool
	modeModal     *ModeModalModel
	textModal     *TextInputModal
	overlay       *overlay.Model
	apiModeRowIdx int // Index of the "API Mode" row in the table
	nameRowIdx    int // Index of the "Name" row in the table
	companyRowIdx int // Index of the "Company Name" row in the table

	// Update checking fields
	latestVersion   string
	updateAvailable bool
	checkingUpdate  bool
	updateCheckErr  error
}

// IsEditing returns true if a modal is active (text input or mode selection)
func (m ConfigModel) IsEditing() bool {
	return m.textModal != nil || m.overlay != nil
}

// InitialConfigModel creates a new config model
func InitialConfigModel() ConfigModel {
	// Create columns for the table
	columns := []table.Column{
		{Title: "Field", Width: 30},
		{Title: "Value", Width: 50},
	}

	// Create the table
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(25),
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
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("252"))
	t.SetStyles(s)

	// Get config
	cfg, err := config.GetConfig()
	if err != nil {
		// Return empty model if config can't be loaded
		return ConfigModel{
			table:    t,
			keys:     DefaultConfigKeyMap(),
			help:     help.New(),
			showHelp: false,
		}
	}

	// Convert config to table rows using buildTableRows
	// We need to create a temporary model to call buildTableRows
	tempModel := ConfigModel{}
	rows, indices := tempModel.buildTableRows(&cfg)

	t.SetRows(rows)

	// Select the first row by default
	if len(rows) > 0 {
		t.SetCursor(0)
	}

	return ConfigModel{
		table:         t,
		keys:          DefaultConfigKeyMap(),
		help:          help.New(),
		showHelp:      false,
		showModeModal: false,
		modeModal:     nil,
		textModal:     nil,
		overlay:       nil,
		apiModeRowIdx: indices.apiModeRowIdx,
		nameRowIdx:    indices.nameRowIdx,
		companyRowIdx: indices.companyRowIdx,
	}
}

// InitialModeModalModel creates a new mode modal model
func InitialModeModalModel(currentMode string) *ModeModalModel {
	// Determine current mode index for modal cursor
	if currentMode == "" {
		currentMode = "local"
	}
	modeCursor := 0
	modes := []string{"local", "dual", "remote"}
	for i, mode := range modes {
		if mode == currentMode {
			modeCursor = i
			break
		}
	}

	return &ModeModalModel{
		cursor: modeCursor,
		keys:   DefaultConfigKeyMap(),
	}
}

func (m ModeModalModel) Init() tea.Cmd {
	return nil
}

// ModeSelectedMsg is sent when a mode is selected
type ModeSelectedMsg struct {
	Mode string
}

// ModeCancelledMsg is sent when modal is cancelled
type ModeCancelledMsg struct{}

func (m ModeModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			// Close modal without saving - send cancel message
			return m, func() tea.Msg {
				return ModeCancelledMsg{}
			}
		case key.Matches(msg, m.keys.Up):
			// Move cursor up in modal
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 2 // Wrap to bottom (remote)
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			// Move cursor down in modal
			m.cursor++
			if m.cursor > 2 {
				m.cursor = 0 // Wrap to top (local)
			}
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			// Save selected mode - send selection message
			modes := []string{"local", "dual", "remote"}
			selectedMode := modes[m.cursor]
			return m, func() tea.Msg {
				return ModeSelectedMsg{Mode: selectedMode}
			}
		}
	}
	return m, nil
}

func (m ModeModalModel) View() string {
	modes := []string{"local", "dual", "remote"}
	modeDescriptions := []string{
		"Use local database only",
		"Use both local DB and remote API (for validation)",
		"Use remote API only",
	}

	// Build modal content
	var modalRows []string
	modalRows = append(modalRows, lipgloss.NewStyle().Bold(true).Render("Select API Mode:"))
	modalRows = append(modalRows, "")

	for i, mode := range modes {
		var style lipgloss.Style
		if i == m.cursor {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)
		}
		row := fmt.Sprintf("  %s - %s", style.Render(mode), modeDescriptions[i])
		modalRows = append(modalRows, row)
	}

	modalRows = append(modalRows, "")
	modalRows = append(modalRows, lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: Select • Enter: Confirm • Esc: Cancel"))

	modalContent := lipgloss.JoinVertical(lipgloss.Left, modalRows...)

	// Style the modal with border
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		Render(modalContent)
}

// maskAPIKey masks the API key showing only first few and last few characters
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// configRowIndices holds the row indices for editable fields
type configRowIndices struct {
	nameRowIdx    int
	companyRowIdx int
	apiModeRowIdx int
}

// buildTableRows builds the configuration table rows with update info
func (m ConfigModel) buildTableRows(cfg *config.Config) ([]table.Row, configRowIndices) {
	var rows []table.Row
	var indices configRowIndices

	// Version Information with update check
	versionValue := version.Version
	if m.updateAvailable && m.latestVersion != "" {
		// Add inline badge showing available update
		badge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")). // Green color
			Bold(true).
			Render(fmt.Sprintf(" (%s available)", m.latestVersion))
		versionValue = versionValue + badge
	} else if m.checkingUpdate {
		// Show loading indicator
		spinner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(" (checking...)")
		versionValue = versionValue + spinner
	}
	rows = append(rows, table.Row{"Version", versionValue})
	rows = append(rows, table.Row{"", ""}) // Empty row for spacing

	// User Information
	rows = append(rows, table.Row{"User Information", ""})
	indices.nameRowIdx = len(rows)
	rows = append(rows, table.Row{"  Name", cfg.Name})
	indices.companyRowIdx = len(rows)
	rows = append(rows, table.Row{"  Company Name", cfg.CompanyName})
	rows = append(rows, table.Row{"  Free Speech", cfg.FreeSpeech})

	// API Server Configuration
	rows = append(rows, table.Row{"API Server", ""})
	rows = append(rows, table.Row{"  Start API Server", fmt.Sprintf("%v", cfg.StartAPIServer)})
	rows = append(rows, table.Row{"  API Port", strconv.Itoa(cfg.APIPort)})

	// API Client Configuration
	rows = append(rows, table.Row{"API Client", ""})
	indices.apiModeRowIdx = len(rows)
	if cfg.APIMode == "" {
		rows = append(rows, table.Row{"  API Mode", "local (default)"})
	} else {
		rows = append(rows, table.Row{"  API Mode", cfg.APIMode})
	}
	if cfg.APIBaseURL == "" {
		rows = append(rows, table.Row{"  API Base URL", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  API Base URL", cfg.APIBaseURL})
	}

	// Database Location
	rows = append(rows, table.Row{"Database", ""})
	if cfg.DBLocation == "" {
		rows = append(rows, table.Row{"  DB Location", "(default)"})
	} else {
		rows = append(rows, table.Row{"  DB Location", cfg.DBLocation})
	}

	// Development Settings
	rows = append(rows, table.Row{"Development", ""})
	rows = append(rows, table.Row{"  Development Mode", fmt.Sprintf("%v", cfg.DevelopmentMode)})

	// Document Settings
	rows = append(rows, table.Row{"Document", ""})
	rows = append(rows, table.Row{"  Send Document Type", cfg.SendDocumentType})

	// Email Configuration
	rows = append(rows, table.Row{"Email", ""})
	rows = append(rows, table.Row{"  Send To Others", fmt.Sprintf("%v", cfg.SendToOthers)})
	rows = append(rows, table.Row{"  Recipient Email", cfg.RecipientEmail})
	rows = append(rows, table.Row{"  Sender Email", cfg.SenderEmail})
	rows = append(rows, table.Row{"  Reply To Email", cfg.ReplyToEmail})
	if cfg.ResendAPIKey != "" {
		// Mask API key for security
		maskedKey := maskAPIKey(cfg.ResendAPIKey)
		rows = append(rows, table.Row{"  Resend API Key", maskedKey})
	} else {
		rows = append(rows, table.Row{"  Resend API Key", "(not set)"})
	}

	// Training Hours Configuration
	rows = append(rows, table.Row{"Training Hours", ""})
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.TrainingHours.YearlyTarget)})
	rows = append(rows, table.Row{"  Category", cfg.TrainingHours.Category})

	// Vacation Hours Configuration
	rows = append(rows, table.Row{"Vacation Hours", ""})
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.VacationHours.YearlyTarget)})
	rows = append(rows, table.Row{"  Category", cfg.VacationHours.Category})

	return rows, indices
}

// checkForUpdates checks GitHub for new releases
func (m ConfigModel) checkForUpdates() tea.Msg {
	checkMutex.Lock()
	defer checkMutex.Unlock()

	// Return cached result if checked within last hour
	if time.Since(lastCheckTime) < time.Hour && cachedResult.latestVersion != "" {
		return cachedResult
	}

	// Perform actual check
	checker := updater.NewUpdateChecker("joelgrimberg", "timesheetz")
	latest, available, err := checker.CheckForUpdate(version.Version)

	result := updateCheckResultMsg{
		latestVersion:   latest,
		updateAvailable: available,
		err:             err,
	}

	// Cache successful results
	if err == nil {
		cachedResult = result
		lastCheckTime = time.Now()
	}

	return result
}

func (m ConfigModel) Init() tea.Cmd {
	// Trigger update check when config tab loads
	return m.checkForUpdates
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle text input modal FIRST - capture all input when modal is active
	if m.textModal != nil {
		// Check for save/cancel messages
		if saveMsg, ok := msg.(TextInputSavedMsg); ok {
			cfg, err := config.GetConfig()
			if err == nil {
				switch saveMsg.FieldName {
				case "Name":
					cfg.Name = saveMsg.Value
				case "Company Name":
					cfg.CompanyName = saveMsg.Value
				}
				config.SaveConfig(cfg)
				// Rebuild the table with updated values
				rows, _ := m.buildTableRows(&cfg)
				m.table.SetRows(rows)
			}
			m.textModal = nil
			return m, nil
		}

		if _, ok := msg.(TextInputCancelledMsg); ok {
			m.textModal = nil
			return m, nil
		}

		// Pass all other messages to the text modal
		updatedForeground, foregroundCmd := m.textModal.Update(msg)
		if updatedModal, ok := updatedForeground.(TextInputModal); ok {
			m.textModal = &updatedModal
		} else if updatedModalPtr, ok := updatedForeground.(*TextInputModal); ok {
			m.textModal = updatedModalPtr
		}
		return m, foregroundCmd
	}

	// Handle update check result
	if resultMsg, ok := msg.(updateCheckResultMsg); ok {
		m.checkingUpdate = false
		if resultMsg.err != nil {
			// Silent failure - just store error for debugging
			m.updateCheckErr = resultMsg.err
		} else {
			m.latestVersion = resultMsg.latestVersion
			m.updateAvailable = resultMsg.updateAvailable
		}
		// Rebuild the table with new version string
		cfg, err := config.GetConfig()
		if err == nil {
			rows, _ := m.buildTableRows(&cfg)
			m.table.SetRows(rows)
		}
		return m, nil
	}

	// Handle mode modal updates (using overlay)
	if m.overlay != nil && m.modeModal != nil {
		updatedForeground, foregroundCmd := m.modeModal.Update(msg)
		if updatedModal, ok := updatedForeground.(ModeModalModel); ok {
			m.modeModal = &updatedModal
		} else if updatedModalPtr, ok := updatedForeground.(*ModeModalModel); ok {
			m.modeModal = updatedModalPtr
		}

		// Recreate overlay with updated modal
		m.overlay = overlay.New(
			m.modeModal,
			m,
			overlay.Center,
			overlay.Center,
			0,
			0,
		)

		return m, foregroundCmd
	}

	// Normal table interactions
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.HelpKey):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			cursor := m.table.Cursor()
			cfg, err := config.GetConfig()
			if err != nil {
				return m, nil
			}

			// Check if we're on the Name row
			if cursor == m.nameRowIdx {
				m.textModal = InitialTextInputModal("Name", cfg.Name)
				return m, m.textModal.Init()
			}

			// Check if we're on the Company Name row
			if cursor == m.companyRowIdx {
				m.textModal = InitialTextInputModal("Company Name", cfg.CompanyName)
				return m, m.textModal.Init()
			}

			// Check if we're on the API Mode row
			if cursor == m.apiModeRowIdx {
				currentMode := cfg.APIMode

				// Create modal
				m.modeModal = InitialModeModalModel(currentMode)

				// Create overlay - pass m (value) not &m (pointer) as background
				// The overlay will handle the model updates
				m.overlay = overlay.New(
					m.modeModal,
					m,
					overlay.Center,
					overlay.Center,
					0,
					0,
				)
				m.showModeModal = true
				return m, nil
			}
		case key.Matches(msg, m.keys.Down):
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		case key.Matches(msg, m.keys.Up):
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
	}

	// Handle other table updates
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m ConfigModel) View() string {
	var helpView string
	if m.showHelp {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Navigation:\n  ↑/↓, k/j: Move up/down\n  ?: Toggle help\n  q: Quit\n\nTabs:\n  <: Previous tab\n  >: Next tab")
	} else {
		helpView = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("↑/↓: Navigate • Enter: Edit • ?: Help • q: Quit • </>: Tabs")
	}

	// If text modal is active, show only the modal
	if m.textModal != nil {
		return fmt.Sprintf(
			"%s\n\n%s",
			titleStyle.Render("Configuration"),
			m.textModal.View(),
		)
	}

	content := fmt.Sprintf(
		"%s\n%s\n%s%s",
		titleStyle.Render("Configuration"),
		lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Render(m.table.View()),
		helpView,
		m.help.View(m.keys),
	)

	// If overlay is active (mode modal), use it to render
	if m.overlay != nil {
		return m.overlay.View()
	}

	return content
}

