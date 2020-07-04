package tviewutil

import (
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

func CreateFocusTextViewOnTypeInputHandler(app *tview.Application, component *tview.TextView) func(event *tcell.EventKey) *tcell.EventKey {
	eventHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModNone {
			if event.Key() == tcell.KeyEnter {
				return event
			}

			if event.Rune() != 0 {
				inputHandler := component.InputHandler()
				if inputHandler != nil {
					app.SetFocus(component)
					inputHandler(event, nil)
					return nil
				}
			}
		}

		return event
	}

	return eventHandler
}
