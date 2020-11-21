// Demo code for the InputField primitive.
package main

import (
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

func main() {
	app := tview.NewApplication(config.Current.VimEnabled)
	inputField := tview.NewInputField().
		SetLabel("Enter a number: ").
		SetPlaceholder("E.g. 1234").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			app.Stop()
		})
	if err := app.SetRoot(inputField, true).Run(); err != nil {
		panic(err)
	}
}
