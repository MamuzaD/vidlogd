package views

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mamuzad/vidlogd/internal/models"
	"github.com/mamuzad/vidlogd/internal/ui"
)

type SyncActionType int

const (
	SyncViewStatus SyncActionType = iota
	SyncRun
	SyncOpenGitTools
	SyncOpenSettings
)

type SyncActionItem struct {
	actionType  SyncActionType
	title       string
	description string
}

type SyncActionItemDelegate struct{}

func (i SyncActionItem) FilterValue() string                               { return i.title }
func (d SyncActionItemDelegate) Height() int                               { return 2 }
func (d SyncActionItemDelegate) Spacing() int                              { return 1 }
func (d SyncActionItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d SyncActionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SyncActionItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = ui.MenuItemStyle.Background(ui.PrimaryColor).Foreground(ui.White)
	} else {
		titleStyle = ui.MenuItemStyle
	}

	title := titleStyle.Render(i.title)

	line1 := title
	line2 := ui.DescriptionStyle.Render(i.description)
	fmt.Fprint(w, line1+"\n"+line2)
}

type BackupModel struct {
	list      list.Model
	state     SyncState
	statusMsg string
	spinner   spinner.Model
}

func NewBackupModel() BackupModel {
	items := []list.Item{
		SyncActionItem{
			actionType:  SyncViewStatus,
			title:       "status",
			description: "view repo status and connection info",
		},
		SyncActionItem{
			actionType:  SyncRun,
			title:       "sync",
			description: "intelligent sync: pull then push (fast-forward only)",
		},
		SyncActionItem{
			actionType:  SyncOpenGitTools,
			title:       "lazygit",
			description: "open lazygit for manual git operations",
		},
		SyncActionItem{
			actionType:  SyncOpenSettings,
			title:       "settings",
			description: "edit remote url, auto sync",
		},
	}

	const defaultWidth = 55
	const listHeight = 16

	s := spinner.New()
	s.Spinner = spinner.Jump

	l := list.New(items, SyncActionItemDelegate{}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowHelp(true)

	return BackupModel{
		list:      l,
		spinner:   s,
		state:     SyncIdle,
		statusMsg: "",
	}
}

func (m BackupModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m BackupModel) Update(msg tea.Msg) (BackupModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case SyncStatusMsg:
		m.state = msg.State
		m.statusMsg = msg.Message
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, ui.GlobalKeyMap.Back) {
			return m, func() tea.Msg {
				return models.NavigateMsg{View: models.MainMenuView}
			}
		}
		if key.Matches(msg, ui.GlobalKeyMap.Select, ui.GlobalKeyMap.Right) {
			return m.handleSelection()
		}
	}

	var cmds []tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m BackupModel) handleSelection() (BackupModel, tea.Cmd) {
	selectedItem, ok := m.list.SelectedItem().(SyncActionItem)
	if !ok {
		return m, nil
	}

	switch selectedItem.actionType {
	case SyncViewStatus:
		m.state = SyncLoading
		m.statusMsg = "checking repository status..."
		// simulate a slow loading state
		return m, tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
			return m.getBackupStatusCmd()
		})
	case SyncRun:
		m.state = SyncLoading
		m.statusMsg = "synchronizing repository..."
		return m, m.performSmartSyncCmd
	case SyncOpenGitTools:
		return m, m.openGitToolsCmd()
	case SyncOpenSettings:
		return m, func() tea.Msg {
			return models.NavigateMsg{
				View:  models.SettingsView,
				State: models.SettingsRouteState{ListIndex: int(BackupRepoEditor)},
			}
		}
	}

	return m, nil
}

func (m BackupModel) View() string {
	header := ui.HeaderStyle.Render("sync & backup")

	repoLine := fmt.Sprintf("󰊢 repo: %s", renderBackupRepo())
	autoLine := fmt.Sprintf("󰓦 auto sync: %s", getBoolString(Settings.AutoSync))
	backupPathLine := fmt.Sprintf("󰉋 backup path: %s", renderBackupPath())

	configBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.Gray).
		Padding(0, 1).
		Width(max(56, m.list.Width()-2)).
		Render(repoLine + "\n" + autoLine + "\n" + backupPathLine)

	statusBox := ""
	if strings.TrimSpace(m.statusMsg) != "" {
		borderColor := ui.Gray
		switch m.state {
		case SyncLoading:
			borderColor = lipgloss.Color(ui.Orange)
		case SyncSuccess:
			borderColor = lipgloss.Color(ui.Green)
		case SyncError:
			borderColor = lipgloss.Color(ui.Red)
		}

		body := m.statusMsg
		title := "status"
		if m.state == SyncLoading {
			title = fmt.Sprintf("%s loading", m.spinner.View())
		}

		statusBox = "\n\n" + lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(max(56, m.list.Width()-2)).
			Render(lipgloss.NewStyle().Bold(true).Render(title)+"\n"+body)
	}

	menuBox := "\n\n" + m.list.View()
	content := header + "\n\n" + configBox + statusBox + "\n\n" + menuBox
	return ui.CenterHorizontally(content, m.list.Width())
}

// --------- status + animation ---------

type SyncState int

const (
	SyncIdle SyncState = iota
	SyncLoading
	SyncSuccess
	SyncError
)

type SyncStatusMsg struct {
	State   SyncState
	Message string
}

func (m BackupModel) getBackupStatusCmd() tea.Msg {
	return SyncStatusMsg{State: SyncIdle, Message: "not implemented"}
}

func (m BackupModel) performSmartSyncCmd() tea.Msg {
	return SyncStatusMsg{State: SyncSuccess, Message: "todo: sync completed successfully"}
}

func (m BackupModel) openGitToolsCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg { return SyncStatusMsg{State: SyncError, Message: "todo: open lazygit or git"} })
}

func renderBackupPath() string {
	dataDir, err := models.DataDir()
	if err != nil {
		return "unknown"
	}
	p := filepath.Join(dataDir, "backup")

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		if strings.HasPrefix(p, home+string(os.PathSeparator)) || p == home {
			return "~" + strings.TrimPrefix(p, home)
		}
	}

	return p
}
