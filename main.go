package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MainMenuView ViewType = iota
	LogVideoView
	LogListView
	LogDetailsView
	SettingsView
	StatsView
)

type Model struct {
	currentView ViewType

	mainMenu   MainMenuModel
	logVideo   LogVideoModel
	logList    LogListModel
	logDetails LogDetailsModel
	settings   SettingsModel
	stats      StatsModel

	// Terminal dimensions for centering
	width  int
	height int
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("vidlogd")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.currentView != MainMenuView {
				m.currentView = MainMenuView
				return m, nil
			}
		}

	case ClearFormMsg:
		// clear the form by creating a new empty one
		m.logVideo = NewLogVideoModel("")
		return m, nil

	case NavigateMsg:
		m.currentView = msg.View
		if msg.View == LogListView {
			m.logList = NewLogListModel()
			return m, m.logList.Init()
		}
		if msg.View == LogVideoView {
			if msg.VideoID == "" && m.logVideo.videoID == "" {
				// preserve existing new video form state
			} else {
				m.logVideo = NewLogVideoModel(msg.VideoID)
			}
			return m, m.logVideo.Init()
		}
		if msg.View == LogDetailsView {
			m.logDetails = NewLogDetailsModel(msg.VideoID)
			return m, m.logDetails.Init()
		}
		if msg.View == SettingsView {
			m.settings = NewSettingsModel()
			return m, m.settings.Init()
		}
		if msg.View == StatsView {
			m.stats = NewStatsModel()
			return m, m.stats.Init()
		}

		return m, nil
	}

	switch m.currentView {
	case MainMenuView:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case LogVideoView:
		m.logVideo, cmd = m.logVideo.Update(msg)
	case LogListView:
		m.logList, cmd = m.logList.Update(msg)
	case LogDetailsView:
		m.logDetails, cmd = m.logDetails.Update(msg)
	case SettingsView:
		m.settings, cmd = m.settings.Update(msg)
	case StatsView:
		m.stats, cmd = m.stats.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var content string

	switch m.currentView {
	case MainMenuView:
		content = m.mainMenu.View()
	case LogVideoView:
		content = m.logVideo.View()
	case LogListView:
		content = m.logList.View()
	case LogDetailsView:
		content = m.logDetails.View()
	case SettingsView:
		content = m.settings.View()
	case StatsView:
		content = m.stats.View()
	}

	title := centerHorizontally(titleStyle.Render("vidlogd"), lipgloss.Width(content))
	// wrap content in popup
	styledContent := popupStyle.Render(title + "\n" + content)
	// center the popup
	if m.width > 0 && m.height > 0 {
		return centerBoth("\n\n"+styledContent, m.width, m.height)
	}

	return styledContent
}

func main() {
	// load settings first
	LoadAndApplySettings()

	m := Model{
		currentView: MainMenuView,
		mainMenu:    NewMainMenuModel(),
		logVideo:    NewLogVideoModel(""),
		settings:    NewSettingsModel(),
		stats:       NewStatsModel(),
	}

	p := tea.
		NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
