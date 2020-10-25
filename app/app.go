package app

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/windowman"
	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/login"
	"github.com/Bios-Marcel/cordless/version"
)

// RunWithAccount launches the whole application and might
// abort in case it encounters an error. The login will attempt
// using the account specified, unless the argument is empty.
// If the account can't be found, the login page will be shown.
func RunWithAccount(account string) {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	windowManager := windowman.GetWindowManager()
	loginWindow := login.NewLoginWindow(configDir)

	app := tview.NewApplication()
	loginScreen := loginWindow.LoginWindowComponent
	runNext := make(chan bool, 1)

	configuration, configLoadError := config.LoadConfig()
	if configLoadError != nil {
		log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
	}

	if strings.TrimSpace(account) != "" {
		configuration.Token = configuration.GetAccountToken(account)
	}

	updateAvailableChannel := version.CheckForUpdate(configuration.DontShowUpdateNotificationFor)
	app.MouseEnabled = configuration.MouseEnabled

	go func() {
		log.Println("entered goroutine")
		shortcutsLoadError := shortcuts.Load()
		if shortcutsLoadError != nil {
			panic(shortcutsLoadError)
		}
		log.Println("attempting login")
		discord, readyEvent := attemptLogin(loginScreen, "", configuration)
		log.Println("attempted login")

		config.Current.Token = discord.Token

		persistError := config.PersistConfig()
		if persistError != nil {
			app.Stop()
			log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
		}

		discord.State.MaxMessageCount = 100

		readstate.Load(discord.State)

		if isUpdateAvailable := <-updateAvailableChannel; isUpdateAvailable {
			waitForUpdateDialogChannel := make(chan bool, 1)

			dialog := tview.NewModal()
			dialog.SetText(fmt.Sprintf("Version %s of cordless is available!\nYou are currently running version %s.\n\nUpdates have to be installed manually or via your package manager.\n\nThe snap package manager isn't supported by cordless anymore!", version.GetLatestRemoteVersion(), version.Version))
			buttonOk := "Thanks for the info"
			buttonDontRemindAgainForThisVersion := fmt.Sprintf("Skip reminders for %s", version.GetLatestRemoteVersion())
			buttonNeverRemindMeAgain := "Never remind me again"
			dialog.AddButtons([]string{buttonOk, buttonDontRemindAgainForThisVersion, buttonNeverRemindMeAgain})
			dialog.SetDoneFunc(func(index int, label string) {
				if label == buttonDontRemindAgainForThisVersion {
					configuration.DontShowUpdateNotificationFor = version.GetLatestRemoteVersion()
					config.PersistConfig()
				} else if label == buttonNeverRemindMeAgain {
					configuration.ShowUpdateNotifications = false
					config.PersistConfig()
				}

				waitForUpdateDialogChannel <- true
			})

			app.QueueUpdateDraw(func() {
				app.SetRoot(dialog, true)
			})

			<-waitForUpdateDialogChannel
			close(waitForUpdateDialogChannel)
		}

		app.QueueUpdateDraw(func() {
			window, createError := ui.NewWindow(runNext, app, discord, readyEvent)

			if createError != nil {
				app.Stop()
				//Otherwise the logger output can't be seen, since we are stopping the TUI either way.
				log.SetOutput(os.Stdout)
				log.Fatalf("Error constructing window (%s).\n", createError.Error())
			}

			window.RegisterCommand(commandimpls.NewVersionCommand())
			statusGetCmd := commandimpls.NewStatusGetCommand(discord)
			statusSetCmd := commandimpls.NewStatusSetCommand(discord)
			statusSetCustomCmd := commandimpls.NewStatusSetCustomCommand(discord)
			window.RegisterCommand(statusSetCmd)
			window.RegisterCommand(statusGetCmd)
			window.RegisterCommand(statusSetCustomCmd)
			window.RegisterCommand(commandimpls.NewStatusCommand(statusGetCmd, statusSetCmd, statusSetCustomCmd))
			window.RegisterCommand(commandimpls.NewFileSendCommand(discord, window))
			accountLogout := commandimpls.NewAccountLogout(runNext, window)
			window.RegisterCommand(accountLogout)
			window.RegisterCommand(commandimpls.NewAccount(accountLogout, window))
			window.RegisterCommand(commandimpls.NewManualCommand(window))
			window.RegisterCommand(commandimpls.NewFixLayoutCommand(window))
			window.RegisterCommand(commandimpls.NewFriendsCommand(discord))
			userSetCmd := commandimpls.NewUserSetCommand(window, discord)
			userGetCmd := commandimpls.NewUserGetCommand(window, discord)
			window.RegisterCommand(userSetCmd)
			window.RegisterCommand(userGetCmd)
			window.RegisterCommand(commandimpls.NewUserCommand(userSetCmd, userGetCmd))
			serverJoinCmd := commandimpls.NewServerJoinCommand(window, discord)
			serverLeaveCmd := commandimpls.NewServerLeaveCommand(window, discord)
			serverCreateCmd := commandimpls.NewServerCreateCommand(discord)
			window.RegisterCommand(serverJoinCmd)
			window.RegisterCommand(serverLeaveCmd)
			window.RegisterCommand(serverCreateCmd)
			window.RegisterCommand(commandimpls.NewServerCommand(serverJoinCmd, serverLeaveCmd, serverCreateCmd))
			window.RegisterCommand(commandimpls.NewNickSetCmd(discord, window))
			tfaEnableCmd := commandimpls.NewTFAEnableCommand(window, discord)
			tfaDisableCmd := commandimpls.NewTFADisableCommand(discord)
			tfaBackupGetCmd := commandimpls.NewTFABackupGetCmd(discord, window)
			tfaBackupResetCmd := commandimpls.NewTFABackupResetCmd(discord, window)
			window.RegisterCommand(commandimpls.NewTFACommand(tfaEnableCmd, tfaDisableCmd, tfaBackupGetCmd, tfaBackupResetCmd))
			window.RegisterCommand(tfaEnableCmd)
			window.RegisterCommand(tfaDisableCmd)
			window.RegisterCommand(tfaBackupGetCmd)
			window.RegisterCommand(tfaBackupResetCmd)
			window.RegisterCommand(commandimpls.NewDMOpenCmd(discord, window))
		})
	}()

	log.Println("running")
	runError := windowManager.Run(&loginWindow)
	if runError != nil {
		log.Fatalf("Error launching View (%v).\n", runError)
	}

	run := <-runNext
	if run {
		Run()
	}
}

// Run launches the whole application and might abort in case
// it encounters an error.
func Run() {
	RunWithAccount("")
}

func attemptLogin(loginScreen *login.Login, loginMessage string, configuration *config.Config) (*discordgo.Session, *discordgo.Ready) {
	var (
		session      *discordgo.Session
		readyEvent   *discordgo.Ready
		discordError error
	)

	if configuration.Token == "" {
		session, discordError = loginScreen.RequestLogin(loginMessage)
	} else {
		session, discordError = discordgo.NewWithToken(configuration.Token)
	}

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(loginScreen, fmt.Sprintf("Error during last login attempt:\n\n[red]%s", discordError), configuration)
	}

	if session == nil {
		configuration.Token = ""
		return attemptLogin(loginScreen, "Error during last login attempt:\n\n[red]Received session is nil", configuration)
	}

	readyChan := make(chan *discordgo.Ready, 1)
	session.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.Ready) {
		readyChan <- event
	})

	discordError = session.Open()

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(loginScreen, fmt.Sprintf("Error during last login attempt:\n\n[red]%s", discordError), configuration)
	}

	readyEvent = <-readyChan
	close(readyChan)

	return session, readyEvent
}
