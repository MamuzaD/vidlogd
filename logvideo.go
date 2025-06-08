package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type LogVideoModel struct {
	form    FormModel
	videoID string
}

func NewLogVideoModel(videoID string) LogVideoModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Label: "YouTube URL:", Required: true, CharLimit: 200, Width: 60},
		{Placeholder: "video title", Label: "Title:", Required: true, CharLimit: 100, Width: 50},
		{Placeholder: "channel name", Label: "Channel:", Required: true, CharLimit: 50, Width: 50},
		{Placeholder: "YYYY-MM-DD", Label: "Video Release Date:", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "YYYY-MM-DD", Label: "Log Date:", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "write your review...", Label: "Review:", Required: false, CharLimit: 500, Width: 60},
	}

	editing := videoID != ""

	// editing exisiting video
	if editing {
		if existingVideo, err := findVideoByID(videoID); err == nil {
			fields[url].Value = existingVideo.URL
			fields[title].Value = existingVideo.Title
			fields[channel].Value = existingVideo.Channel
			fields[release].Value = existingVideo.ReleaseDate
			fields[logDate].Value = existingVideo.LogDate
			fields[review].Value = existingVideo.Review
		}
	}

	var formTitle string
	var buttonText string

	if editing {
		formTitle = "edit video log"
		buttonText = "update video"
	} else {
		formTitle = "log a video"
		buttonText = "save video"
	}

	form := NewForm(
		formTitle,
		fields,
		buttonText,
	)

	form.SetHandlers(
		func(f FormModel) tea.Cmd {
			var video Video
			if editing {
				existingVideo, err := findVideoByID(videoID)
				if err != nil {
					// TODO: add errors ui
					return func() tea.Msg { return NavigateMsg{View: LogListView} }
				}
				video = createVideoFromForm(f)
				video.ID = existingVideo.ID // preserve the original ID

				if err := updateVideo(video); err != nil {
					// TODO: add errors ui
				}
				return func() tea.Msg { return NavigateMsg{View: LogListView} }
			} else {
				// create new video
				video = createVideoFromForm(f)
				if err := saveVideo(video); err != nil {
					// TODO: add errors ui
				}
				return func() tea.Msg { return NavigateMsg{View: MainMenuView} }
			}
		},
		func() tea.Cmd {
			if videoID != "" {
				return func() tea.Msg { return NavigateMsg{View: LogListView} }
			} else {
				return func() tea.Msg { return NavigateMsg{View: MainMenuView} }
			}
		},
	)

	return LogVideoModel{form: form, videoID: videoID}
}

func (m LogVideoModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m LogVideoModel) Update(msg tea.Msg) (LogVideoModel, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m LogVideoModel) View() string {
	return m.form.View()
}
