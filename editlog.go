package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type EditLogModel struct {
	form    FormModel
	videoID string
}

func NewEditLogModel(videoID string) EditLogModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Label: "YouTube URL:", Required: true, CharLimit: 200, Width: 60},
		{Placeholder: "video title", Label: "Title:", Required: true, CharLimit: 100, Width: 50},
		{Placeholder: "channel name", Label: "Channel:", Required: true, CharLimit: 50, Width: 50},
		{Placeholder: "YYYY-MM-DD", Label: "Video Release Date:", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "YYYY-MM-DD", Label: "Log Date:", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "write your review...", Label: "Review:", Required: false, CharLimit: 500, Width: 60},
	}

	// If editing an existing video, populate the fields with existing data
	if videoID != "" {
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

	if videoID != "" {
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
			if videoID != "" {
				// Load existing video and update it
				existingVideo, err := findVideoByID(videoID)
				if err != nil {
					// TODO: add errors ui
					return func() tea.Msg { return NavigateMsg{View: LogListView} }
				}
				video = *existingVideo
				video.URL = f.GetValue(url)
				video.Title = f.GetValue(title)
				video.Channel = f.GetValue(channel)
				video.ReleaseDate = f.GetValue(release)
				video.LogDate = f.GetValue(logDate)
				video.Review = f.GetValue(review)

				if err := updateVideo(video); err != nil {
					// TODO: add errors ui
				}
			} else {
				// Create new video
				video = createVideoFromForm(f)
				if err := saveVideo(video); err != nil {
					// TODO: add errors ui
				}
			}

			return func() tea.Msg { return NavigateMsg{View: LogListView} }
		},
		func() tea.Cmd {
			return func() tea.Msg { return NavigateMsg{View: LogListView} }
		},
	)

	return EditLogModel{form: form, videoID: videoID}
}

func (m EditLogModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m EditLogModel) Update(msg tea.Msg) (EditLogModel, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m EditLogModel) View() string {
	return m.form.View()
}
