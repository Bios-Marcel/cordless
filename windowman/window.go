package windowman

import (
	"github.com/gdamore/tcell"
)

type Window interface {
	// Show resets the window state and returns the tview.Primitive that the caller should show.
	// The setFocus argument is used by the Window to change the focus
	Show(ApplicationControl) error

	HandleKeyEvent(*tcell.EventKey) *tcell.EventKey
}
