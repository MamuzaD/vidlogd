package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor      = lipgloss.Color("#D32F2F")
	primaryBackground = lipgloss.Color("#A52A2A")
	white             = lipgloss.Color("#FFFFFF")
)

var (
	// main popup container style
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2).
			Margin(0)

	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 2)
)

// centerHorizontally centers content horizontally in the terminal
func centerHorizontally(content string, width int) string {
	return lipgloss.Place(width, lipgloss.Height(content), lipgloss.Center, lipgloss.Top, content)
}

// centerVertically centers content vertically in the terminal
func centerVertically(content string, height int) string {
	return lipgloss.Place(lipgloss.Width(content), height, lipgloss.Left, lipgloss.Center, content)
}

// centerBoth centers content both horizontally and vertically
func centerBoth(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
