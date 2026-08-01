package ui

import "github.com/charmbracelet/lipgloss"

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
	footerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	weekendStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dimmer style for weekends
	yankedStyle  = lipgloss.NewStyle().
			Background(lipgloss.Color("#5F5FDF")). // Blue background
			Foreground(lipgloss.Color("255")).     // White text for contrast
			Bold(true)
	infoStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))             // Light blue for info text
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")) // Pink for table headers
	tableRowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))            // White for table rows
	statusBarStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240"))
	statusBarTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")) // Same as titleStyle but no margin
	statusMessageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))             // Green for status messages

	// Timesheet delta styles — used in every TimesheetModel.View() call.
	// Pre-built at package init to avoid allocating a new lipgloss.Style
	// on every frame.
	expectedLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	expectedValueStyle = lipgloss.NewStyle().Bold(true)
	deltaBehindStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	deltaAheadStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	deltaOnTargetStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78"))

	// Sync status styles — pre-built variants for the four sync states in
	// app.go's View. Selected by state; never constructed per-frame.
	syncStatusErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // red
	syncStatusSyncingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))            // amber
	syncStatusRecentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Bold(true)  // green
	syncStatusOldStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))            // dim

	// Overview view styles.
	overviewDimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	overviewLabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	overviewValueStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78"))
	overviewContentStyle   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(2, 4)
)
