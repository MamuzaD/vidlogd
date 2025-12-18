package models

import "time"

type ViewType int

const (
	MainMenuView ViewType = iota
	LogVideoView
	LogListView
	LogDetailsView
	SettingsView
	StatsView
)

type NavigateMsg struct {
	View  ViewType
	State any // arbitrary view state (route params)
}

type BackMsg struct{}

type Route struct {
	View  ViewType
	State any
}

// -- route state payloads --

type VideoRouteState struct {
	VideoID string
}

type SettingsRouteState struct {
	ListIndex int
}

type ClearFormMsg struct{}

type Video struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Channel     string    `json:"channel"`
	ReleaseDate string    `json:"release_date"`
	LogDate     string    `json:"log_date"`
	Rating      float64   `json:"rating"`
	Rewatched   bool      `json:"rewatched"`
	Review      string    `json:"review"`
	CreatedAt   time.Time `json:"created_at"`
}

type AppSettings struct {
	VimMotions bool   `json:"vim_motions"`
	Theme      string `json:"theme"`
	APIKey     string `json:"api_key"`
}

var (
	ISODateFormat  string = "2006-01-02"
	MonthFormat    string = "01/06"
	DateTimeFormat string = "2006-01-02 3:04 PM"
)

// @TODO :: need to reorganize
func GetDefaultSettings() AppSettings {
	return AppSettings{
		VimMotions: true,
		Theme:      "red",
		APIKey:     "",
	}
}
