// Demo code for the Frame primitive.
package main

import (
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

func main() {
	app := tview.NewApplication(config.Current.VimEnabled)
	frame := tview.NewFrame(tview.NewBox().SetBackgroundColor(tcell.ColorBlue)).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Header left", true, tview.AlignLeft, tcell.ColorWhite).
		AddText("Header middle", true, tview.AlignCenter, tcell.ColorWhite).
		AddText("Header right", true, tview.AlignRight, tcell.ColorWhite).
		AddText("Header second middle", true, tview.AlignCenter, tcell.ColorRed).
		AddText("Footer middle", false, tview.AlignCenter, tcell.ColorGreen).
		AddText("Footer second middle", false, tview.AlignCenter, tcell.ColorGreen)
	if err := app.SetRoot(frame, true).Run(); err != nil {
		panic(err)
	}
}
