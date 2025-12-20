package views

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/ui"
)

type MenuItem struct {
	title string
}

// necessary for list
type MenuItemDelegate struct{}

func (i MenuItem) FilterValue() string                               { return i.title }
func (d MenuItemDelegate) Height() int                               { return 1 }
func (d MenuItemDelegate) Spacing() int                              { return 0 }
func (d MenuItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d MenuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(MenuItem)
	if !ok {
		return
	}

	style := ui.MenuItemStyle
	if index == m.Index() {
		style = style.Background(ui.PrimaryColor).Foreground(ui.White)
	}

	styledText := style.Render(i.title)
	centeredText := ui.CenterHorizontally(styledText, m.Width())
	fmt.Fprint(w, centeredText)
}

type MainMenuModel struct {
	list list.Model
}

func NewMainMenuModel() MainMenuModel {
	items := []list.Item{
		MenuItem{title: "log video"},
		MenuItem{title: "view logs"},
		MenuItem{title: "stats"},
		MenuItem{title: "sync"},
		MenuItem{title: "settings"},
		MenuItem{title: "exit"},
	}

	const defaultWidth = 40
	const listHeight = 14

	l := list.New(items, MenuItemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(true)
	l.KeyMap.Quit.SetKeys()
	l.KeyMap.Quit.SetHelp("", "")

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
		if key.Matches(msg, ui.GlobalKeyMap.Back) {
			return m, tea.Quit
		}
		if key.Matches(msg, ui.GlobalKeyMap.Select) {
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
			return models.NavigateMsg{View: models.LogVideoView}
		}
	case "view logs":
		return m, func() tea.Msg {
			return models.NavigateMsg{View: models.LogListView}
		}
	case "stats":
		return m, func() tea.Msg {
			return models.NavigateMsg{View: models.StatsView}
		}
	case "sync":
		return m, func() tea.Msg {
			return models.NavigateMsg{View: models.BackupView}
		}
	case "settings":
		return m, func() tea.Msg {
			return models.NavigateMsg{View: models.SettingsView}
		}
	case "exit":
		return m, tea.Quit
	}
	return m, nil
}

func (m MainMenuModel) View() string {
	width := m.list.Width()

	return ui.CenterHorizontally(m.list.View(), width)
}
