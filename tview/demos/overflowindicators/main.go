package main

import (
	"github.com/Bios-Marcel/cordless/tview"
	"strings"
)

func main() {
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetScrollable(true).
		SetText(strings.Repeat("OwO\n", 100)).
		SetIndicateOverflow(true).
		SetBorderSides(true, true, true, true).
		SetBorder(true).
		SetBorderPadding(0, 0, 0, 10)
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewTextView(), 1, -1, false)
	flex.AddItem(textView, 10, -1, false)

	root := tview.NewTreeNode("Root")
	root.AddChild(tview.NewTreeNode("A"))
	root.AddChild(tview.NewTreeNode("B"))
	root.AddChild(tview.NewTreeNode("C"))
	root.AddChild(tview.NewTreeNode("D"))
	root.AddChild(tview.NewTreeNode("E"))
	root.AddChild(tview.NewTreeNode("F"))
	tree := tview.NewTreeView().SetRoot(root)
	tree.SetBorder(true).SetIndicateOverflow(true)
	flex.AddItem(tree, 6, -1, false)

	list := tview.NewList()
	list.AddItem("A", "", 0, nil)
	list.AddItem("B", "", 0, nil)
	list.AddItem("C", "", 0, nil)
	list.AddItem("D", "", 0, nil)
	list.AddItem("E", "", 0, nil)
	list.AddItem("F", "", 0, nil)
	list.SetBorder(true).SetIndicateOverflow(true)
	list.ShowSecondaryText(false)
	flex.AddItem(list, 6, -1, true)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
