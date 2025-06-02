package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MainMenuView ViewType = iota
	LogVideoView
)

type Model struct {
	currentView ViewType

	mainMenu MainMenuModel
	logVideo LogVideoModel
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("VidLogd")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.currentView != MainMenuView {
				m.currentView = MainMenuView
				return m, nil
			}
		}

	case NavigateMsg:
		m.currentView = msg.View
		return m, nil
	}

	switch m.currentView {
	case MainMenuView:
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case LogVideoView:
		m.logVideo, cmd = m.logVideo.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	var content string

	switch m.currentView {
	case MainMenuView:
		content = m.mainMenu.View()
	case LogVideoView:
		content = m.logVideo.View()
	}

	return content
}

func main() {
	m := Model{
		currentView: MainMenuView,
		mainMenu:    NewMainMenuModel(),
		logVideo:    NewLogVideoModel(),
	}

	p := tea.
		NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
