package ui

import (
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"os"
)

const splashText = `
 ██████╗ ██████╗ ██████╗ ██████╗ ██╗     ███████╗███████╗███████╗              
██╔════╝██╔═══██╗██╔══██╗██╔══██╗██║     ██╔════╝██╔════╝██╔════╝     /   \    
██║     ██║   ██║██████╔╝██║  ██║██║     █████╗  ███████╗███████╗ ████-   -████
██║     ██║   ██║██╔══██╗██║  ██║██║     ██╔══╝  ╚════██║╚════██║     \   /    
╚██████╗╚██████╔╝██║  ██║██████╔╝███████╗███████╗███████║███████║              
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚══════╝╚══════╝              `

type Login struct {
	*tview.Flex
	app                *tview.Application
	tokenInput         *tview.InputField
	tokenInputMasked   bool
	tokenInputMaskRune rune
	loginButton        *tview.Button
	tokenChannel       chan string
	messageText        *tview.TextView
	runNext            chan bool
}

// NewLogin creates a new login screen with the login components hidden by default.
func NewLogin(app *tview.Application, configDir string) *Login {
	login := &Login{
		Flex:               tview.NewFlex().SetDirection(tview.FlexRow),
		app:                app,
		tokenChannel:       make(chan string, 1),
		tokenInput:         tview.NewInputField(),
		tokenInputMasked:   true,
		tokenInputMaskRune: '*',
		loginButton:        tview.NewButton("Login"),
		messageText:        tview.NewTextView(),
	}

	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tviewutil.Escape(splashText + "\n\nConfig lies at: " + configDir))

	login.messageText.SetDynamicColors(true)

	login.tokenInput.SetVisible(false)
	login.tokenInput.SetBorder(true)
	login.tokenInput.SetFieldWidth(66)
	login.tokenInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			login.submit()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.loginButton)
			return nil
		}

		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			os.Exit(0)
			return nil
		}

		if event.Key() == tcell.KeyCtrlV {
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.tokenInput.Insert(content)
		}

		if event.Key() == tcell.KeyCtrlR {
			if login.tokenInputMasked {
				login.tokenInputMasked = false
				login.tokenInput.SetMaskCharacter(0)
			} else {
				login.tokenInputMasked = true
				login.tokenInput.SetMaskCharacter(login.tokenInputMaskRune)
			}
		}

		return event
	})

	login.loginButton.SetVisible(false)
	login.loginButton.SetSelectedFunc(func() {
		login.submit()
	})

	login.loginButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.tokenInput)
			return nil
		}

		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			os.Exit(0)
			return nil
		}

		return event
	})

	login.loginButton.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			login.submit()
			return true
		}

		return false
	})

	login.AddItem(splashScreen, 12, 0, false)
	login.AddItem(createCenteredComponent(login.messageText, 66), 0, 1, false)
	login.AddItem(createCenteredComponent(login.tokenInput, 68), 3, 0, false)
	login.AddItem(createCenteredComponent(login.loginButton, 68), 1, 0, false)
	login.AddItem(tview.NewBox(), 0, 1, false)

	return login
}

func createCenteredComponent(component tview.Primitive, width int) tview.Primitive {
	padding := tview.NewFlex().SetDirection(tview.FlexColumn)
	padding.AddItem(tview.NewBox(), 0, 1, false)
	padding.AddItem(component, width, 0, false)
	padding.AddItem(tview.NewBox(), 0, 1, false)

	return padding
}

func (login *Login) submit() {
	login.loginButton.SetVisible(false)
	login.tokenInput.SetVisible(false)
	login.messageText.SetText("Attempting to log in ...")
	login.tokenChannel <- login.tokenInput.GetText()
}

// RequestToken shows the UI components and waits til the user has entered a token.
func (login *Login) RequestToken(message string) string {
	login.tokenInput.SetMaskCharacter(login.tokenInputMaskRune)
	login.tokenInput.SetVisible(true)
	login.loginButton.SetVisible(true)
	login.messageText.SetText(message)

	login.app.SetFocus(login.tokenInput)
	//Easier, since we don't have to call QueueUpdateDraw. Without this, we won't see any oft he primitives.
	login.app.ForceDraw()

	return <-login.tokenChannel
}
