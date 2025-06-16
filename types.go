package main

import "time"

type ViewType int

type NavigateMsg struct {
	View    ViewType
	VideoID string // video to edit
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
}

var (
	ISODateFormat  string = "2006-01-02"
	MonthFormat    string = "01/06"
	DateTimeFormat string = "2006-01-02 3:04 PM"
)
