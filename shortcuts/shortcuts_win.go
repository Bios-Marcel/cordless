// +build windows

package shortcuts

import tcell "github.com/gdamore/tcell/v2"

var DeleteLeft = addShortcut("delete_left", "Delete left", multilineTextInput, tcell.NewEventKey(tcell.KeyBackspace, rune(tcell.KeyBackspace), tcell.ModNone))
