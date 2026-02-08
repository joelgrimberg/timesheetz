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

// CheckForUpdatesCmd returns a command that checks for updates
// Can be called from AppModel.Init() to check on startup
func CheckForUpdatesCmd() tea.Cmd {
	return func() tea.Msg {
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
	table             table.Model
	keys              ConfigKeyMap
	help              help.Model
	showHelp          bool
	showModeModal     bool
	modeModal         *ModeModalModel
	languageModal     *LanguageModalModel
	documentTypeModal *DocumentTypeModalModel
	boolModal         *BoolModalModel
	textModal         *TextInputModal
	overlay           *overlay.Model

	// Row indices for editable fields
	nameRowIdx             int
	companyRowIdx          int
	freeSpeechRowIdx       int
	startAPIServerRowIdx   int
	apiPortRowIdx          int
	apiModeRowIdx          int
	apiBaseURLRowIdx       int
	dbLocationRowIdx       int
	developmentModeRowIdx  int
	documentTypeRowIdx     int
	exportLangRowIdx       int
	sendToOthersRowIdx     int
	recipientEmailRowIdx   int
	senderEmailRowIdx      int
	replyToEmailRowIdx     int
	resendAPIKeyRowIdx     int
	trainingTargetRowIdx   int
	trainingCategoryRowIdx int
	vacationTargetRowIdx   int
	vacationCategoryRowIdx int

	// Update checking fields
	latestVersion   string
	updateAvailable bool
	checkingUpdate  bool
	updateCheckErr  error
}

// IsEditing returns true if a modal is active (text input or mode selection)
func (m ConfigModel) IsEditing() bool {
	return m.textModal != nil || m.overlay != nil || m.languageModal != nil || m.documentTypeModal != nil || m.boolModal != nil
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
		// Copy all row indices
		nameRowIdx:             indices.nameRowIdx,
		companyRowIdx:          indices.companyRowIdx,
		freeSpeechRowIdx:       indices.freeSpeechRowIdx,
		startAPIServerRowIdx:   indices.startAPIServerRowIdx,
		apiPortRowIdx:          indices.apiPortRowIdx,
		apiModeRowIdx:          indices.apiModeRowIdx,
		apiBaseURLRowIdx:       indices.apiBaseURLRowIdx,
		dbLocationRowIdx:       indices.dbLocationRowIdx,
		developmentModeRowIdx:  indices.developmentModeRowIdx,
		documentTypeRowIdx:     indices.documentTypeRowIdx,
		exportLangRowIdx:       indices.exportLangRowIdx,
		sendToOthersRowIdx:     indices.sendToOthersRowIdx,
		recipientEmailRowIdx:   indices.recipientEmailRowIdx,
		senderEmailRowIdx:      indices.senderEmailRowIdx,
		replyToEmailRowIdx:     indices.replyToEmailRowIdx,
		resendAPIKeyRowIdx:     indices.resendAPIKeyRowIdx,
		trainingTargetRowIdx:   indices.trainingTargetRowIdx,
		trainingCategoryRowIdx: indices.trainingCategoryRowIdx,
		vacationTargetRowIdx:   indices.vacationTargetRowIdx,
		vacationCategoryRowIdx: indices.vacationCategoryRowIdx,
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

// LanguageModalModel represents the modal for selecting export language
type LanguageModalModel struct {
	cursor int
	keys   ConfigKeyMap
}

// LanguageSelectedMsg is sent when a language is selected
type LanguageSelectedMsg struct {
	Language string
}

// LanguageCancelledMsg is sent when language modal is cancelled
type LanguageCancelledMsg struct{}

func InitialLanguageModalModel(currentLang string) *LanguageModalModel {
	if currentLang == "" {
		currentLang = "en"
	}
	langCursor := 0
	langs := []string{"en", "nl"}
	for i, lang := range langs {
		if lang == currentLang {
			langCursor = i
			break
		}
	}
	return &LanguageModalModel{
		cursor: langCursor,
		keys:   DefaultConfigKeyMap(),
	}
}

func (m LanguageModalModel) Init() tea.Cmd {
	return nil
}

func (m LanguageModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			return m, func() tea.Msg {
				return LanguageCancelledMsg{}
			}
		case key.Matches(msg, m.keys.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 1
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.cursor++
			if m.cursor > 1 {
				m.cursor = 0
			}
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			langs := []string{"en", "nl"}
			return m, func() tea.Msg {
				return LanguageSelectedMsg{Language: langs[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m LanguageModalModel) View() string {
	langs := []string{"en", "nl"}
	langDescriptions := []string{
		"English",
		"Nederlands (Dutch)",
	}

	var modalRows []string
	modalRows = append(modalRows, lipgloss.NewStyle().Bold(true).Render("Select Export Language:"))
	modalRows = append(modalRows, "")

	for i, lang := range langs {
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
		row := fmt.Sprintf("  %s - %s", style.Render(lang), langDescriptions[i])
		modalRows = append(modalRows, row)
	}

	modalRows = append(modalRows, "")
	modalRows = append(modalRows, lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: Select • Enter: Confirm • Esc: Cancel"))

	modalContent := lipgloss.JoinVertical(lipgloss.Left, modalRows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		Render(modalContent)
}

// DocumentTypeModalModel represents the modal for selecting document type
type DocumentTypeModalModel struct {
	cursor int
	keys   ConfigKeyMap
}

// DocumentTypeSelectedMsg is sent when a document type is selected
type DocumentTypeSelectedMsg struct {
	DocumentType string
}

// DocumentTypeCancelledMsg is sent when document type modal is cancelled
type DocumentTypeCancelledMsg struct{}

func InitialDocumentTypeModalModel(currentType string) *DocumentTypeModalModel {
	if currentType == "" {
		currentType = "excel"
	}
	typeCursor := 0
	types := []string{"excel", "pdf"}
	for i, t := range types {
		if t == currentType {
			typeCursor = i
			break
		}
	}
	return &DocumentTypeModalModel{
		cursor: typeCursor,
		keys:   DefaultConfigKeyMap(),
	}
}

func (m DocumentTypeModalModel) Init() tea.Cmd {
	return nil
}

func (m DocumentTypeModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			return m, func() tea.Msg {
				return DocumentTypeCancelledMsg{}
			}
		case key.Matches(msg, m.keys.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 1
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.cursor++
			if m.cursor > 1 {
				m.cursor = 0
			}
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			types := []string{"excel", "pdf"}
			return m, func() tea.Msg {
				return DocumentTypeSelectedMsg{DocumentType: types[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m DocumentTypeModalModel) View() string {
	types := []string{"excel", "pdf"}
	typeDescriptions := []string{
		"Excel spreadsheet (.xlsx)",
		"PDF document (.pdf)",
	}

	var modalRows []string
	modalRows = append(modalRows, lipgloss.NewStyle().Bold(true).Render("Select Document Type:"))
	modalRows = append(modalRows, "")

	for i, t := range types {
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
		row := fmt.Sprintf("  %s - %s", style.Render(t), typeDescriptions[i])
		modalRows = append(modalRows, row)
	}

	modalRows = append(modalRows, "")
	modalRows = append(modalRows, lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: Select • Enter: Confirm • Esc: Cancel"))

	modalContent := lipgloss.JoinVertical(lipgloss.Left, modalRows...)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(60).
		Render(modalContent)
}

// BoolModalModel represents the modal for toggling boolean values
type BoolModalModel struct {
	cursor    int
	fieldName string
	keys      ConfigKeyMap
}

// BoolSelectedMsg is sent when a boolean value is selected
type BoolSelectedMsg struct {
	FieldName string
	Value     bool
}

// BoolCancelledMsg is sent when bool modal is cancelled
type BoolCancelledMsg struct{}

func InitialBoolModalModel(fieldName string, currentValue bool) *BoolModalModel {
	cursor := 0
	if currentValue {
		cursor = 0
	} else {
		cursor = 1
	}
	return &BoolModalModel{
		cursor:    cursor,
		fieldName: fieldName,
		keys:      DefaultConfigKeyMap(),
	}
}

func (m BoolModalModel) Init() tea.Cmd {
	return nil
}

func (m BoolModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			return m, func() tea.Msg {
				return BoolCancelledMsg{}
			}
		case key.Matches(msg, m.keys.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 1
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			m.cursor++
			if m.cursor > 1 {
				m.cursor = 0
			}
			return m, nil
		case key.Matches(msg, m.keys.Enter):
			value := m.cursor == 0 // 0 = true, 1 = false
			return m, func() tea.Msg {
				return BoolSelectedMsg{FieldName: m.fieldName, Value: value}
			}
		}
	}
	return m, nil
}

func (m BoolModalModel) View() string {
	options := []string{"true", "false"}

	var modalRows []string
	modalRows = append(modalRows, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Set %s:", m.fieldName)))
	modalRows = append(modalRows, "")

	for i, opt := range options {
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
		modalRows = append(modalRows, fmt.Sprintf("  %s", style.Render(opt)))
	}

	modalRows = append(modalRows, "")
	modalRows = append(modalRows, lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("↑/↓: Select • Enter: Confirm • Esc: Cancel"))

	modalContent := lipgloss.JoinVertical(lipgloss.Left, modalRows...)

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

// maskPostgresURL masks the password in a PostgreSQL connection URL
func maskPostgresURL(url string) string {
	// URL format: postgres://user:password@host:port/db?params
	// We want to show: postgres://user:****@host:port/db
	if len(url) < 11 {
		return url
	}

	// Find the :// prefix
	prefixEnd := 0
	for i := 0; i < len(url)-2; i++ {
		if url[i:i+3] == "://" {
			prefixEnd = i + 3
			break
		}
	}
	if prefixEnd == 0 {
		return url
	}

	// Find the @ symbol (end of credentials)
	atIdx := -1
	for i := prefixEnd; i < len(url); i++ {
		if url[i] == '@' {
			atIdx = i
			break
		}
	}
	if atIdx == -1 {
		return url // No credentials in URL
	}

	// Find the : between user and password
	colonIdx := -1
	for i := prefixEnd; i < atIdx; i++ {
		if url[i] == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return url // No password in URL
	}

	// Reconstruct with masked password
	return url[:colonIdx+1] + "****" + url[atIdx:]
}

// configRowIndices holds the row indices for editable fields
type configRowIndices struct {
	nameRowIdx             int
	companyRowIdx          int
	freeSpeechRowIdx       int
	startAPIServerRowIdx   int
	apiPortRowIdx          int
	apiModeRowIdx          int
	apiBaseURLRowIdx       int
	dbLocationRowIdx       int
	developmentModeRowIdx  int
	documentTypeRowIdx     int
	exportLangRowIdx       int
	sendToOthersRowIdx     int
	recipientEmailRowIdx   int
	senderEmailRowIdx      int
	replyToEmailRowIdx     int
	resendAPIKeyRowIdx     int
	trainingTargetRowIdx   int
	trainingCategoryRowIdx int
	vacationTargetRowIdx   int
	vacationCategoryRowIdx int
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
	indices.freeSpeechRowIdx = len(rows)
	rows = append(rows, table.Row{"  Free Speech", cfg.FreeSpeech})

	// API Server Configuration
	rows = append(rows, table.Row{"API Server", ""})
	indices.startAPIServerRowIdx = len(rows)
	rows = append(rows, table.Row{"  Start API Server", fmt.Sprintf("%v", cfg.StartAPIServer)})
	indices.apiPortRowIdx = len(rows)
	rows = append(rows, table.Row{"  API Port", strconv.Itoa(cfg.APIPort)})

	// API Client Configuration
	rows = append(rows, table.Row{"API Client", ""})
	indices.apiModeRowIdx = len(rows)
	if cfg.APIMode == "" {
		rows = append(rows, table.Row{"  API Mode", "local (default)"})
	} else {
		rows = append(rows, table.Row{"  API Mode", cfg.APIMode})
	}
	indices.apiBaseURLRowIdx = len(rows)
	if cfg.APIBaseURL == "" {
		rows = append(rows, table.Row{"  API Base URL", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  API Base URL", cfg.APIBaseURL})
	}

	// Database Configuration
	rows = append(rows, table.Row{"Database", ""})
	// Show database type (read-only, set via CLI/env)
	dbType := config.GetDBType()
	if dbType == "postgres" {
		rows = append(rows, table.Row{"  DB Type", "PostgreSQL"})
	} else {
		rows = append(rows, table.Row{"  DB Type", "SQLite"})
	}
	indices.dbLocationRowIdx = len(rows)
	if dbType == "postgres" {
		// For PostgreSQL, show connection info (masked)
		postgresURL := config.GetPostgresURL()
		if postgresURL != "" {
			// Mask the password in the URL for display
			rows = append(rows, table.Row{"  Connection", maskPostgresURL(postgresURL)})
		} else {
			rows = append(rows, table.Row{"  Connection", "(not configured)"})
		}
	} else {
		// For SQLite, show file location
		if cfg.DBLocation == "" {
			rows = append(rows, table.Row{"  DB Location", "(default)"})
		} else {
			rows = append(rows, table.Row{"  DB Location", cfg.DBLocation})
		}
	}

	// Development Settings
	rows = append(rows, table.Row{"Development", ""})
	indices.developmentModeRowIdx = len(rows)
	rows = append(rows, table.Row{"  Development Mode", fmt.Sprintf("%v", cfg.DevelopmentMode)})

	// Document Settings
	rows = append(rows, table.Row{"Document", ""})
	indices.documentTypeRowIdx = len(rows)
	docType := cfg.SendDocumentType
	if docType == "" {
		docType = "excel (default)"
	}
	rows = append(rows, table.Row{"  Send Document Type", docType})
	indices.exportLangRowIdx = len(rows)
	exportLang := cfg.ExportLanguage
	if exportLang == "" {
		exportLang = "en (default)"
	}
	rows = append(rows, table.Row{"  Export Language", exportLang})

	// Email Configuration
	rows = append(rows, table.Row{"Email", ""})
	indices.sendToOthersRowIdx = len(rows)
	rows = append(rows, table.Row{"  Send To Others", fmt.Sprintf("%v", cfg.SendToOthers)})
	indices.recipientEmailRowIdx = len(rows)
	if cfg.RecipientEmail == "" {
		rows = append(rows, table.Row{"  Recipient Email", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  Recipient Email", cfg.RecipientEmail})
	}
	indices.senderEmailRowIdx = len(rows)
	if cfg.SenderEmail == "" {
		rows = append(rows, table.Row{"  Sender Email", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  Sender Email", cfg.SenderEmail})
	}
	indices.replyToEmailRowIdx = len(rows)
	if cfg.ReplyToEmail == "" {
		rows = append(rows, table.Row{"  Reply To Email", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  Reply To Email", cfg.ReplyToEmail})
	}
	indices.resendAPIKeyRowIdx = len(rows)
	if cfg.ResendAPIKey != "" {
		// Mask API key for security
		maskedKey := maskAPIKey(cfg.ResendAPIKey)
		rows = append(rows, table.Row{"  Resend API Key", maskedKey})
	} else {
		rows = append(rows, table.Row{"  Resend API Key", "(not set)"})
	}

	// Training Hours Configuration
	rows = append(rows, table.Row{"Training Hours", ""})
	indices.trainingTargetRowIdx = len(rows)
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.TrainingHours.YearlyTarget)})
	indices.trainingCategoryRowIdx = len(rows)
	if cfg.TrainingHours.Category == "" {
		rows = append(rows, table.Row{"  Category", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  Category", cfg.TrainingHours.Category})
	}

	// Vacation Hours Configuration
	rows = append(rows, table.Row{"Vacation Hours", ""})
	indices.vacationTargetRowIdx = len(rows)
	rows = append(rows, table.Row{"  Yearly Target", strconv.Itoa(cfg.VacationHours.YearlyTarget)})
	indices.vacationCategoryRowIdx = len(rows)
	if cfg.VacationHours.Category == "" {
		rows = append(rows, table.Row{"  Category", "(not set)"})
	} else {
		rows = append(rows, table.Row{"  Category", cfg.VacationHours.Category})
	}

	return rows, indices
}

func (m ConfigModel) Init() tea.Cmd {
	// Trigger update check when config tab loads
	return CheckForUpdatesCmd()
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
				case "Free Speech":
					cfg.FreeSpeech = saveMsg.Value
				case "API Port":
					if port, err := strconv.Atoi(saveMsg.Value); err == nil {
						cfg.APIPort = port
					}
				case "API Base URL":
					cfg.APIBaseURL = saveMsg.Value
				case "DB Location":
					cfg.DBLocation = saveMsg.Value
				case "Recipient Email":
					cfg.RecipientEmail = saveMsg.Value
				case "Sender Email":
					cfg.SenderEmail = saveMsg.Value
				case "Reply To Email":
					cfg.ReplyToEmail = saveMsg.Value
				case "Resend API Key":
					cfg.ResendAPIKey = saveMsg.Value
				case "Training Yearly Target":
					if target, err := strconv.Atoi(saveMsg.Value); err == nil {
						cfg.TrainingHours.YearlyTarget = target
					}
				case "Training Category":
					cfg.TrainingHours.Category = saveMsg.Value
				case "Vacation Yearly Target":
					if target, err := strconv.Atoi(saveMsg.Value); err == nil {
						cfg.VacationHours.YearlyTarget = target
					}
				case "Vacation Category":
					cfg.VacationHours.Category = saveMsg.Value
				}
				config.SaveConfig(cfg)
				// Rebuild the table with updated values
				rows, _ := m.buildTableRows(&cfg)
				m.table.SetRows(rows)
			}
			m.textModal = nil
			return m, SetStatus("Configuration saved")
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

	// Handle language modal updates (using overlay)
	if m.overlay != nil && m.languageModal != nil {
		updatedForeground, foregroundCmd := m.languageModal.Update(msg)
		if updatedModal, ok := updatedForeground.(LanguageModalModel); ok {
			m.languageModal = &updatedModal
		} else if updatedModalPtr, ok := updatedForeground.(*LanguageModalModel); ok {
			m.languageModal = updatedModalPtr
		}

		m.overlay = overlay.New(
			m.languageModal,
			m,
			overlay.Center,
			overlay.Center,
			0,
			0,
		)

		return m, foregroundCmd
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

	// Handle document type modal updates (using overlay)
	if m.overlay != nil && m.documentTypeModal != nil {
		updatedForeground, foregroundCmd := m.documentTypeModal.Update(msg)
		if updatedModal, ok := updatedForeground.(DocumentTypeModalModel); ok {
			m.documentTypeModal = &updatedModal
		} else if updatedModalPtr, ok := updatedForeground.(*DocumentTypeModalModel); ok {
			m.documentTypeModal = updatedModalPtr
		}

		m.overlay = overlay.New(
			m.documentTypeModal,
			m,
			overlay.Center,
			overlay.Center,
			0,
			0,
		)

		return m, foregroundCmd
	}

	// Handle bool modal updates (using overlay)
	if m.overlay != nil && m.boolModal != nil {
		updatedForeground, foregroundCmd := m.boolModal.Update(msg)
		if updatedModal, ok := updatedForeground.(BoolModalModel); ok {
			m.boolModal = &updatedModal
		} else if updatedModalPtr, ok := updatedForeground.(*BoolModalModel); ok {
			m.boolModal = updatedModalPtr
		}

		m.overlay = overlay.New(
			m.boolModal,
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

			// Text input fields
			if cursor == m.nameRowIdx {
				m.textModal = InitialTextInputModal("Name", cfg.Name)
				return m, m.textModal.Init()
			}
			if cursor == m.companyRowIdx {
				m.textModal = InitialTextInputModal("Company Name", cfg.CompanyName)
				return m, m.textModal.Init()
			}
			if cursor == m.freeSpeechRowIdx {
				m.textModal = InitialTextInputModal("Free Speech", cfg.FreeSpeech)
				return m, m.textModal.Init()
			}
			if cursor == m.apiPortRowIdx {
				m.textModal = InitialTextInputModal("API Port", strconv.Itoa(cfg.APIPort))
				return m, m.textModal.Init()
			}
			if cursor == m.apiBaseURLRowIdx {
				m.textModal = InitialTextInputModal("API Base URL", cfg.APIBaseURL)
				return m, m.textModal.Init()
			}
			if cursor == m.dbLocationRowIdx {
				m.textModal = InitialTextInputModal("DB Location", cfg.DBLocation)
				return m, m.textModal.Init()
			}
			if cursor == m.recipientEmailRowIdx {
				m.textModal = InitialTextInputModal("Recipient Email", cfg.RecipientEmail)
				return m, m.textModal.Init()
			}
			if cursor == m.senderEmailRowIdx {
				m.textModal = InitialTextInputModal("Sender Email", cfg.SenderEmail)
				return m, m.textModal.Init()
			}
			if cursor == m.replyToEmailRowIdx {
				m.textModal = InitialTextInputModal("Reply To Email", cfg.ReplyToEmail)
				return m, m.textModal.Init()
			}
			if cursor == m.resendAPIKeyRowIdx {
				m.textModal = InitialTextInputModal("Resend API Key", cfg.ResendAPIKey)
				return m, m.textModal.Init()
			}
			if cursor == m.trainingTargetRowIdx {
				m.textModal = InitialTextInputModal("Training Yearly Target", strconv.Itoa(cfg.TrainingHours.YearlyTarget))
				return m, m.textModal.Init()
			}
			if cursor == m.trainingCategoryRowIdx {
				m.textModal = InitialTextInputModal("Training Category", cfg.TrainingHours.Category)
				return m, m.textModal.Init()
			}
			if cursor == m.vacationTargetRowIdx {
				m.textModal = InitialTextInputModal("Vacation Yearly Target", strconv.Itoa(cfg.VacationHours.YearlyTarget))
				return m, m.textModal.Init()
			}
			if cursor == m.vacationCategoryRowIdx {
				m.textModal = InitialTextInputModal("Vacation Category", cfg.VacationHours.Category)
				return m, m.textModal.Init()
			}

			// Boolean toggle fields
			if cursor == m.startAPIServerRowIdx {
				m.boolModal = InitialBoolModalModel("Start API Server", cfg.StartAPIServer)
				m.overlay = overlay.New(m.boolModal, m, overlay.Center, overlay.Center, 0, 0)
				return m, nil
			}
			if cursor == m.developmentModeRowIdx {
				m.boolModal = InitialBoolModalModel("Development Mode", cfg.DevelopmentMode)
				m.overlay = overlay.New(m.boolModal, m, overlay.Center, overlay.Center, 0, 0)
				return m, nil
			}
			if cursor == m.sendToOthersRowIdx {
				m.boolModal = InitialBoolModalModel("Send To Others", cfg.SendToOthers)
				m.overlay = overlay.New(m.boolModal, m, overlay.Center, overlay.Center, 0, 0)
				return m, nil
			}

			// Dropdown fields
			if cursor == m.exportLangRowIdx {
				m.languageModal = InitialLanguageModalModel(cfg.ExportLanguage)
				m.overlay = overlay.New(m.languageModal, m, overlay.Center, overlay.Center, 0, 0)
				return m, nil
			}
			if cursor == m.documentTypeRowIdx {
				m.documentTypeModal = InitialDocumentTypeModalModel(cfg.SendDocumentType)
				m.overlay = overlay.New(m.documentTypeModal, m, overlay.Center, overlay.Center, 0, 0)
				return m, nil
			}
			if cursor == m.apiModeRowIdx {
				m.modeModal = InitialModeModalModel(cfg.APIMode)
				m.overlay = overlay.New(m.modeModal, m, overlay.Center, overlay.Center, 0, 0)
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
		return m.textModal.View()
	}

	content := fmt.Sprintf(
		"%s\n%s%s",
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
