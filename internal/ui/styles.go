package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	Red      = "#D32F2F"
	RedBg    = "#A52A2A"
	Blue     = "#1976D2"
	BlueBg   = "#1565C0"
	Green    = "#388E3C"
	GreenBg  = "#2E7D32"
	Purple   = "#7B1FA2"
	PurpleBg = "#6A1B9A"
	Orange   = "#F57C00"
	OrangeBg = "#E65100"
	Teal     = "#00796B"
	TealBg   = "#00695C"
	Pink     = "#C2185B"
	PinkBg   = "#AD1457"

	PrimaryColor      = lipgloss.Color(Red)
	PrimaryBackground = lipgloss.Color(RedBg)
	White             = lipgloss.Color("#FFFFFF")
	Gray              = lipgloss.Color("#909090")
)

// set theme colors and update all styles
func SetThemeColors(primary, primaryBg string) {
	PrimaryColor = lipgloss.Color(primary)
	PrimaryBackground = lipgloss.Color(primaryBg)

	initStyles()
}

func initStyles() {
	// main popup container style
	PopupStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)

	// title style for `vidlogd`
	TitleStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2).
		Margin(0, 0, 1, 0).
		AlignHorizontal(lipgloss.Center)

	// items in main menu
	MenuItemStyle = lipgloss.NewStyle().
		Padding(0, 2)

	// normal form field style
	FormFieldStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Gray)

	// focused form field
	FormFieldFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryBackground)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(White).
		Bold(true).
		BorderBottom(true).
		Width(15).
		AlignHorizontal(lipgloss.Left)

	StarStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor)

	ModeStyle = lipgloss.NewStyle().
		Foreground(Gray).
		Width(50).
		AlignHorizontal(lipgloss.Right)

	// form button style
	ButtonStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	ButtonStyleFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Background(PrimaryBackground).
		Padding(0, 1)

	TableHeaderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		Background(PrimaryColor).
		Foreground(White).
		Bold(true).
		Padding(0, 1)

	// table styles
	TableStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(White).
		Padding(1, 2)

	TableSelectedRowStyle = lipgloss.NewStyle().
		Background(PrimaryBackground).
		Foreground(White).
		Bold(true)

	// log details styles
	ReviewStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(70)

	// description style for settings and other items
	DescriptionStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Faint(true)

	// search box style
	SearchStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Padding(0, 1).
		Margin(0, 1).
		Height(1)
}

var (
	// main popup container style
	PopupStyle lipgloss.Style

	// title style for `vidlogd`
	TitleStyle lipgloss.Style

	// items in main menu
	MenuItemStyle lipgloss.Style

	// form style
	HeaderStyle           lipgloss.Style
	FormFieldStyle        lipgloss.Style
	FormFieldFocusedStyle lipgloss.Style
	StarStyle             lipgloss.Style
	ModeStyle             lipgloss.Style

	// form button style
	ButtonStyle        lipgloss.Style
	ButtonStyleFocused lipgloss.Style

	// table styles
	TableStyle            lipgloss.Style
	TableHeaderStyle      lipgloss.Style
	TableSelectedRowStyle lipgloss.Style

	// log details styles
	ReviewStyle lipgloss.Style

	// description style
	DescriptionStyle lipgloss.Style

	// search box style
	SearchStyle lipgloss.Style
)

// centerHorizontally centers content horizontally in the terminal
func CenterHorizontally(content string, width int) string {
	return lipgloss.Place(width, lipgloss.Height(content), lipgloss.Center, lipgloss.Top, content)
}

// centerVertically centers content vertically in the terminal
func CenterVertically(content string, height int) string {
	return lipgloss.Place(lipgloss.Width(content), height, lipgloss.Left, lipgloss.Center, content)
}

// centerBoth centers content both horizontally and vertically
func CenterBoth(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func init() {
	initStyles()
}
