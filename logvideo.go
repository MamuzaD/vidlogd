package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	url = iota
	title
	channel
	release
	logDate
	review
	count
)

type LogVideoModel struct {
	inputs   []textinput.Model
	focused  int
	errorMsg string
}

func NewLogVideoModel() LogVideoModel {
	inputs := make([]textinput.Model, count)

	// url
	inputs[url] = textinput.New()
	inputs[url].Placeholder = "https://youtube.com/watch?v=..."
	inputs[url].Focus()
	inputs[url].CharLimit = 200
	inputs[url].Width = 60

	// title
	inputs[title] = textinput.New()
	inputs[title].Placeholder = "Video title"
	inputs[title].CharLimit = 100
	inputs[title].Width = 50

	// channel
	inputs[channel] = textinput.New()
	inputs[channel].Placeholder = "Channel name"
	inputs[channel].CharLimit = 50
	inputs[channel].Width = 50

	// release date
	inputs[release] = textinput.New()
	inputs[release].Placeholder = "YYYY-MM-DD"
	inputs[release].CharLimit = 20
	inputs[release].Width = 30

	// log date
	inputs[logDate] = textinput.New()
	inputs[logDate].Placeholder = "YYYY-MM-DD"
	inputs[logDate].SetValue(time.Now().Format("2006-01-02"))
	inputs[logDate].CharLimit = 20
	inputs[logDate].Width = 30

	// review
	inputs[review] = textinput.New()
	inputs[review].Placeholder = "write your review..."
	inputs[review].CharLimit = 500
	inputs[review].Width = 70

	return LogVideoModel{
		inputs:  inputs,
		focused: 0,
	}
}

func (m LogVideoModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LogVideoModel) Update(msg tea.Msg) (LogVideoModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return NavigateMsg{View: MainMenuView} }
		case "tab", "down":
			m.nextInput()
		case "shift+tab", "up":
			m.prevInput()
		case "enter":
			if m.focused == len(m.inputs) {
				return m.saveVideo()
			} else {
				m.nextInput()
			}
		}
	}

	// update the focused input
	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m LogVideoModel) View() string {
	var s strings.Builder

	s.WriteString("log a video\n\n")

	if m.errorMsg != "" {
		s.WriteString("Error: " + m.errorMsg + "\n\n")
	}

	labels := []string{
		"YouTube URL:",
		"Title:",
		"Channel:",
		"Video Release Date:",
		"Log Date:",
		"Review:",
	}

	// render all input fields
	for i, label := range labels {
		s.WriteString(label + "\n")

		// Add focus indicator
		if m.focused == i {
			s.WriteString(m.inputs[i].View() + "\n\n")
		} else {
			s.WriteString(m.inputs[i].View() + "\n\n")
		}
	}

	// save button
	if m.focused == len(m.inputs) {
		s.WriteString("> [Save Video] <\n\n")
	} else {
		s.WriteString("  [Save Video]\n\n")
	}

	s.WriteString("tab/↑↓ to navigate, enter to save, esc to cancel")

	return s.String()
}

// nextInput moves focus to the next input
func (m *LogVideoModel) nextInput() {
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused++
	if m.focused > len(m.inputs) {
		m.focused = 0
	}

	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

// prevInput moves focus to the previous input
func (m *LogVideoModel) prevInput() {
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}

	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs)
	}

	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

func (m LogVideoModel) saveVideo() (LogVideoModel, tea.Cmd) {
	requiredFields := []struct {
		index int
		name  string
	}{
		{url, "URL"},
		{title, "title"},
		{channel, "channel"},
		{release, "video release date"},
		{logDate, "log date"},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(m.inputs[field.index].Value()) == "" {
			m.errorMsg = field.name + " is required"
			return m, nil
		}
	}

	// TODO: save video data here
	return m, func() tea.Msg { return NavigateMsg{View: MainMenuView} }
}
