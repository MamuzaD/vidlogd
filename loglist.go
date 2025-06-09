package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type LogListKeyMap struct{}

func (k LogListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		GlobalKeyMap.Up,
		GlobalKeyMap.Down,
		GlobalKeyMap.Select,
		GlobalKeyMap.Edit,
		GlobalKeyMap.Delete,
		GlobalKeyMap.Back,
		GlobalKeyMap.Help,
	}
}

func (k LogListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			GlobalKeyMap.Up,
			GlobalKeyMap.Down,
			GlobalKeyMap.Select,
		},
		{
			GlobalKeyMap.Edit,
			GlobalKeyMap.Delete,
			GlobalKeyMap.Back,
		},
		{
			GlobalKeyMap.Exit,
			GlobalKeyMap.Help,
		},
	}
}

type LogListModel struct {
	table  table.Model
	videos []Video
	help   help.Model
}

func NewLogListModel() LogListModel {
	columns := []table.Column{
		{Title: "Title", Width: 30},
		{Title: "Channel", Width: 20},
		{Title: "Rating", Width: 15},
		{Title: "Date Logged", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = tableHeaderStyle
	s.Selected = tableSelectedRowStyle

	t.SetStyles(s)

	h := help.New()
	h.ShowAll = false // start with compact help

	return LogListModel{
		table: t,
		help:  h,
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case LoadVideosMsg:
		m.videos = msg.videos
		m.updateTableRows()
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, GlobalKeyMap.Edit): // quick edit shortcut
			if len(m.videos) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.videos) {
					videoToEdit := m.videos[selectedRow]
					return m, func() tea.Msg {
						return NavigateMsg{
							View:    LogVideoView,
							VideoID: videoToEdit.ID,
						}
					}
				}
			}
			return m, nil
		case key.Matches(msg, GlobalKeyMap.Delete): // quick delete shortcut
			if len(m.videos) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.videos) {
					videoToDelete := m.videos[selectedRow]
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
			}
			return m, nil
		case key.Matches(msg, GlobalKeyMap.Select):
			return m.handleSelection()
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *LogListModel) updateTableRows() {
	rows := make([]table.Row, len(m.videos))
	for i, video := range m.videos {
		title := video.Title
		if title == "" {
			title = "Untitled"
		}

		channel := video.Channel
		if channel == "" {
			channel = "Unknown Channel"
		}

		// format rating as stars
		ratingStr := ""
		if video.Rating > 0 {
			for j := 1; j <= 5; j++ {
				starValue := float64(j)
				if video.Rating >= starValue {
					ratingStr += "★"
				} else if video.Rating >= starValue-0.5 {
					ratingStr += "⯨" // half star
				} else {
					ratingStr += "☆"
				}
			}
		} else {
			ratingStr = "Not rated"
		}

		logDate := video.LogDate
		if logDate == "" {
			logDate = "No date"
		}

		rows[i] = table.Row{title, channel, ratingStr, logDate}
	}
	m.table.SetRows(rows)
}

func (m LogListModel) handleSelection() (LogListModel, tea.Cmd) {
	if len(m.videos) == 0 {
		return m, nil
	}

	selectedRow := m.table.Cursor()
	if selectedRow < len(m.videos) {
		selectedVideo := m.videos[selectedRow]
		return m, func() tea.Msg {
			return NavigateMsg{View: LogDetailsView, VideoID: selectedVideo.ID}
		}
	}
	return m, nil
}

func (m LogListModel) View() string {
	var s strings.Builder

	s.WriteString("video logs\n\n")

	if len(m.videos) == 0 {
		s.WriteString("no videos logged yet\n\n")
		// Add help even when no videos
		keymap := LogListKeyMap{}
		s.WriteString(m.help.View(keymap))
		return s.String()
	}

	tableContent := m.table.View()
	styledTable := tableStyle.Render(tableContent)
	s.WriteString(styledTable)

	// Add help at the bottom
	keymap := LogListKeyMap{}
	s.WriteString("\n\n" + m.help.View(keymap))

	return s.String()
}
