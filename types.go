package main

type ViewType int

type NavigateMsg struct {
	View ViewType // The target view to switch to
}
