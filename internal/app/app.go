package app

import (
	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/ui"
	"github.com/mamuzad/vidlogd/internal/ui/views"
)

type Model struct {
	currentView  models.ViewType
	currentRoute models.Route
	history      []models.Route

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

func routeEqual(a, b models.Route) bool {
	if a.View != b.View {
		return false
	}
	// state can hold non-comparable values
	return reflect.DeepEqual(a.State, b.State)
}

func (m Model) applyRoute(r models.Route) (Model, tea.Cmd) {
	m.currentRoute = r
	m.currentView = r.View

	switch r.View {
	case models.LogListView:
		m.logList = views.NewLogListModel()
		return m, m.logList.Init()
	case models.LogVideoView:
		// no state meaning preserve existing "new video" form state
		if r.State != nil {
			if st, ok := r.State.(models.VideoRouteState); ok && st.VideoID != "" {
				m.logVideo = views.NewLogVideoModel(st.VideoID)
			}
		}
		return m, m.logVideo.Init()
	case models.LogDetailsView:
		videoID := ""
		if st, ok := r.State.(models.VideoRouteState); ok {
			videoID = st.VideoID
		}
		m.logDetails = views.NewLogDetailsModel(videoID)
		return m, m.logDetails.Init()
	case models.SettingsView:
		m.settings = views.NewSettingsModel()
		return m, m.settings.Init()
	case models.StatsView:
		m.stats = views.NewStatsModel()
		return m, m.stats.Init()
	default:
		return m, nil
	}
}

func (m Model) navigateTo(r models.Route) (Model, tea.Cmd) {
	// only push if diff
	if routeEqual(m.currentRoute, r) {
		return m, nil
	}
	m.history = append(m.history, m.currentRoute)
	return m.applyRoute(r)
}

func (m Model) back() (Model, tea.Cmd) {
	if len(m.history) == 0 {
		return m.applyRoute(models.Route{View: models.MainMenuView})
	}

	last := len(m.history) - 1
	prev := m.history[last]
	m.history = m.history[:last]
	return m.applyRoute(prev)
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
		}

	case models.ClearFormMsg:
		// clear the form by creating a new empty one
		m.logVideo = views.NewLogVideoModel("")
		return m, nil

	case models.BackMsg:
		return m.back()

	case models.NavigateMsg:
		return m.navigateTo(models.Route(msg))
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
		currentRoute: models.Route{
			View: models.MainMenuView,
		},
		history:  []models.Route{},
		mainMenu: views.NewMainMenuModel(),
		logVideo: views.NewLogVideoModel(""),
		settings: views.NewSettingsModel(),
		stats:    views.NewStatsModel(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		return err
	}

	return nil
}
