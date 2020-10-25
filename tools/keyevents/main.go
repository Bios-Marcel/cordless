package main

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

func main() {
	input := tview.NewTextView()
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fmt.Fprintln(input, event.Key(), event.Modifiers(), event.Rune())

		return event
	})
	tview.NewApplication().SetRoot(input, true).Run()
}
