// +build !windows

package shortcuts

import tcell "github.com/gdamore/tcell/v2"

func addDeleteLeftShortcut() *Shortcut {
	return addShortcut("delete_left", "Delete left", multilineTextInput, tcell.NewEventKey(tcell.KeyBackspace2, rune(tcell.KeyBackspace2), tcell.ModNone))
}
