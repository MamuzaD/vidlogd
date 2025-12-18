package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/ui"
	"github.com/sahilm/fuzzy"
)

type LogListKeyMap struct{}

func (k LogListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		ui.GlobalKeyMap.Up,
		ui.GlobalKeyMap.Down,
		ui.GlobalKeyMap.Select,
		ui.GlobalKeyMap.Edit,
		ui.GlobalKeyMap.Delete,
		ui.GlobalKeyMap.Back,
		ui.GlobalKeyMap.Search,
		ui.GlobalKeyMap.Help,
	}
}

func (k LogListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			ui.GlobalKeyMap.Up,
			ui.GlobalKeyMap.Down,
			ui.GlobalKeyMap.Select,
		},
		{
			ui.GlobalKeyMap.Edit,
			ui.GlobalKeyMap.Delete,
			ui.GlobalKeyMap.Back,
		},
		{
			ui.GlobalKeyMap.Exit,
			ui.GlobalKeyMap.Search,
			ui.GlobalKeyMap.Help,
		},
	}
}

type LogListModel struct {
	table      table.Model
	videos     []models.Video
	help       help.Model
	search     textinput.Model
	filtered   []models.Video
	isFiltered bool
	focused    bool
}

func NewLogListModel() LogListModel {
	columns := []table.Column{
		{Title: "Title", Width: 35},
		{Title: "Channel", Width: 15},
		{Title: "Rating", Width: 8},
		{Title: "Date Logged", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = ui.TableHeaderStyle
	s.Selected = ui.TableSelectedRowStyle

	t.SetStyles(s)

	h := help.New()
	h.ShowAll = false // start with compact help

	search := textinput.New()
	search.Placeholder = "search videos..."
	search.Prompt = "  "
	search.CharLimit = 50
	search.Width = 50

	return LogListModel{
		table:      t,
		help:       h,
		search:     search,
		filtered:   []models.Video{},
		isFiltered: false,
		focused:    false,
	}
}

func (m LogListModel) Init() tea.Cmd {
	return func() tea.Msg {
		videos, err := models.LoadVideos()
		if err != nil {
			return err
		}
		return LoadVideosMsg{videos: videos}
	}
}

type LoadVideosMsg struct {
	videos []models.Video
}

func (m *LogListModel) filterVideos() {
	if m.search.Value() == "" {
		m.isFiltered = false
		m.filtered = m.videos
		return
	}

	m.isFiltered = true
	searchable := make([]string, len(m.videos))
	for i, v := range m.videos {
		searchable[i] = strings.ToLower(v.Title + " " + v.Channel)
	}

	matches := fuzzy.Find(m.search.Value(), searchable)
	m.filtered = make([]models.Video, len(matches))
	for i, match := range matches {
		m.filtered[i] = m.videos[match.Index]
	}
}

func (m LogListModel) Update(msg tea.Msg) (LogListModel, tea.Cmd) {
	var cmd tea.Cmd
	var searchCmd tea.Cmd

	switch msg := msg.(type) {
	case LoadVideosMsg:
		m.videos = msg.videos
		m.filtered = msg.videos
		m.updateTableRows()
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, ui.GlobalKeyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, ui.GlobalKeyMap.Search), (m.focused && key.Matches(msg, ui.GlobalKeyMap.SearchBack)):
			if m.focused {
				m.focused = false
				m.search.Blur()
				m.table.Focus()
				m.filterVideos()
			} else {
				m.focused = true
				m.search.Focus()
				m.table.Blur()
			}
			return m, nil
		case m.focused:
			// when search is focused, only handle search input
			m.search, searchCmd = m.search.Update(msg)
			m.filterVideos()
			m.updateTableRows()
			return m, searchCmd
		case key.Matches(msg, ui.GlobalKeyMap.Edit): // quick edit shortcut
			if len(m.videos) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.videos) {
					videoToEdit := m.videos[selectedRow]
					return m, func() tea.Msg {
						return models.NavigateMsg{
							View:  models.LogVideoView,
							State: models.VideoRouteState{VideoID: videoToEdit.ID},
						}
					}
				}
			}
			return m, nil
		case key.Matches(msg, ui.GlobalKeyMap.Delete): // quick delete shortcut
			if len(m.videos) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.videos) {
					videoToDelete := m.videos[selectedRow]
					return m, func() tea.Msg {
						err := models.DeleteVideo(videoToDelete.ID)
						if err != nil {
							return err
						}
						// reload videos after deletion
						videos, err := models.LoadVideos()
						if err != nil {
							return err
						}
						return LoadVideosMsg{videos: videos}
					}
				}
			}
			return m, nil
		case key.Matches(msg, ui.GlobalKeyMap.Select):
			return m.handleSelection()
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *LogListModel) updateTableRows() {
	videosToUse := m.videos
	if m.isFiltered {
		videosToUse = m.filtered
	}

	rows := make([]table.Row, len(videosToUse))
	for i, video := range videosToUse {
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
					ratingStr += "★" // filled star
				} else if video.Rating >= starValue-0.5 {
					ratingStr += "⯨" // half star
				} else {
					ratingStr += "☆"
				}
			}
		}

		logDate := video.LogDate
		// add a leading space if the hour is a single digit.
		colon := 12
		hour := 10
		if logDate == "" {
			logDate = "No date"
		} else if len(logDate) > 1 && logDate[colon] == ':' {
			logDate = logDate[0:hour] + " " + logDate[hour:]
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
			return models.NavigateMsg{View: models.LogDetailsView, State: models.VideoRouteState{VideoID: selectedVideo.ID}}
		}
	}
	return m, nil
}

func (m LogListModel) View() string {
	var s strings.Builder

	s.WriteString(ui.HeaderStyle.Render("video logs") + "\n")

	if len(m.videos) == 0 {
		s.WriteString("\t\t\tno videos logged yet\n\n")
		// Add help even when no videos
		keymap := LogListKeyMap{}
		s.WriteString(m.help.View(keymap))
		return s.String()
	}

	// add red border when focused
	currentSearchStyle := ui.SearchStyle
	if m.focused {
		currentSearchStyle = ui.SearchStyle.BorderForeground(ui.PrimaryColor)
	}

	s.WriteString("\n" + currentSearchStyle.Render(m.search.View()) + "\n")

	tableContent := m.table.View()
	styledTable := ui.TableStyle.Render(tableContent)
	s.WriteString("\n" + styledTable)

	// Add help at the bottom
	keymap := LogListKeyMap{}
	s.WriteString("\n\n" + m.help.View(keymap))

	return s.String()
}
