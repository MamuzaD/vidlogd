package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type LogListModel struct {
	cursor int
	videos []Video
}

func NewLogListModel() LogListModel {
	return LogListModel{
		cursor: 0,
	}
}

func (m LogListModel) Init() tea.Cmd {
	return func() tea.Msg {
		videos, err := loadVideos()
		if err != nil {
			return err
		}
		return LoadVideosMsg{videos: videos}
	}
}

type LoadVideosMsg struct {
	videos []Video
}

func (m LogListModel) Update(msg tea.Msg) (LogListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadVideosMsg:
		m.videos = msg.videos
		// adjust cursor if it's out of bounds after deletion
		if m.cursor >= len(m.videos) && len(m.videos) > 0 {
			m.cursor = len(m.videos) - 1
		} else if len(m.videos) == 0 {
			m.cursor = 0
		}

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		// delete video
		case "x":
			if len(m.videos) > 0 {
				videoToDelete := m.videos[m.cursor]
				return m, func() tea.Msg {
					err := deleteVideo(videoToDelete.ID)
					if err != nil {
						return err
					}
					// reload videos after deletion
					videos, err := loadVideos()
					if err != nil {
						return err
					}
					return LoadVideosMsg{videos: videos}
				}
			}
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.videos)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.handleSelection()
		}
	}
	return m, nil
}

func (m LogListModel) handleSelection() (LogListModel, tea.Cmd) {
	if len(m.videos) == 0 {
		return m, nil
	}

	selectedVideo := m.videos[m.cursor]
	return m, func() tea.Msg {
		return NavigateMsg{View: LogVideoView, VideoID: selectedVideo.ID}
	}
}

func (m LogListModel) View() string {
	var s strings.Builder

	s.WriteString("video logs\n\n")

	if len(m.videos) == 0 {
		s.WriteString("No videos logged yet.\n\n")
		s.WriteString("Press 'q' to go back to main menu\n")
		return s.String()
	}

	for i, video := range m.videos {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}

		// Format the video entry
		title := video.Title
		if title == "" {
			title = "Untitled"
		}

		channel := video.Channel
		if channel == "" {
			channel = "Unknown Channel"
		}

		logDate := video.LogDate
		if logDate == "" {
			logDate = "No date"
		}

		s.WriteString(fmt.Sprintf("%s%s\n", cursor, title))

		// show rating stars if available
		ratingStr := ""
		if video.Rating > 0 {
			for i := 1; i <= 5; i++ {
				starValue := float64(i)
				if video.Rating >= starValue {
					ratingStr += "★"
				} else if video.Rating >= starValue-0.5 {
					ratingStr += "⯨" // half star
				} else {
					ratingStr += "☆"
				}
			}
			ratingStr = fmt.Sprintf("   Rating: %s\n", ratingStr)
		}

		s.WriteString(fmt.Sprintf("   Channel: %s\n   Logged: %s\n%s\n", channel, logDate, ratingStr))

		if video.Review != "" {
			review := video.Review
			if len(review) > 60 {
				review = review[:57] + "..."
			}
			s.WriteString(fmt.Sprintf("   Review: %s\n", review))
		}

	}

	return s.String()
}
