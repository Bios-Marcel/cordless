package tviewutil

import (
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

func CreateFocusTextViewOnTypeInputHandler(app *tview.Application, component *tview.TextView) func(event *tcell.EventKey) *tcell.EventKey {
	eventHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModNone {
			if event.Key() == tcell.KeyEnter {
				return event
			}

			if event.Rune() != 0 {
				inputHandler := component.InputHandler()
				if inputHandler != nil {
					app.SetFocus(component)
					inputHandler(event, nil)
					return nil
				}
			}
		}

		return event
	}

	return eventHandler
}

// GetNodeByReference returns the first matched node where the given reference
// is equal. If no node with a matching reference exists, the return value
// is nil.
func GetNodeByReference(reference interface{}, tree *tview.TreeView) *tview.TreeNode {
	var matchedNode *tview.TreeNode
	tree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		if node.GetReference() == reference {
			matchedNode = node
			return false
		}

		return true
	})

	return matchedNode
}
