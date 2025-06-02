package main

import "time"

// Form field indices
const (
	url = iota
	title
	channel
	release
	logDate
	review
)

type ViewType int

const (
	MainMenuView ViewType = iota
	LogVideoView
	EditLogView
)

type NavigateMsg struct {
	View    ViewType
	VideoID string // video to edit
}

type Video struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Channel     string    `json:"channel"`
	ReleaseDate string    `json:"release_date"`
	LogDate     string    `json:"log_date"`
	Review      string    `json:"review"`
	CreatedAt   time.Time `json:"created_at"`
}
