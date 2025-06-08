package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type LogVideoModel struct {
	form    FormModel
	videoID string
}

func NewLogVideoModel(videoID string) LogVideoModel {
	editing := videoID != ""
	var existingVideo *Video

	// load existing video if editing
	if editing {
		if video, err := findVideoByID(videoID); err == nil {
			existingVideo = video
		}
	}

	form := NewVideoLogForm(editing, existingVideo)

	form.SetHandlers(
		func(f FormModel) tea.Cmd {
			if editing {
				// update existing video
				if existingVideo != nil {
					video := createVideoFromForm(f)
					video.ID = existingVideo.ID // preserve the original ID

					if err := updateVideo(video); err != nil {
						// TODO: add errors ui
					}
				}
				return func() tea.Msg { return NavigateMsg{View: LogListView} }
			} else {
				// create new video
				video := createVideoFromForm(f)
				if err := saveVideo(video); err != nil {
					// TODO: add errors ui
				}
				// clear form by sending clear message then navigate
				return tea.Batch(
					func() tea.Msg { return ClearFormMsg{} },
					func() tea.Msg { return NavigateMsg{View: MainMenuView} },
				)
			}
		},
		func() tea.Cmd {
			if editing {
				// when editing, cancel without saving (reset)
				return func() tea.Msg { return NavigateMsg{View: LogListView} }
			} else {
				// when creating new, preserve form state and go back to main menu
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

	switch msg := msg.(type) {
	case MetadataFetchedMsg:
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	default:
		m.form, cmd = m.form.Update(msg)
		return m, cmd
	}
}

func (m LogVideoModel) View() string {
	return m.form.View()
}
