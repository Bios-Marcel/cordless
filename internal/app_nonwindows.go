// +build !windows

package internal

import (
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

func defineColorTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.NewRGBColor(70, 70, 70)
	tview.Styles.ContrastBackgroundColor = tcell.NewRGBColor(104, 142, 196)
	tview.Styles.MoreContrastBackgroundColor = tcell.NewRGBColor(79, 79, 79)
	tview.Styles.BorderColor = tcell.NewRGBColor(213, 220, 229)
	tview.Styles.BorderFocusColor = tcell.NewRGBColor(104, 142, 196)
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorWhite
	tview.Styles.TertiaryTextColor = tcell.ColorWhite
	tview.Styles.InverseTextColor = tcell.NewRGBColor(104, 142, 196)
	tview.Styles.ContrastSecondaryTextColor = tcell.NewRGBColor(104, 142, 196)
}
