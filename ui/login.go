package ui

import (
	"errors"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"os"
	"strconv"
	"strings"
)

const splashText = `
 ██████╗ ██████╗ ██████╗ ██████╗ ██╗     ███████╗███████╗███████╗              
██╔════╝██╔═══██╗██╔══██╗██╔══██╗██║     ██╔════╝██╔════╝██╔════╝     /   \    
██║     ██║   ██║██████╔╝██║  ██║██║     █████╗  ███████╗███████╗ ████-   -████
██║     ██║   ██║██╔══██╗██║  ██║██║     ██╔══╝  ╚════██║╚════██║     \   /    
╚██████╗╚██████╔╝██║  ██║██████╔╝███████╗███████╗███████║███████║              
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚══════╝╚══════╝              `

type LoginType int

const (
	None     LoginType = 0
	Token    LoginType = 1
	Password LoginType = 2

	userAgent = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:70.0) Gecko/20100101 Firefox/70.0"
)

type Login struct {
	*tview.Flex
	app                *tview.Application
	tokenInput         *tview.InputField
	tokenInputMasked   bool
	tokenInputMaskRune rune
	usernameInput      *tview.InputField
	passwordInput      *tview.InputField
	tfaTokenInput      *tview.InputField
	loginType          LoginType

	loginTypeTokenButton    *tview.Button
	loginTypePasswordButton *tview.Button
	loginButton             *tview.Button
	sessionChannel          chan *loginAttempt
	messageText             *tview.TextView
	runNext                 chan bool

	content           *tview.Flex
	loginChoiceView   tview.Primitive
	passwordInputView tview.Primitive
	tokenInputView    tview.Primitive
}

type loginAttempt struct {
	session    *discordgo.Session
	loginError error
}

// NewLogin creates a new login screen with the login components hidden by default.
func NewLogin(app *tview.Application, configDir string) *Login {
	login := &Login{
		Flex:                    tview.NewFlex().SetDirection(tview.FlexRow),
		app:                     app,
		sessionChannel:          make(chan *loginAttempt, 1),
		tokenInput:              tview.NewInputField(),
		usernameInput:           tview.NewInputField(),
		passwordInput:           tview.NewInputField(),
		tfaTokenInput:           tview.NewInputField(),
		content:                 tview.NewFlex().SetDirection(tview.FlexRow),
		tokenInputMasked:        true,
		tokenInputMaskRune:      '*',
		loginTypeTokenButton:    tview.NewButton("Login via Authentication-Token"),
		loginTypePasswordButton: tview.NewButton("Login via E-Mail and password (Optinally Supports 2FA)"),

		loginButton: tview.NewButton("Login"),
		messageText: tview.NewTextView(),
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			os.Exit(0)
			return nil
		}

		return event
	})

	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tviewutil.Escape(splashText + "\n\nConfig lies at: " + configDir))
	login.AddItem(splashScreen, 12, 0, false)
	login.AddItem(createCenteredComponent(login.messageText, 66), 0, 1, false)

	login.messageText.SetDynamicColors(true)

	configureInputComponent(login.tokenInput)
	login.tokenInput.SetTitle("Token")
	configureInputComponent(login.usernameInput)
	login.usernameInput.SetTitle("E-Mail")
	configureInputComponent(login.passwordInput)
	login.passwordInput.SetTitle("Password")
	login.passwordInput.SetMaskCharacter('*')
	configureInputComponent(login.tfaTokenInput)
	login.tfaTokenInput.SetTitle("Two-Factor-Authentication Code")

	login.tokenInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			login.attemptLogin()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.loginButton)
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

	login.usernameInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			login.attemptLogin()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.passwordInput)
			return nil
		}

		if event.Key() == tcell.KeyCtrlV {
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.usernameInput.Insert(content)
		}

		return event
	})

	login.passwordInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			login.attemptLogin()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.tfaTokenInput)
			return nil
		}

		if event.Key() == tcell.KeyCtrlV {
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.passwordInput.Insert(content)
		}

		return event
	})

	login.tfaTokenInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			login.attemptLogin()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			login.app.SetFocus(login.loginButton)
			return nil
		}

		if event.Key() == tcell.KeyCtrlV {
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.tfaTokenInput.Insert(content)
		}

		return event
	})

	login.loginButton.SetVisible(false)
	login.loginButton.SetSelectedFunc(func() {
		login.attemptLogin()
	})

	login.loginButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			switch login.loginType {
			case None:
				panic("Was in state loginType=None even though Login-Button was visible")
			case Token:
				login.app.SetFocus(login.tokenInput)
			case Password:
				login.app.SetFocus(login.usernameInput)
			}
			return nil
		}

		return event
	})

	//FIXME Mouse won't work for some reason.

	login.loginTypeTokenButton.SetSelectedFunc(func() {
		login.loginType = Token
		login.showTokenLogin()
	})
	login.loginTypeTokenButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab || event.Key() == tcell.KeyDown ||
			event.Key() == tcell.KeyUp || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight {
			login.app.SetFocus(login.loginTypePasswordButton)
			return nil
		}

		return event
	})

	login.loginTypePasswordButton.SetSelectedFunc(func() {
		login.loginType = Password
		login.showPasswordLogin()
	})
	login.loginTypePasswordButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab || event.Key() == tcell.KeyDown ||
			event.Key() == tcell.KeyUp || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight {
			login.app.SetFocus(login.loginTypeTokenButton)
			return nil
		}

		return event
	})

	loginChoiceView := tview.NewFlex().SetDirection(tview.FlexRow)
	loginChoiceView.AddItem(createCenteredComponent(login.loginTypeTokenButton, 66), 1, 0, true)
	loginChoiceView.AddItem(tview.NewBox(), 1, 0, false)
	loginChoiceView.AddItem(createCenteredComponent(tview.NewTextView().SetText("or").SetTextAlign(tview.AlignCenter), 66), 1, 0, false)
	loginChoiceView.AddItem(tview.NewBox(), 1, 0, false)
	loginChoiceView.AddItem(createCenteredComponent(login.loginTypePasswordButton, 66), 1, 0, false)

	passwordInputView := tview.NewFlex().SetDirection(tview.FlexRow)
	passwordInputView.AddItem(createCenteredComponent(login.usernameInput, 68), 3, 0, false)
	passwordInputView.AddItem(createCenteredComponent(login.passwordInput, 68), 3, 0, false)
	passwordInputView.AddItem(createCenteredComponent(login.tfaTokenInput, 68), 3, 0, false)

	login.AddItem(login.content, 0, 0, false)
	login.AddItem(createCenteredComponent(login.loginButton, 68), 1, 0, false)
	login.AddItem(tview.NewBox(), 0, 1, false)

	login.loginChoiceView = loginChoiceView
	login.passwordInputView = passwordInputView
	login.tokenInputView = login.tokenInput

	return login
}

func configureInputComponent(component *tview.InputField) {
	component.SetBorder(true)
	component.SetFieldWidth(66)
}

func createCenteredComponent(component tview.Primitive, width int) tview.Primitive {
	padding := tview.NewFlex().SetDirection(tview.FlexColumn)
	padding.AddItem(tview.NewBox(), 0, 1, false)
	padding.AddItem(component, width, 0, false)
	padding.AddItem(tview.NewBox(), 0, 1, false)

	return padding
}

func (login *Login) attemptLogin() {
	//Following two lines are little hack to prevent anything from being leftover
	//in the terminal buffer from the previous view. This is a bug in the tview
	//drawing logic.
	login.app.SetFocus(nil)
	login.ResizeItem(login.content, 0, 0)

	login.loginButton.SetVisible(false)
	login.messageText.SetText("Attempting to log in ...")

	switch login.loginType {
	case None:
		panic("Was in state loginType=None during login attempt.")
	case Token:
		session, loginError := discordgo.NewWithToken(userAgent, login.tokenInput.GetText())
		login.sessionChannel <- &loginAttempt{session, loginError}
	case Password:
		// Even if the login is supposed to be without two-factor-authentication, we
		// attempt parsing a 2fa code, since the underlying rest-call can also handle
		// non-2fa login calls.
		var mfaToken int64
		mfaTokenText := strings.ReplaceAll(login.tfaTokenInput.GetText(), " ", "")
		if mfaTokenText != "" {
			var parseError error
			mfaToken, parseError = strconv.ParseInt(mfaTokenText, 10, 32)
			if parseError != nil {
				login.sessionChannel <- &loginAttempt{nil, errors.New("[red]Two-Factor-Authentication Code incorrect.\n\n[red]Correct example: 564 231")}
			}
		}
		session, loginError := discordgo.NewWithPasswordAndMFA(userAgent, login.usernameInput.GetText(), login.passwordInput.GetText(), int(mfaToken))
		login.sessionChannel <- &loginAttempt{session, loginError}
	}

}

// RequestLogin prompts the user for a discord login. Login can either be done
// via e-mail and password or an authentication token.
func (login *Login) RequestLogin(additionalMessage string) (*discordgo.Session, error) {
	switch login.loginType {
	case None:
		login.showLoginTypeChoice()
	case Token:
		login.showTokenLogin()
	case Password:
		login.showPasswordLogin()
	}

	if additionalMessage != "" {
		login.messageText.SetText(additionalMessage + "\n\n" + login.messageText.GetText(false))
	}

	//Easier, since we don't have to call QueueUpdateDraw. Without this, we won't see any of the primitives.
	login.app.ForceDraw()

	loginAttempt := <-login.sessionChannel
	login.loginType = None
	return loginAttempt.session, loginAttempt.loginError
}

func (login *Login) showLoginTypeChoice() {
	login.showView(login.loginChoiceView, 5)
	login.loginButton.SetVisible(false)
	login.app.SetFocus(login.loginTypeTokenButton)
	login.messageText.SetText("Please decide for a login method.\n\nLogging in as a Bot will only work using a Authentication-Token.")
}

func (login *Login) showPasswordLogin() {
	login.showView(login.passwordInputView, 9)
	login.loginButton.SetVisible(true)
	login.app.SetFocus(login.usernameInput)
	login.messageText.SetText("Please input your E-Mail and password.\n\nIf you haven't enabled Two-Factor-Authentication on your account, just leave the field empty.")
}

func (login *Login) showTokenLogin() {
	login.showView(login.tokenInputView, 3)
	login.loginButton.SetVisible(true)
	login.app.SetFocus(login.tokenInput)
	login.messageText.SetText("Prepend 'Bot ' for bot tokens.\n\nFor information on how to retrieve your token, check:\nhttps://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token\n\nToken input is hidden by default, toggle with Ctrl + R.")
}

func (login *Login) showView(view tview.Primitive, size int) {
	login.content.RemoveAllItems()
	login.ResizeItem(login.content, 3, 0)
	login.content.AddItem(createCenteredComponent(view, 68), size, 0, false)
}
