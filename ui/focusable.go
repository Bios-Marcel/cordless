package ui

import (
	"github.com/Bios-Marcel/cordless/tview"
)

// Focusable UI elements can be given focus
type Focusable interface {
	// Focus tells a UI component to change the focus
	// of the given application to itself
	SetFocus(app *tview.Application)
}
