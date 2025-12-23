package ui

type ViewType int

const (
	MainMenuView ViewType = iota
	LogVideoView
	LogListView
	LogDetailsView
	SettingsView
	StatsView
	SyncView
)

type Route struct {
	View  ViewType
	State any
}

type (
	VideoRouteState struct {
		VideoID string
	}
	SettingsRouteState struct {
		ListIndex int
	}
)

type (
	NavigateMsg struct {
		View  ViewType
		State any // arbitrary view state (route params)
	}
	ClearFormMsg struct{}
	BackMsg      struct{}
	UIRefreshMsg struct{}
)
