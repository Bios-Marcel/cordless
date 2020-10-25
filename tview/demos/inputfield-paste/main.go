package main

import (
	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
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