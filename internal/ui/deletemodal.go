package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mamuzad/vidlogd/internal/models"
)

type DeleteConfirmMsg struct {
	TargetID string
}

type DeleteCancelMsg struct{}

type DeleteModal struct {
	Visible bool
	Target  *models.Video
}

func NewDeleteModal() DeleteModal {
	return DeleteModal{}
}

func (m *DeleteModal) Show(target *models.Video) {
	m.Visible = true
	m.Target = target
}

func (m *DeleteModal) Hide() {
	m.Visible = false
	m.Target = nil
}

func (m *DeleteModal) Update(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if !m.Visible {
		return false, nil
	}

	switch {
	case key.Matches(msg, GlobalKeyMap.Yes):
		targetID := ""
		if m.Target != nil {
			targetID = m.Target.ID
		}
		m.Hide()
		return true, func() tea.Msg { return DeleteConfirmMsg{TargetID: targetID} }

	default:
		m.Hide()
		return true, func() tea.Msg { return DeleteCancelMsg{} }

	}
}

func (m DeleteModal) View(width int, py int, px int) string {
	if !m.Visible || m.Target == nil {
		return ""
	}

	title := m.Target.Title
	yesHelp := GlobalKeyMap.Yes.Help().Key
	noHelp := GlobalKeyMap.No.Help().Key
	body := ModalStyle.Padding(py, px).Render(
		DangerStyle.Render("Confirm delete") + "\n\n" +
			fmt.Sprintf("Delete \"%s\"?\n\n", title) +
			DescriptionStyle.Render(fmt.Sprintf("%s: delete   %s: cancel", yesHelp, noHelp)),
	)

	if width <= 0 {
		width = 60
	}

	return "\n\n" + CenterHorizontally(body, width)
}
