package windowman

import (
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

var (
	windowManagerSingleton WindowManager = nil
)

type WindowManager interface {
	Show(window Window) error
	Dialog(dialog Dialog) error
	Run(window Window) error
}

type concreteWindowManager struct {
	tviewApp tview.Application
}

func GetWindowManager() WindowManager {
	if windowManagerSingleton == nil {
		windowManagerSingleton = newWindowManager()
	}

	return windowManagerSingleton
}

func newWindowManager() WindowManager {
	wm := &concreteWindowManager{
		tviewApp: tview.NewApplication(),
	}

	// WindowManager sets the root input handler.
	// It captures exit application shortcuts, and exits the application,
	// or otherwise allows the event to bubble down.
	wm.tviewApp.SetInputHandler(func(event *tcell.EventKey) *tcell.EventKey {
		if shortcuts.ExitApplication.Equals(event) {
			wm.tviewApp.Stop()
			return nil
		}
		return event
	})
}

func (wm *concreteWindowManager) Show(window Window) error {
	primitive := window.Show(createSetFocusCallback(wm.tviewApp))
	wm.tviewApp.SetRoot(primitive, true)
}
func (wm *concreteWindowManager) Dialog(dialog Dialog) error {
	panic("not implemented")
}
func (wm *concreteWindowManager) Run(window Window) error {
	err := wm.Show(window)
	if err != nil {
		return err
	}

	return wm.tviewApp.Run()
}

func createSetFocusCallback(app *tview.Application) Focusser {
	return func(primitive tview.Primtive) error {
		app.SetFocus(primitive)
		return nil
	}
}
