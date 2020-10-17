package shortcutdialog

import (
	"log"

	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

func ShowShortcutsDialog(app *tview.Application, onClose func()) {
	var table *ShortcutTable
	var shortcutDescription *tview.TextView
	var exitButton *tview.Button
	var resetButton *tview.Button

	table = NewShortcutTable()
	table.SetShortcuts(shortcuts.Shortcuts)

	exitButton = tview.NewButton("Go back")
	exitButton.SetSelectedFunc(onClose)
	exitButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(table.GetPrimitive())
		} else if event.Key() == tcell.KeyBacktab {
			app.SetFocus(resetButton)
		} else {
			return event
		}

		return nil
	})

	resetButton = tview.NewButton("Restore all defaults")
	resetButton.SetSelectedFunc(func() {
		for _, shortcut := range shortcuts.Shortcuts {
			shortcut.Reset()
		}
		shortcuts.Persist()

		table.SetShortcuts(shortcuts.Shortcuts)
	})
	resetButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(exitButton)
		} else if event.Key() == tcell.KeyBacktab {
			app.SetFocus(table.GetPrimitive())
		} else {
			return event
		}

		return nil
	})

	primitiveBGColor := tviewutil.ColorToHex(config.GetTheme().PrimitiveBackgroundColor)
	primaryTextColor := tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)

	shortcutDescription = tview.NewTextView()
	shortcutDescription.SetDynamicColors(true).SetBorderPadding(1, 0, 0, 0)
	if tview.IsVtxxx {
		shortcutDescription.SetText("R [::r]Reset shortcut[::-]" +
			"[::-]  Backspace [::r]Delete shortcut" +
			"[::-]  Enter [::r]Change shortcut" +
			"[::-]  Esc [::r]Close dialog")
	} else {
		shortcutDescription.SetText("[" + primaryTextColor + "][:" + primitiveBGColor + "]R [:" + primaryTextColor + "][" + primitiveBGColor + "]Reset shortcut" +
			"[" + primaryTextColor + "][:" + primitiveBGColor + "]  Backspace [:" + primaryTextColor + "][" + primitiveBGColor + "]Delete shortcut" +
			"[" + primaryTextColor + "][:" + primitiveBGColor + "]  Enter [:" + primaryTextColor + "][" + primitiveBGColor + "]Change shortcut" +
			"[" + primaryTextColor + "][:" + primitiveBGColor + "]  Esc [:" + primaryTextColor + "][" + primitiveBGColor + "]Close dialog")
	}
	table.SetFocusNext(func() { app.SetFocus(resetButton) })
	table.SetFocusPrevious(func() { app.SetFocus(exitButton) })

	buttonBar := tview.NewFlex()
	buttonBar.SetDirection(tview.FlexColumn)

	buttonBar.AddItem(resetButton, 0, 1, false)
	buttonBar.AddItem(tview.NewBox(), 1, 0, false)
	buttonBar.AddItem(exitButton, 0, 1, false)

	shortcutsView := tview.NewFlex()
	shortcutsView.SetDirection(tview.FlexRow)
	shortcutsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if table.IsDefiningShortcut() {
			return event
		}

		if event.Key() == tcell.KeyESC {
			onClose()
			return nil
		}

		return event
	})
	shortcutsView.AddItem(table.GetPrimitive(), 0, 1, false)
	shortcutsView.AddItem(buttonBar, 1, 0, false)
	shortcutsView.AddItem(shortcutDescription, 2, 0, false)

	app.SetRoot(shortcutsView, true)
	app.SetFocus(table.GetPrimitive())
}

func RunShortcutsDialogStandalone() {
	loadError := shortcuts.Load()
	if loadError != nil {
		log.Fatalf("Error loading shortcuts: %s\n", loadError)
	}
	app := tview.NewApplication()
	ShowShortcutsDialog(app, func() {
		app.Stop()
	})
	startError := app.Run()
	if startError != nil {
		log.Fatalf("Error launching shortcuts dialog: %s\n", startError)
	}
}
