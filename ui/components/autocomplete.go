package components

import (
	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

// AutocompleteView is a simple treeview meant for displaying autocomplete
// choices. The speccial part about this component is, that it can redirect
// certain events to a different component, as only certain events are
// treated directly.
type AutocompleteView struct {
	*tview.TreeView
}

// NewAutocompleteView creates a ready-to-use AutocompleteView.
func NewAutocompleteView() *AutocompleteView {
	treeView := tview.NewTreeView()
	//Has to be disabled, as we need the events for the related editor.
	treeView.
		SetSearchOnTypeEnabled(false).
		SetVimBindingsEnabled(false).
		SetTopLevel(1).
		SetCycleSelection(true)

	return &AutocompleteView{treeView}
}

// InputHandler returns the handler for this primitive.
func (a *AutocompleteView) InputHandler() tview.InputHandlerFunc {
	return a.WrapInputHandler(a.DefaultInputHandler)
}

// WrapInputHandler unlike Box.WrapInputHandler calls the default handler
// first, as all other shortcuts are meant to be forwarded. However, not
// all events are handled by the TreeView, as some features aren't desired.
func (a *AutocompleteView) WrapInputHandler(inputHandler tview.InputHandlerFunc) tview.InputHandlerFunc {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		switch key := event.Key(); key {
		//FIXME Maybe this should be made configurable as well
		case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgDn, tcell.KeyPgUp,
			tcell.KeyEnter, tcell.KeyHome, tcell.KeyEnd:
			if inputHandler != nil {
				event = inputHandler(event, setFocus)
			}
		}

		if event != nil {
			inputCapture := a.GetInputCapture()
			if inputCapture != nil {
				event = inputCapture(event)
			}
		}

		return event
	}
}
