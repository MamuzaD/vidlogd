package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/services"
)

type LogVideoModel struct {
	form    FormModel
	videoID string
}

func NewLogVideoModel(videoID string) LogVideoModel {
	editing := videoID != ""
	var existingVideo *models.Video

	// load existing video if editing
	if editing {
		if video, err := models.FindVideoByID(videoID); err == nil {
			existingVideo = video
		}
	}

	form := NewVideoLogForm(editing, existingVideo)

	form.SetHandlers(
		func(f FormModel) tea.Cmd {
			if editing {
				// update existing video
				if existingVideo != nil {
					video := CreateVideoFromForm(f)
					video.ID = existingVideo.ID // preserve the original ID

					if err := models.UpdateVideo(video); err != nil {
						// TODO: add errors ui
					}
				}
				return func() tea.Msg { return models.BackMsg{} }
			} else {
				// create new video
				video := CreateVideoFromForm(f)
				if err := models.SaveVideo(video); err != nil {
					// TODO: add errors ui
				}
				// clear form by sending clear message then navigate
				return tea.Batch(
					func() tea.Msg { return models.ClearFormMsg{} },
					func() tea.Msg { return models.BackMsg{} },
				)
			}
		},
		func() tea.Cmd {
			if editing {
				// when editing, cancel without saving (reset)
				return func() tea.Msg { return models.BackMsg{} }
			} else {
				// when creating new, preserve form state and go back to main menu
				return func() tea.Msg { return models.BackMsg{} }
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
	case services.MetadataFetchedMsg:
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

func (m LogVideoModel) VideoID() string { return m.videoID }

// UpdateVimMode updates the vim mode setting for the form
func (m *LogVideoModel) UpdateVimMode() {
	m.form.UpdateVimMode()
}
