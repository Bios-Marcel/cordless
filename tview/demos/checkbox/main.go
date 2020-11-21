// Demo code for the Checkbox primitive.
package main

import (
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
)

func main() {
	app := tview.NewApplication(config.Current.VimEnabled)
	checkbox := tview.NewCheckbox().SetLabel("Hit Enter to check box: ")
	if err := app.SetRoot(checkbox, true).Run(); err != nil {
		panic(err)
	}
}
