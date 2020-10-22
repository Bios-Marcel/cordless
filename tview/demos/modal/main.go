// Demo code for the Modal primitive.
package main

import (
	"strings"

	tcell "github.com/gdamore/tcell/v2"

	"github.com/Bios-Marcel/cordless/tview"
)

func main() {
	app := tview.NewApplication()

	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, false).
				AddItem(nil, 0, 1, false), width, 1, false).
			AddItem(nil, 0, 1, false)
	}

	background := tview.NewTextView().
		SetTextColor(tcell.ColorBlue).
		SetText(strings.Repeat("background ", 1000))

	box := tview.NewBox().
		SetBorder(true).
		SetTitle("Centered Box")

	pages := tview.NewPages().
		AddPage("background", background, true, true).
		AddPage("modal", modal(box, 40, 10), true, true)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
