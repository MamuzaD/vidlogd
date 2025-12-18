package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/mamuzad/vidlogd/internal/models"
)

// Global keymap instance
var GlobalKeyMap KeyMap

// Initialize global keymap
func InitKeyMap() {
	// Load settings to get vim preference
	UpdateKeyMap()
}

// Update keymap based on current VimMotions setting
func UpdateKeyMap() {
	settings := models.LoadSettings()
	GlobalKeyMap = NewKeyMap(settings.VimMotions)
}

type KeyMap struct {
	// global actions
	Exit key.Binding
	Back key.Binding
	Help key.Binding // for toggling help

	// navigation
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding

	// specific actions
	Edit       key.Binding
	Delete     key.Binding
	Save       key.Binding
	Cancel     key.Binding
	Search     key.Binding
	SearchBack key.Binding

	// form navigation
	NextField key.Binding
	PrevField key.Binding

	// stat navigation
	Cycle     key.Binding
	CycleBack key.Binding

	// vim-specific
	InsertMode key.Binding
	NormalMode key.Binding
	VisualMode key.Binding
	Paste      key.Binding
	Yank       key.Binding

	// rating control
	RatingUp   key.Binding
	RatingDown key.Binding

	// rating input (for form)
	Rating     key.Binding // HACK: show rating keys as single line
	RatingHalf key.Binding // for adding 0.5
}

func NewKeyMap(useVim bool) KeyMap {
	// base keymap
	km := KeyMap{
		Exit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "exit")),
		Back: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "back")),
		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),

		// navigation
		Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),

		// stat navigation
		Cycle:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle")),
		CycleBack: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "cycle back")),

		// common actions (include space for select)
		Select:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Edit:       key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:     key.NewBinding(key.WithKeys("x", "d"), key.WithHelp("x/d", "delete")),
		Save:       key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Search:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		SearchBack: key.NewBinding(key.WithKeys("esc")),

		// rating number inputs
		Rating: key.NewBinding(
			key.WithKeys("0", "1", "2", "3", "4", "5"),
			key.WithHelp("0-5", "set rating"),
		),
		RatingHalf: key.NewBinding(key.WithKeys("."), key.WithHelp(".", "add 0.5")),
	}

	if useVim {
		km.Cancel = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "cancel"))

		// vim-specific form navigation
		km.NextField = key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "next field"))
		km.PrevField = key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "prev field"))

		// vim modes
		km.InsertMode = key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "insert mode"))
		km.NormalMode = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "normal mode"))
		km.VisualMode = key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "visual mode"))
		km.Paste = key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "paste"))
		km.Yank = key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yank"))

		// vim-style rating control
		km.RatingUp = key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "increase rating"))
		km.RatingDown = key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "decrease rating"))
	} else {
		km.Cancel = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

		// standard form navigation
		km.NextField = key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("↓/tab", "next field"))
		km.PrevField = key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("↑/shift+tab", "prev field"))

		// no vim modes in standard mode
		km.InsertMode = key.NewBinding(key.WithKeys(), key.WithHelp("", ""))
		km.NormalMode = key.NewBinding(key.WithKeys(), key.WithHelp("", ""))
		km.VisualMode = key.NewBinding(key.WithKeys(), key.WithHelp("", ""))
		km.Paste = key.NewBinding(key.WithKeys(), key.WithHelp("", ""))
		km.Yank = key.NewBinding(key.WithKeys(), key.WithHelp("", ""))

		// standard rating control
		km.RatingUp = key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "increase rating"))
		km.RatingDown = key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "decrease rating"))
	}

	return km
}
