package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// necessary for list
type ActionItem struct {
	title string
}

func (i ActionItem) FilterValue() string { return i.title }

type ActionItemDelegate struct{}

func (d ActionItemDelegate) Height() int                               { return 1 }
func (d ActionItemDelegate) Spacing() int                              { return 0 }
func (d ActionItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d ActionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ActionItem)
	if !ok {
		return
	}

	style := menuItemStyle
	if index == m.Index() {
		style = style.Background(primaryColor).Foreground(white)
	}

	styledText := style.Render(i.title)
	fmt.Fprint(w, styledText)
}

type LogDetailsKeyMap struct{}

func (k LogDetailsKeyMap) ShortHelp() []key.Binding {
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

func (k LogDetailsKeyMap) FullHelp() [][]key.Binding {
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

type LogDetailsModel struct {
	video       *Video
	actionsList list.Model
	help        help.Model
}

func NewLogDetailsModel(videoID string) LogDetailsModel {
	var video *Video
	if foundVideo, err := findVideoByID(videoID); err == nil {
		video = foundVideo
	}

	items := []list.Item{
		ActionItem{title: "edit"},
		ActionItem{title: "delete"},
		ActionItem{title: "back"},
	}

	const defaultWidth = 40
	const listHeight = 5

	l := list.New(items, ActionItemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	h := help.New()
	h.ShowAll = false // start with compact help

	return LogDetailsModel{
		video:       video,
		actionsList: l,
		help:        h,
	}
}

func (m LogDetailsModel) Init() tea.Cmd {
	return nil
}

func (m LogDetailsModel) Update(msg tea.Msg) (LogDetailsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, GlobalKeyMap.Edit): // quick edit shortcut
			if m.video != nil {
				return m, func() tea.Msg {
					return NavigateMsg{
						View:    LogVideoView,
						VideoID: m.video.ID,
					}
				}
			}
		case key.Matches(msg, GlobalKeyMap.Delete): // quick delete shortcut
			if m.video != nil {
				if err := deleteVideo(m.video.ID); err == nil {
					return m, func() tea.Msg {
						return NavigateMsg{View: LogListView}
					}
				}
			}
		case key.Matches(msg, GlobalKeyMap.Select):
			selectedItem, ok := m.actionsList.SelectedItem().(ActionItem)
			if !ok {
				return m, nil
			}

			switch selectedItem.title {
			case "edit":
				if m.video != nil {
					return m, func() tea.Msg {
						return NavigateMsg{
							View:    LogVideoView,
							VideoID: m.video.ID,
						}
					}
				}
			case "delete":
				if m.video != nil {
					if err := deleteVideo(m.video.ID); err == nil {
						return m, func() tea.Msg {
							return NavigateMsg{View: LogListView}
						}
					}
				}
			case "back":
				return m, func() tea.Msg {
					return NavigateMsg{View: LogListView}
				}
			}
		case key.Matches(msg, GlobalKeyMap.Back):
			return m, func() tea.Msg {
				return NavigateMsg{View: LogListView}
			}
		}
	}

	var cmd tea.Cmd
	m.actionsList, cmd = m.actionsList.Update(msg)
	return m, cmd
}

// helper to render stars
func renderStars(rating float64) string {
	var stars strings.Builder
	for i := 1; i <= 5; i++ {
		starValue := float64(i)
		if rating >= starValue {
			stars.WriteString("★") // filled star
		} else if rating >= starValue-0.5 {
			stars.WriteString("⯨") // half star
		} else {
			stars.WriteString("☆")
		}
	}
	return stars.String()
}

func (m LogDetailsModel) View() string {
	if m.video == nil {
		return "Log not found"
	}

	var s strings.Builder

	s.WriteString(headerStyle.Render("log details"))

	// video info
	s.WriteString("Title: " + m.video.Title + "\n\n")
	s.WriteString("Channel: " + m.video.Channel + "\n\n")
	s.WriteString("URL: " + m.video.URL + "\n\n")
	s.WriteString("Release Date: " + m.video.ReleaseDate + "\n\n")

	s.WriteString("Date Logged: " + m.video.LogDate + "\n\n")
	var rewatched string
	if m.video.Rewatched {
		rewatched = "  rewatched"
	} else {
		rewatched = "  first watch"
	}
	s.WriteString(fmt.Sprintf(
		"rating: %s (%.1f/5)  %s\n\n",
		renderStars(m.video.Rating),
		m.video.Rating,
		rewatched,
	))

	if m.video.Review != "" {
		s.WriteString("Review:" + "\n")
		s.WriteString(reviewStyle.Render(m.video.Review) + "\n\n")
	} else {
		s.WriteString("Review:" + "\n")
		s.WriteString(reviewStyle.Render("no review") + "\n\n")
	}

	s.WriteString("actions" + "\n\n")
	s.WriteString(m.actionsList.View() + "\n")

	keymap := LogDetailsKeyMap{}
	s.WriteString(m.help.View(keymap))

	return s.String()
}
