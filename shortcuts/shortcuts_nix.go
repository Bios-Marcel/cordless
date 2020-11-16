// +build !windows

package shortcuts

import tcell "github.com/gdamore/tcell/v2"

func addDeleteLeftShortcut() *Shortcut {
	// FIXME don't know what to do with this vim binding
	return addShortcut("delete_left", "Delete left", multilineTextInput, tcell.NewEventKey(tcell.KeyBackspace2, rune(tcell.KeyBackspace2), tcell.ModNone),addVimEvent(NullVimEvent,nil,NullVimEvent))
}
