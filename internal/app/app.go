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

	mainMenu   *views.MainMenuModel
	logVideo   *views.LogVideoModel
	logList    *views.LogListModel
	logDetails *views.LogDetailsModel
	settings   *views.SettingsModel
	stats      *views.StatsModel

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
	case models.MainMenuView:
		if m.mainMenu == nil {
			mm := views.NewMainMenuModel()
			m.mainMenu = &mm
		}
		return m, m.mainMenu.Init()
	case models.LogListView:
		if m.logList == nil {
			ll := views.NewLogListModel()
			m.logList = &ll
		}
		return m, m.logList.Init()
	case models.LogVideoView:
		if m.logVideo == nil {
			lv := views.NewLogVideoModel("")
			m.logVideo = &lv
		}

		targetID := ""
		if st, ok := r.State.(models.VideoRouteState); ok {
			targetID = st.VideoID
		}

		// route param changes.
		if targetID != "" && (m.logVideo == nil || m.logVideo.VideoID() != targetID) {
			lv := views.NewLogVideoModel(targetID)
			m.logVideo = &lv
		}
		// switching from editing -> new: reset
		if targetID == "" && m.logVideo != nil && m.logVideo.VideoID() != "" {
			lv := views.NewLogVideoModel("")
			m.logVideo = &lv
		}

		return m, m.logVideo.Init()
	case models.LogDetailsView:
		videoID := ""
		if st, ok := r.State.(models.VideoRouteState); ok {
			videoID = st.VideoID
		}
		if m.logDetails == nil || m.logDetails.VideoID() != videoID {
			ld := views.NewLogDetailsModel(videoID)
			m.logDetails = &ld
		}
		return m, m.logDetails.Init()
	case models.SettingsView:
		index := 0
		if st, ok := r.State.(models.SettingsRouteState); ok {
			index = st.ListIndex
		}
		if m.settings == nil {
			s := views.NewSettingsModel(index)
			m.settings = &s
		} else {
			m.settings.SelectIndex(index)
		}
		return m, m.settings.Init()
	case models.StatsView:
		if m.stats == nil {
			s := views.NewStatsModel()
			m.stats = &s
		}
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
		if m.logVideo == nil {
			lv := views.NewLogVideoModel("")
			m.logVideo = &lv
		} else {
			*m.logVideo = views.NewLogVideoModel("")
		}
		return m, nil

	case models.BackMsg:
		return m.back()

	case models.NavigateMsg:
		return m.navigateTo(models.Route(msg))
	}

	switch m.currentView {
	case models.MainMenuView:
		cmd = updatePtr(&m.mainMenu, msg, views.NewMainMenuModel)
	case models.LogVideoView:
		cmd = updatePtr(&m.logVideo, msg, func() views.LogVideoModel { return views.NewLogVideoModel("") })
	case models.LogListView:
		cmd = updatePtr(&m.logList, msg, views.NewLogListModel)
	case models.LogDetailsView:
		cmd = updatePtr(&m.logDetails, msg, func() views.LogDetailsModel { return views.NewLogDetailsModel("") })
	case models.SettingsView:
		cmd = updatePtr(&m.settings, msg, func() views.SettingsModel { return views.NewSettingsModel(0) })
	case models.StatsView:
		cmd = updatePtr(&m.stats, msg, views.NewStatsModel)
	}

	return m, cmd
}

func (m Model) View() string {
	var content string

	switch m.currentView {
	case models.MainMenuView:
		if m.mainMenu != nil {
			content = m.mainMenu.View()
		}
	case models.LogVideoView:
		if m.logVideo != nil {
			content = m.logVideo.View()
		}
	case models.LogListView:
		if m.logList != nil {
			content = m.logList.View()
		}
	case models.LogDetailsView:
		if m.logDetails != nil {
			content = m.logDetails.View()
		}
	case models.SettingsView:
		if m.settings != nil {
			content = m.settings.View()
		}

	case models.StatsView:
		if m.stats != nil {
			content = m.stats.View()
		}
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
		mainMenu: func() *views.MainMenuModel { mm := views.NewMainMenuModel(); return &mm }(),
		logVideo: func() *views.LogVideoModel { lv := views.NewLogVideoModel(""); return &lv }(),
		settings: func() *views.SettingsModel { s := views.NewSettingsModel(0); return &s }(),
		stats:    func() *views.StatsModel { s := views.NewStatsModel(); return &s }(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		return err
	}

	return nil
}
