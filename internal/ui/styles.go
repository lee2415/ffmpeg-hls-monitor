package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#00ff88")
	secondaryColor = lipgloss.Color("#88aaff")
	errorColor     = lipgloss.Color("#ff6b6b")
	warningColor   = lipgloss.Color("#ffaa00")
	mutedColor     = lipgloss.Color("#666666")
	highlightColor = lipgloss.Color("#00ddff")
	
	// Base styles
	BaseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(mutedColor)
	
	// Header styles
	HeaderStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)
	
	TitleStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 1)
	
	// Status styles
	StatusRunningStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)
	
	StatusStoppedStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)
	
	StatusSelectedStyle = lipgloss.NewStyle().
		Background(highlightColor).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)
	
	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center)
	
	TableCellStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Align(lipgloss.Left)
	
	TableSelectedRowStyle = lipgloss.NewStyle().
		Background(highlightColor).
		Foreground(lipgloss.Color("#000000")).
		Bold(true)
	
	// Panel styles
	PanelStyle = BaseStyle.Copy().
		Padding(1, 2).
		Margin(1, 1)
	
	// Help styles
	HelpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Padding(1, 2)
)

// Status colors
func GetStatusColor(status string) lipgloss.Style {
	switch status {
	case "RUN", "Running":
		return StatusRunningStyle
	case "STOP", "Stopped":
		return StatusStoppedStyle
	default:
		return lipgloss.NewStyle().Foreground(warningColor)
	}
}

// Truncate text to fit width
func TruncateText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	if width < 4 {
		return text[:width]
	}
	return text[:width-3] + "..."
}