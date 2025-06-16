package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	red      = "#D32F2F"
	redBg    = "#A52A2A"
	blue     = "#1976D2"
	blueBg   = "#1565C0"
	green    = "#388E3C"
	greenBg  = "#2E7D32"
	purple   = "#7B1FA2"
	purpleBg = "#6A1B9A"
	orange   = "#F57C00"
	orangeBg = "#E65100"
	teal     = "#00796B"
	tealBg   = "#00695C"
	pink     = "#C2185B"
	pinkBg   = "#AD1457"

	primaryColor      = lipgloss.Color(red)
	primaryBackground = lipgloss.Color(redBg)
	white             = lipgloss.Color("#FFFFFF")
	gray              = lipgloss.Color("#909090")
)

// set theme colors and update all styles
func SetThemeColors(primary, primaryBg string) {
	primaryColor = lipgloss.Color(primary)
	primaryBackground = lipgloss.Color(primaryBg)

	initStyles()
}

func initStyles() {
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

	headerStyle = lipgloss.NewStyle().
		Foreground(white).
		Bold(true).
		BorderBottom(true).
		Width(15).
		AlignHorizontal(lipgloss.Left)

	starStyle = lipgloss.NewStyle().
		Foreground(primaryColor)

	modeStyle = lipgloss.NewStyle().
		Foreground(gray).
		Width(50).
		AlignHorizontal(lipgloss.Right)

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

	// description style for settings and other items
	descriptionStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Faint(true)

	// search box style
	searchStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1).
		Margin(0, 1).
		Height(1)
}

var (
	// main popup container style
	popupStyle lipgloss.Style

	// title style for `vidlogd`
	titleStyle lipgloss.Style

	// items in main menu
	menuItemStyle lipgloss.Style

	// form style
	headerStyle           lipgloss.Style
	formFieldStyle        lipgloss.Style
	formFieldFocusedStyle lipgloss.Style
	starStyle             lipgloss.Style
	modeStyle             lipgloss.Style

	// form button style
	buttonStyle        lipgloss.Style
	buttonStyleFocused lipgloss.Style

	// table styles
	tableStyle            lipgloss.Style
	tableHeaderStyle      lipgloss.Style
	tableSelectedRowStyle lipgloss.Style

	// log details styles
	reviewStyle lipgloss.Style

	// description style
	descriptionStyle lipgloss.Style

	// search box style
	searchStyle lipgloss.Style
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

func init() {
	initStyles()
}
