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
)
