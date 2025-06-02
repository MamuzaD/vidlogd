package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type LogVideoModel struct {
	form FormModel
}

func NewLogVideoModel() LogVideoModel {
	fields := []FormField{
		{Placeholder: "https://youtube.com/watch?v=...", Required: true, CharLimit: 200, Width: 60},
		{Placeholder: "video title", Required: true, CharLimit: 100, Width: 50},
		{Placeholder: "channel name", Required: true, CharLimit: 50, Width: 50},
		{Placeholder: "YYYY-MM-DD", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "YYYY-MM-DD", Required: true, CharLimit: 10, Width: 12},
		{Placeholder: "write your review...", Required: false, CharLimit: 500, Width: 60},
	}

	labels := []string{
		"YouTube URL:",
		"Title:",
		"Channel:",
		"Video Release Date:",
		"Log Date:",
		"Review:",
	}

	form := NewForm(
		"log a video",
		fields,
		labels,
		"save video",
	)

	form.SetHandlers(
		func(f FormModel) tea.Cmd {
			video := createVideoFromForm(f)

			if err := saveVideo(video); err != nil {
				// TODO: add errors ui
			}

			return func() tea.Msg { return NavigateMsg{View: MainMenuView} }
		},
		func() tea.Cmd {
			return func() tea.Msg { return NavigateMsg{View: MainMenuView} }
		},
	)

	return LogVideoModel{form: form}
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
