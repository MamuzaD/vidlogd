package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type MainMenuModel struct {
	cursor  int
	choices []string
}

func NewMainMenuModel() MainMenuModel {
	return MainMenuModel{
		cursor: 0,
		choices: []string{
			"log video",
			"exit",
		},
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (MainMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.handleSelection()
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainMenuModel) handleSelection() (MainMenuModel, tea.Cmd) {
	switch m.cursor {
	case 0:
		return m, func() tea.Msg {
			return NavigateMsg{View: LogVideoView}
		}
	case 1:
		return m, tea.Quit
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	var s strings.Builder

	s.WriteString("Welcome to VidLogd!\n\n")
	s.WriteString("What would you like to do?\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
		}

		s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	s.WriteString("\n")

	instructions := "↑/↓ (or j/k) to navigate, Enter to select, 'q' to quit"
	s.WriteString(instructions + "\n")

	return s.String()
}
