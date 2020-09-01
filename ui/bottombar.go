package ui

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/ui/shortcutdialog"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// BottomBar custom simple component to render static information at the bottom
// of the application.
type BottomBar struct {
	*tview.Box
	items []*bottomBarItem
}

type bottomBarItem struct {
	name    string
	content string
}

// Draw draws this primitive onto the screen. Implementers can call the
// screen's ShowCursor() function but should only do so when they have focus.
// (They will need to keep track of this themselves.)
func (b *BottomBar) Draw(screen tcell.Screen) bool {
	hasDrawn := b.Box.Draw(screen)
	if !hasDrawn {
		return false
	}

	style := tcell.StyleDefault.
		//Background(config.GetTheme().PrimitiveBackgroundColor).
		Foreground(config.GetTheme().PrimaryTextColor).
		Reverse(true)

	xPos, yPos, _, _ := b.GetInnerRect()
	for _, item := range b.items {
		gr := uniseg.NewGraphemes(item.content)
		for gr.Next() {
			r := gr.Runes()
			width := runewidth.StringWidth(gr.Str())
			var comb []rune
			if len(r) > 1 {
				comb = r[1:]
			}

			screen.SetContent(xPos, yPos, r[0], comb, style)
			xPos += width
		}

		//Spacing between items
		xPos++
	}

	return true
}

// NewBottomBar creates a new bar to be put at the bottom aplication.
// It contains static information and hints.
func NewBottomBar(username string) *BottomBar {
	loggedInAsText := fmt.Sprintf("Logged in as: '%s'", tviewutil.Escape(username))
	shortcutInfoText := fmt.Sprintf("View / Change shortcuts: %s", shortcutdialog.EventToString(shortcutsDialogShortcut))

	bottomBar := &BottomBar{
		Box: tview.NewBox(),
		items: []*bottomBarItem{
			{
				name:    "username",
				content: loggedInAsText,
			},
			{
				name:    "shortcut-info",
				content: shortcutInfoText,
			},
		},
	}
	bottomBar.SetBorder(false)

	return bottomBar
}
