package shortcutdialog

import (
	tcell "github.com/gdamore/tcell/v2"

	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/ui/components"
	"github.com/Bios-Marcel/cordless/windowman"
)

func newShortcutView(setFocus func(tview.Primitive) error, onClose func()) (tview.Primitive, tview.Primitive) {
	var table *ShortcutTable
	var shortcutDescription *components.BottomBar
	var exitButton *tview.Button
	var resetButton *tview.Button

	table = NewShortcutTable()
	table.SetShortcuts(shortcuts.Shortcuts)

	exitButton = tview.NewButton("Go back")
	if onClose != nil {
		exitButton.SetSelectedFunc(onClose)
	}
	exitButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			setFocus(table.GetPrimitive())
		} else if event.Key() == tcell.KeyBacktab {
			setFocus(resetButton)
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
			setFocus(exitButton)
		} else if event.Key() == tcell.KeyBacktab {
			setFocus(table.GetPrimitive())
		} else {
			return event
		}

		return nil
	})

	shortcutDescription = components.NewBottomBar()
	shortcutDescription.SetBorderPadding(1, 0, 0, 0)

	shortcutDescription.AddItem("R - Reset shortcut")
	shortcutDescription.AddItem("Backspace - Delete shortcut")
	shortcutDescription.AddItem("Enter - Change shortcut")
	shortcutDescription.AddItem("ESC - Close dialog")

	table.SetFocusNext(func() { setFocus(resetButton) })
	table.SetFocusPrevious(func() { setFocus(exitButton) })

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

		if event.Key() == tcell.KeyESC && onClose != nil {
			onClose()
			return nil
		}

		return event
	})
	shortcutsView.AddItem(table.GetPrimitive(), 0, 1, false)
	shortcutsView.AddItem(buttonBar, 1, 0, false)
	shortcutsView.AddItem(shortcutDescription, 2, 0, false)
	return shortcutsView, table.GetPrimitive()
}

func ShowShortcutsDialog(app *tview.Application, onClose func()) {
	shortcutsView, table := newShortcutView(func(primitive tview.Primitive) error {
		app.SetFocus(primitive)
		return nil
	}, onClose)
	app.SetRoot(shortcutsView, true)
	app.SetFocus(table)
}

func NewShortcutsDialog() *ShortcutDialog {
	window := NewShortcutWindow()
	return &ShortcutDialog{
		ShortcutWindow: *window,
	}
}

type ShortcutDialog struct {
	ShortcutWindow

	closer windowman.DialogCloser
} 

func (sd *ShortcutDialog) Open(close windowman.DialogCloser) error {
	sd.closer = close
	return nil
}

func (sd *ShortcutDialog) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		sd.closer()
		return nil
	}

	return sd.ShortcutWindow.HandleKeyEvent(event)
}

type ShortcutWindow struct {
	windowman.Window

	root       tview.Primitive
	setFocus   func(tview.Primitive) error
	focusFirst tview.Primitive
	focussed   tview.Primitive
}

// Show resets the window state and returns the tview.Primitive that the caller should show.
// The setFocus argument is used by the Window to change the focus
func (sw *ShortcutWindow) Show(appCtl windowman.ApplicationControl) error {
	appCtl.SetRoot(sw.root, true)
	sw.setFocus = func(primitive tview.Primitive) error {
		appCtl.SetFocus(primitive)
		sw.focussed = primitive
		return nil
	}
	return sw.setFocus(sw.focusFirst)
}

func (sw *ShortcutWindow) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	handler := sw.focussed.InputHandler()
	return handler(event, func(p tview.Primitive) {
		sw.setFocus(p)
	})
}

func NewShortcutWindow() *ShortcutWindow {
	shortcutWindow := &ShortcutWindow{}
	//FIXME Shortcuts view doesn't close on ESC anymore, this needs to be solved.
	//A solution might be to tell the window manager and therefore the application
	//to initiate a shutdown.
	shortcutsView, table := newShortcutView(func(primitive tview.Primitive) error {
		shortcutWindow.setFocus(primitive)
		return nil
	}, nil)
	shortcutWindow.root = shortcutsView
	shortcutWindow.focusFirst = table
	return shortcutWindow
}
