package main

import (
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

func main() {
	field := tview.NewInputField()
	field.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlV {
			field.Insert("Fotze")
			return nil
		}

		return event
	})
	app := tview.NewApplication().SetRoot(field,true)
	app.Run()
}