package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type MenuItem struct {
	title string
}

func (i MenuItem) FilterValue() string { return i.title }

// necessary for list
type MenuItemDelegate struct{}

func (d MenuItemDelegate) Height() int                               { return 1 }
func (d MenuItemDelegate) Spacing() int                              { return 0 }
func (d MenuItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d MenuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	style := menuItemStyle
	if index == m.Index() {
		style = style.Background(primaryColor).Foreground(white)
	}

	styledText := style.Render(i.title)
	centeredText := centerHorizontally(styledText, m.Width())
	fmt.Fprint(w, centeredText)
}

type MainMenuModel struct {
	list list.Model
}

func NewMainMenuModel() MainMenuModel {
	items := []list.Item{
		MenuItem{title: "log video"},
		MenuItem{title: "view logs"},
		MenuItem{title: "exit"},
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, MenuItemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)

	return MainMenuModel{
		list: l,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (MainMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter", " ":
			return m.handleSelection()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MainMenuModel) handleSelection() (MainMenuModel, tea.Cmd) {
	selectedItem, ok := m.list.SelectedItem().(MenuItem)
	if !ok {
		return m, nil
	}

	switch selectedItem.title {
	case "log video":
		return m, func() tea.Msg {
			return NavigateMsg{View: LogVideoView}
		}
	case "view logs":
		return m, func() tea.Msg {
			return NavigateMsg{View: LogListView}
		}
	case "exit":
		return m, tea.Quit
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	width := m.list.Width()
	if width == 0 {
		width = 80 // fallback to default width
	}

	title := centerHorizontally(titleStyle.Render("vidlogd"), width)
	return title + "\n\n" + centerHorizontally(m.list.View(), width)
}
