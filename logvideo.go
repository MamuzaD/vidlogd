package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type LogVideoModel struct{}

func NewLogVideoModel() LogVideoModel {
	return LogVideoModel{}
}

func (m LogVideoModel) Init() tea.Cmd {
	return nil
}

func (m LogVideoModel) Update(msg tea.Msg) (LogVideoModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return NavigateMsg{View: MainMenuView} }
		case "enter":
			return m, func() tea.Msg { return NavigateMsg{View: MainMenuView} }
		}
	}
	return m, nil
}

func (m LogVideoModel) View() string {
	var s strings.Builder

	s.WriteString("test in log video")

	return s.String()
}
