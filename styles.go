package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	primaryColor      = lipgloss.Color("#D32F2F")
	primaryBackground = lipgloss.Color("#A52A2A")
	white             = lipgloss.Color("#FFFFFF")
	gray              = lipgloss.Color("#909090")
)

var (
	// main popup container style
	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// title style for `vidlogd`
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2).
			Margin(0, 0, 1, 0).
			AlignHorizontal(lipgloss.Center)

	// items in main menu
	menuItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	// normal form field style
	formFieldStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray)

	// focused form field
	formFieldFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryBackground)

	// form button style
	buttonStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	buttonStyleFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Background(primaryBackground).
				Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderBottom(true).
				Background(primaryColor).
				Foreground(white).
				Bold(true).
				Padding(0, 1)

	// table styles
	tableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(white).
			Padding(1, 2)

	tableSelectedRowStyle = lipgloss.NewStyle().
				Background(primaryBackground).
				Foreground(white).
				Bold(true)

	// log details styles
	reviewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Width(70)
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
