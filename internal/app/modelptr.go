package app

import tea "github.com/charmbracelet/bubbletea"

// ensurePtr initializes *p with ctor() if nil.
func ensurePtr[T any](p **T, ctor func() T) {
	if *p != nil {
		return
	}
	v := ctor()
	*p = &v
}

// updatePtr ensures *p is non-nil, calls Update, stores the updated value back
// into the existing pointer, and returns the resulting command.
type updatable[T any] interface {
	Update(tea.Msg) (T, tea.Cmd)
}

func updatePtr[T updatable[T]](p **T, msg tea.Msg, ctor func() T) tea.Cmd {
	ensurePtr(p, ctor)
	v, cmd := (**p).Update(msg)
	**p = v
	return cmd
}
