package tview

import "github.com/gdamore/tcell"

// MouseSupport defines wether a component supports accepting mouse events
type MouseSupport interface {
	MouseHandler() func(event *tcell.EventMouse) bool
}
