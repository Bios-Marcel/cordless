package windowman

import "github.com/Bios-Marcel/cordless/tview"

// Focusser allows the caller to instruct the application to move the focus to the given primitive
type Focusser func(tview.Primtive) error

type Window interface {
	// Show resets the window state and returns the tview.Primitive that the caller should show.
	// The setFocus argument is used by the Window to change the focus
	Show(setFocus Focusser) tview.Primitive
}
