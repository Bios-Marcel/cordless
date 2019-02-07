package treeview

import (
	"strings"
	"time"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

func CreateFocusTextViewOnTypeInputHandler(treeView *tview.Box, app *tview.Application, component *tview.TextView) func(event *tcell.EventKey) *tcell.EventKey {
	eventHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() != 0 && event.Modifiers() == tcell.ModNone {
			inputHandler := component.InputHandler()
			if inputHandler != nil {
				app.SetFocus(component)
				inputHandler(event, nil)
				return nil
			}
		}

		return event
	}

	return eventHandler
}

func CreateSearchOnTypeInuptHandler(treeView *tview.TreeView, rootNode *tview.TreeNode, jumpTime *time.Time, jumpBuffer *string) func(event *tcell.EventKey) *tcell.EventKey {
	var traversalFunction func(node *tview.TreeNode) bool
	traversalFunction = func(node *tview.TreeNode) bool {
		if len(node.GetChildren()) > 0 {
			for _, subNode := range node.GetChildren() {
				returnValue := traversalFunction(subNode)
				if returnValue {
					return true
				}
			}
		} else if node.IsSelectable() {
			if strings.HasPrefix(strings.ToLower(node.GetText()), *jumpBuffer) {
				treeView.SetCurrentNode(node)
				return true
			}
		}

		return false
	}

	eventHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if time.Since(*jumpTime) > (500 * time.Millisecond) {
			*jumpBuffer = ""
		}

		*jumpTime = time.Now()

		if event.Rune() != 0 {
			*jumpBuffer += string(event.Rune())
			*jumpBuffer = strings.ToLower(*jumpBuffer)

			for _, node := range rootNode.GetChildren() {
				stopIterating := traversalFunction(node)
				if stopIterating {
					return nil
				}
			}
		}

		return event
	}

	return eventHandler
}
