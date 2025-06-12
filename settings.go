package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// global settings
var Settings AppSettings

type SettingType int

const (
	VimMotionsToggle SettingType = iota
	ThemeSelector
)

func getDefaultSettings() AppSettings {
	return AppSettings{
		VimMotions: true,
		Theme:      "red",
	}
}

type SettingItem struct {
	settingType SettingType
	title       string
	description string
	value       string
	options     []string // for dropdown-style settings
}

// necessary for list
type SettingItemDelegate struct{}

func (i SettingItem) FilterValue() string                               { return i.title }
func (d SettingItemDelegate) Height() int                               { return 2 }
func (d SettingItemDelegate) Spacing() int                              { return 1 }
func (d SettingItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d SettingItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SettingItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var titleStyle, valueStyle lipgloss.Style
	if isSelected {
		titleStyle = menuItemStyle.Background(primaryColor).Foreground(white)
		valueStyle = menuItemStyle.Background(primaryColor).Foreground(white)
	} else {
		titleStyle = menuItemStyle
		valueStyle = menuItemStyle.Foreground(gray)
	}

	title := titleStyle.Render(i.title)
	value := valueStyle.Render(fmt.Sprintf("[%s]", i.value))

	line1 := title + " " + value
	line2 := descriptionStyle.Render(i.description)
	content := line1 + "\n" + line2
	fmt.Fprint(w, content)
}

type SettingsModel struct {
	list list.Model
}

func NewSettingsModel() SettingsModel {
	Settings = loadSettings()

	UpdateKeyMap()

	items := []list.Item{
		SettingItem{
			settingType: VimMotionsToggle,
			title:       "Vim Motions",
			description: "enable vim-style keyboard navigation",
			value:       getBoolString(Settings.VimMotions),
			options:     []string{"enabled", "disabled"},
		},
		SettingItem{
			settingType: ThemeSelector,
			title:       "Theme",
			description: "color theme for the vidlogd",
			value:       Settings.Theme,
			options:     []string{"red", "blue", "green", "purple", "orange", "teal", "pink"},
		},
	}

	const defaultWidth = 40
	const listHeight = 16

	l := list.New(items, SettingItemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(true)

	return SettingsModel{
		list: l,
	}
}

func getBoolString(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, GlobalKeyMap.Back) {
			return m, func() tea.Msg {
				return NavigateMsg{View: MainMenuView}
			}
		}
		if key.Matches(msg, GlobalKeyMap.Select, GlobalKeyMap.Right) {
			return m.cycleSetting()
		}
		if key.Matches(msg, GlobalKeyMap.Left) {
			return m.cycleSetting()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SettingsModel) cycleSetting() (SettingsModel, tea.Cmd) {
	selectedItem, ok := m.list.SelectedItem().(SettingItem)
	if !ok {
		return m, nil
	}

	currentIndex := 0
	for i, option := range selectedItem.options {
		if option == selectedItem.value {
			currentIndex = i
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(selectedItem.options)
	newValue := selectedItem.options[nextIndex]

	// update the app settings
	switch selectedItem.settingType {
	case VimMotionsToggle:
		Settings.VimMotions = newValue == "enabled"
		UpdateKeyMap()
	case ThemeSelector:
		Settings.Theme = newValue
		ApplyTheme(Settings.Theme)
	}

	// save settings to file
	if err := saveSettings(Settings); err != nil {
		// TODO: add error ui
	}

	// update the list item
	items := m.list.Items()
	for i, item := range items {
		if settingItem, ok := item.(SettingItem); ok && settingItem.settingType == selectedItem.settingType {
			settingItem.value = newValue
			items[i] = settingItem
			break
		}
	}
	m.list.SetItems(items)

	return m, nil
}

func (m SettingsModel) View() string {
	header := headerStyle.Render("settings")
	content := header + "\n\n" + m.list.View()

	return centerHorizontally(content, m.list.Width())
}

// load and apply all settings at startup
func LoadAndApplySettings() {
	Settings = loadSettings()
	UpdateKeyMap()
	ApplyTheme(Settings.Theme)
}

// update theme color given basic color
func ApplyTheme(theme string) {
	switch theme {
	case "blue":
		SetThemeColors(blue, blueBg)
	case "green":
		SetThemeColors(green, greenBg)
	case "purple":
		SetThemeColors(purple, purpleBg)
	case "orange":
		SetThemeColors(orange, orangeBg)
	case "teal":
		SetThemeColors(teal, tealBg)
	case "pink":
		SetThemeColors(pink, pinkBg)
	default: // red
		SetThemeColors(red, redBg)
	}
}
