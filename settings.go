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
	APIKeyEditor
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
	form *FormModel // for API key editing
}

func NewSettingsModel() SettingsModel {
	Settings = loadSettings()

	UpdateKeyMap()

	displayAPIKey := renderAPIKey()

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
		SettingItem{
			settingType: APIKeyEditor,
			title:       "YouTube API Key",
			description: "set your YouTube Data API v3 key",
			value:       displayAPIKey,
			options:     []string{"edit"},
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
	case ClearSettingsFormMsg:
		m.form = nil

		items := m.list.Items()
		for i, item := range items {
			if settingItem, ok := item.(SettingItem); ok && settingItem.settingType == APIKeyEditor {
				displayAPIKey := renderAPIKey()
				settingItem.value = displayAPIKey
				items[i] = settingItem
				break
			}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, GlobalKeyMap.Back) {
			return m, func() tea.Msg {
				return NavigateMsg{View: MainMenuView}
			}
		}
		// let form handle select if active
		if m.form == nil && key.Matches(msg, GlobalKeyMap.Select, GlobalKeyMap.Right) {
			return m.handleSettingSelection()
		}
		if key.Matches(msg, GlobalKeyMap.Left) {
			return m.cycleSetting()
		}
	}

	// handle form updates
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		m.form = &form
		return m, cmd
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

func (m SettingsModel) handleSettingSelection() (SettingsModel, tea.Cmd) {
	selectedItem, ok := m.list.SelectedItem().(SettingItem)
	if !ok {
		return m, nil
	}

	switch selectedItem.settingType {
	case APIKeyEditor:
		fields := []FormField{
			{Placeholder: "your_youtube_api_key_here", Label: "YouTube API Key:", Required: false, CharLimit: 100, Width: 60, Type: FormFieldText, Value: Settings.APIKey},
		}

		form := NewForm("YouTube API Key", fields, "save")
		form.SetHandlers(
			func(f FormModel) tea.Cmd {
				apiKeyValue := f.GetValue(0)
				Settings.APIKey = apiKeyValue
				if err := saveSettings(Settings); err != nil {
					// TODO: handle error
				}

				return func() tea.Msg {
					return ClearSettingsFormMsg{}
				}
			},
			func() tea.Cmd {
				return func() tea.Msg {
					return ClearSettingsFormMsg{}
				}
			},
		)
		m.form = &form
		return m, nil
	default:
		return m.cycleSetting()
	}
}

func (m SettingsModel) View() string {
	if m.form != nil {
		return m.form.View()
	}

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

type ClearSettingsFormMsg struct{}

func renderAPIKey() (apiKey string) {
	apiKey = Settings.APIKey
	if apiKey != "" {
		if len(apiKey) > 8 {
			apiKey = apiKey[:8] + "***"
		} else {
			apiKey = "***"
		}
	} else {
		apiKey = "not set"
	}

	return
}
