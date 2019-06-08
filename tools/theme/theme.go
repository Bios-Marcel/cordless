package main

import (
	"encoding/json"
	"os"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

func main() {
	theme := &config.Theme{
		Theme: &tview.Theme{
			PrimitiveBackgroundColor:    tcell.NewRGBColor(70, 70, 70),
			ContrastBackgroundColor:     tcell.NewRGBColor(104, 142, 196),
			MoreContrastBackgroundColor: tcell.NewRGBColor(79, 79, 79),
			BorderColor:                 tcell.NewRGBColor(213, 220, 229),
			BorderFocusColor:            tcell.NewRGBColor(104, 142, 196),
			TitleColor:                  tcell.ColorWhite,
			GraphicsColor:               tcell.ColorWhite,
			PrimaryTextColor:            tcell.ColorWhite,
			SecondaryTextColor:          tcell.ColorWhite,
			TertiaryTextColor:           tcell.ColorWhite,
			InverseTextColor:            tcell.NewRGBColor(104, 142, 196),
			ContrastSecondaryTextColor:  tcell.NewRGBColor(104, 142, 196),
		}}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	encoder.Encode(theme)
}
