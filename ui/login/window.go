package login

import (
	"github.com/Bios-Marcel/cordless/windowman"
	tcell "github.com/gdamore/tcell/v2"
)

type LoginWindow struct {
	windowman.Window

	LoginWindowComponent *Login
}

func NewLoginWindow(configDirectory string) LoginWindow {
	return LoginWindow{
		LoginWindowComponent: NewLogin(configDirectory),
	}
}

// Show resets the window state and returns the tview.Primitive that the caller should show.
// The setFocus argument is used by the Window to change the focus
func (lw *LoginWindow) Show(appCtl windowman.ApplicationControl) error {
	lw.LoginWindowComponent.SetAppControl(appCtl)
	appCtl.SetRoot(lw.LoginWindowComponent, true)
	return nil
}

func (lw *LoginWindow) HandleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	return event
}
