package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mamuzad/vidlogd/internal/models"
)

type DeleteConfirmMsg struct{}

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

	switch msg.String() {
	case "y", "Y":
		m.Hide()
		return true, func() tea.Msg { return DeleteConfirmMsg{} }
	case "n", "N", "esc", "q", "Q":
		m.Hide()
		return true, func() tea.Msg { return DeleteCancelMsg{} }
	default:
		// swallow all other keys while the modal is open
		return true, nil
	}
}

func (m DeleteModal) View(width int, padding int) string {
	if !m.Visible || m.Target == nil {
		return ""
	}

	title := m.Target.Title
	body := ModalStyle.Padding(padding, 4).Render(
		DangerStyle.Render("Confirm delete") + "\n\n" +
			fmt.Sprintf("Delete \"%s\"?\n\n", title) +
			DescriptionStyle.Render("y: delete   esc/n: cancel"),
	)

	if width <= 0 {
		width = 60
	}

	return "\n\n" + CenterHorizontally(body, width)
}
