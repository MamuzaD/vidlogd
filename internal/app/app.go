package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/ui"
	"github.com/mamuzad/vidlogd/internal/ui/views"
)

type Model struct {
	currentView models.ViewType

	mainMenu   views.MainMenuModel
	logVideo   views.LogVideoModel
	logList    views.LogListModel
	logDetails views.LogDetailsModel
	settings   views.SettingsModel
	stats      views.StatsModel

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
			if m.currentView != models.MainMenuView {
				m.currentView = models.MainMenuView
				return m, nil
			}
		}

	case models.ClearFormMsg:
		// clear the form by creating a new empty one
		m.logVideo = views.NewLogVideoModel("")
		return m, nil

	case models.NavigateMsg:
		m.currentView = msg.View
		if msg.View == models.LogListView {
			m.logList = views.NewLogListModel()
			return m, m.logList.Init()
		}
		if msg.View == models.LogVideoView {
			if msg.VideoID == "" {
				// preserve existing new video form state
			} else {
				m.logVideo = views.NewLogVideoModel(msg.VideoID)
			}
			return m, m.logVideo.Init()
		}
		if msg.View == models.LogDetailsView {
			m.logDetails = views.NewLogDetailsModel(msg.VideoID)
			return m, m.logDetails.Init()
		}
		if msg.View == models.SettingsView {
			m.settings = views.NewSettingsModel()
			return m, m.settings.Init()
		}
		if msg.View == models.StatsView {
			m.stats = views.NewStatsModel()
			return m, m.stats.Init()
		}

		return m, nil
	}

	switch m.currentView {
	case models.MainMenuView:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case models.LogVideoView:
		m.logVideo, cmd = m.logVideo.Update(msg)
	case models.LogListView:
		m.logList, cmd = m.logList.Update(msg)
	case models.LogDetailsView:
		m.logDetails, cmd = m.logDetails.Update(msg)
	case models.SettingsView:
		m.settings, cmd = m.settings.Update(msg)
	case models.StatsView:
		m.stats, cmd = m.stats.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var content string

	switch m.currentView {
	case models.MainMenuView:
		content = m.mainMenu.View()
	case models.LogVideoView:
		content = m.logVideo.View()
	case models.LogListView:
		content = m.logList.View()
	case models.LogDetailsView:
		content = m.logDetails.View()
	case models.SettingsView:
		content = m.settings.View()
	case models.StatsView:
		content = m.stats.View()
	}

	title := ui.CenterHorizontally(ui.TitleStyle.Render("vidlogd"), lipgloss.Width(content))
	// wrap content in popup
	styledContent := ui.PopupStyle.Render(title + "\n" + content)
	// center the popup
	if m.width > 0 && m.height > 0 {
		return ui.CenterBoth("\n\n"+styledContent, m.width, m.height)
	}

	return styledContent
}

func Run() error {
	// load settings first
	views.LoadAndApplySettings()

	m := Model{
		currentView: models.MainMenuView,
		mainMenu:    views.NewMainMenuModel(),
		logVideo:    views.NewLogVideoModel(""),
		settings:    views.NewSettingsModel(),
		stats:       views.NewStatsModel(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		return err
	}

	return nil
}
