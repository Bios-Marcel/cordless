package app

import (
	"fmt"
	"github.com/Bios-Marcel/cordless/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/version"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"log"
	"os"
)

const (
	userAgent = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:70.0) Gecko/20100101 Firefox/70.0"
)

var suspended = false

// Run launches the whole application and might abort in case it encounters an
//error.
func Run() {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	themeLoadingError := config.LoadTheme()
	if themeLoadingError == nil {
		tview.Styles = *config.GetTheme().Theme
	}

	app := tview.NewApplication()
	loginScreen := ui.NewLogin(app, configDir)
	app.SetRoot(loginScreen, true)
	runNext := make(chan bool, 1)

	configuration, configLoadError := config.LoadConfig()

	if configLoadError != nil {
		log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
	}

	updateAvailableChannel := make(chan bool, 1)
	if configuration.ShowUpdateNotifications {
		go func() {
			updateAvailableChannel <- version.IsLocalOutdated(configuration.DontShowUpdateNotificationFor)
		}()
	} else {
		updateAvailableChannel <- false
	}

	app.MouseEnabled = configuration.MouseEnabled

	go func() {
		shortcutsLoadError := shortcuts.Load()
		if shortcutsLoadError != nil {
			panic(shortcutsLoadError)
		}

		discord, readyEvent := attemptLogin(loginScreen, "", configuration)

		config.Current.Token = discord.Token

		persistError := config.PersistConfig()
		if persistError != nil {
			app.Stop()
			log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
		}

		discord.State.MaxMessageCount = 100

		readstate.Load(discord.State)

		isUpdateAvailable := <-updateAvailableChannel
		close(updateAvailableChannel)
		if isUpdateAvailable {
			waitForUpdateDialogChannel := make(chan bool, 1)

			dialog := tview.NewModal()
			dialog.SetText(fmt.Sprintf("Version %s of cordless is available!\nYou are currently running version %s.\n\nUpdates have to be installed manually or via your package manager.", version.GetLatestRemoteVersion(), version.Version))
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
			window.RegisterCommand(statusSetCmd)
			window.RegisterCommand(statusGetCmd)
			window.RegisterCommand(commandimpls.NewStatusCommand(statusGetCmd, statusSetCmd))
			window.RegisterCommand(commandimpls.NewFileSendCommand(discord, window))
			window.RegisterCommand(commandimpls.NewAccount(runNext, window))
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
			window.RegisterCommand(serverJoinCmd)
			window.RegisterCommand(serverLeaveCmd)
			window.RegisterCommand(commandimpls.NewServerCommand(serverJoinCmd, serverLeaveCmd))
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
		})
	}()

	runError := app.Run()
	if runError != nil {
		log.Fatalf("Error launching View (%s).\n", runError.Error())
	}

	run := <-runNext
	if run {
		Run()
	}
}

func attemptLogin(loginScreen *ui.Login, loginMessage string, configuration *config.Config) (*discordgo.Session, *discordgo.Ready) {
	var (
		session      *discordgo.Session
		readyEvent   *discordgo.Ready
		discordError error
	)

	if configuration.Token == "" {
		session, discordError = loginScreen.RequestLogin(loginMessage)
	} else {
		session, discordError = discordgo.NewWithToken(userAgent, configuration.Token)
	}

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(loginScreen, fmt.Sprintf("Error during last login attempt:\n\n[red]%s", discordError), configuration)
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
