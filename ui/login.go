package ui

import (
	"errors"
	"os"

	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/text"
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
	sessionChannel          chan *loginAttempt
	messageText             *tview.TextView

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
		loginTypePasswordButton: tview.NewButton("Login via E-Mail and password (Optionally Supports 2FA)"),
		messageText:             tview.NewTextView(),
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if shortcuts.ExitApplication.Equals(event) {
			app.Stop()
			//We call exit as we'd otherwise be waiting for a value
			//from the runNext channel.
			os.Exit(0)
			return nil
		}

		return event
	})

	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tviewutil.Escape(splashText + "\n\nConfig lies at: " + configDir))
	login.AddItem(splashScreen, 12, 0, false)
	login.AddItem(tviewutil.CreateCenteredComponent(login.messageText, 66), 0, 1, false)

	login.messageText.SetDynamicColors(true)

	configureInputComponent(login.tokenInput)
	login.tokenInput.SetTitle("Token")
	login.tokenInput.SetMaskCharacter(login.tokenInputMaskRune)
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
		switch event.Key() {
		case tcell.KeyEnter:
			login.attemptLogin()
			return nil
		case tcell.KeyTAB, tcell.KeyDown:
			login.app.SetFocus(login.passwordInput)
			return nil
		case tcell.KeyBacktab, tcell.KeyUp:
			login.app.SetFocus(login.tfaTokenInput)
			return nil
		case tcell.KeyCtrlV:
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.usernameInput.Insert(content)
		}
		return event
	})

	login.passwordInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			login.attemptLogin()
			return nil
		case tcell.KeyTAB, tcell.KeyDown:
			login.app.SetFocus(login.tfaTokenInput)
			return nil
		case tcell.KeyBacktab, tcell.KeyUp:
			login.app.SetFocus(login.usernameInput)
			return nil
		case tcell.KeyCtrlV:
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.passwordInput.Insert(content)
		}
		return event
	})

	login.tfaTokenInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			login.attemptLogin()
			return nil
		case tcell.KeyTAB, tcell.KeyDown:
			login.app.SetFocus(login.usernameInput)
			return nil
		case tcell.KeyBacktab, tcell.KeyUp:
			login.app.SetFocus(login.passwordInput)
			return nil
		case tcell.KeyCtrlV:
			content, clipError := clipboard.ReadAll()
			if clipError != nil {
				panic(clipError)
			}
			login.tfaTokenInput.Insert(content)
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
	loginChoiceView.AddItem(tviewutil.CreateCenteredComponent(login.loginTypeTokenButton, 66), 1, 0, true)
	loginChoiceView.AddItem(tview.NewBox(), 1, 0, false)
	loginChoiceView.AddItem(tviewutil.CreateCenteredComponent(tview.NewTextView().SetText("or").SetTextAlign(tview.AlignCenter), 66), 1, 0, false)
	loginChoiceView.AddItem(tview.NewBox(), 1, 0, false)
	loginChoiceView.AddItem(tviewutil.CreateCenteredComponent(login.loginTypePasswordButton, 66), 1, 0, false)

	passwordInputView := tview.NewFlex().SetDirection(tview.FlexRow)
	passwordInputView.AddItem(tviewutil.CreateCenteredComponent(login.usernameInput, 68), 3, 0, false)
	passwordInputView.AddItem(tviewutil.CreateCenteredComponent(login.passwordInput, 68), 3, 0, false)
	passwordInputView.AddItem(tviewutil.CreateCenteredComponent(login.tfaTokenInput, 68), 3, 0, false)

	login.AddItem(login.content, 0, 0, false)
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

func (login *Login) attemptLogin() {
	//Following two lines are little hack to prevent anything from being leftover
	//in the terminal buffer from the previous view. This is a bug in the tview
	//drawing logic.
	login.app.SetFocus(nil)
	login.ResizeItem(login.content, 0, 0)

	login.messageText.SetText("Attempting to log in ...")

	switch login.loginType {
	case None:
		panic("Was in state loginType=None during login attempt.")
	case Token:
		session, loginError := discordgo.NewWithToken(login.tokenInput.GetText())
		login.sessionChannel <- &loginAttempt{session, loginError}
	case Password:
		// Even if the login is supposed to be without two-factor-authentication, we
		// attempt parsing a 2fa code, since the underlying rest-call can also handle
		// non-2fa login calls.
		input := login.tfaTokenInput.GetText()
		var mfaTokenText string
		if input != "" {
			var parseError error
			mfaTokenText, parseError = text.ParseTFACode(input)
			if parseError != nil {
				login.sessionChannel <- &loginAttempt{nil, errors.New("[red]Two-Factor-Authentication Code incorrect.\n\n[red]Correct example: 564 231")}
			}
		}

		session, loginError := discordgo.NewWithPasswordAndMFA(login.usernameInput.GetText(), login.passwordInput.GetText(), mfaTokenText)
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
	login.app.SetFocus(login.loginTypeTokenButton)
	login.messageText.SetText("Please choose a login method.\n\nLogging in as a Bot will only work using a Authentication-Token.")
}

func (login *Login) showPasswordLogin() {
	login.showView(login.passwordInputView, 9)
	login.app.SetFocus(login.usernameInput)
	login.messageText.SetText("Please input your E-Mail and password.\n\nIf you haven't enabled Two-Factor-Authentication on your account, just leave the field empty.")
}

func (login *Login) showTokenLogin() {
	login.showView(login.tokenInputView, 3)
	login.app.SetFocus(login.tokenInput)
	login.messageText.SetText("Prepend 'Bot ' for bot tokens.\n\nFor information on how to retrieve your token, check:\nhttps://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token\n\nToken input is hidden by default, toggle with Ctrl + R.")
}

func (login *Login) showView(view tview.Primitive, size int) {
	login.content.RemoveAllItems()
	login.ResizeItem(login.content, size, 0)
	login.content.AddItem(tviewutil.CreateCenteredComponent(view, 68), size, 0, false)
}
