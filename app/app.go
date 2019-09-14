package app

import (
	"fmt"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"log"
	"os"

	"github.com/Bios-Marcel/cordless/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

const (
	userSession = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:66.0) Gecko/20100101 Firefox/66.0"
	defaultLoginMessage = "Input your token. Prepend 'Bot ' for bot tokens.\n\nFor information on how to retrieve your token, check:\nhttps://github.com/Bios-Marcel/cordless/wiki/Retrieving-your-token"
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
	loginScreen := ui.NewLogin( app, configDir)
	app.SetRoot(loginScreen, true)
	runNext := make(chan bool, 1)

	configuration, configLoadError := config.LoadConfig()

	if configLoadError != nil {
		log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
	}

	app.MouseEnabled = configuration.MouseEnabled

	go func() {
		shortcutsLoadError := shortcuts.Load()
		if shortcutsLoadError != nil {
			panic(shortcutsLoadError)
		}

		discord := attemptLogin(loginScreen, defaultLoginMessage, app, configuration)

		config.GetConfig().Token = discord.Token

		persistError := config.PersistConfig()
		if persistError != nil {
			app.Stop()
			log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
		}

		readyChan := make(chan *discordgo.Ready)
		discord.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.Ready) {
			readyChan <- event
		})

		discordError := discord.Open()
		if discordError != nil {
			app.Stop()
			log.Fatalln("Error establishing web socket connection", discordError)
		}

		readyEvent := <-readyChan

		readstate.Load(discord.State)

		app.QueueUpdateDraw(func() {
			window, createError := ui.NewWindow(runNext, app, discord, readyEvent)

			if createError != nil {
				app.Stop()
				//Otherwise the logger output can't be seen, since we are stopping the TUI either way.
				log.SetOutput(os.Stdout)
				log.Fatalf("Error constructing window (%s).\n", createError.Error())
			}

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

func attemptLogin(loginScreen *ui.Login, loginMessage string, app *tview.Application, configuration *config.Config) *discordgo.Session {
	var (
		session      *discordgo.Session
		discordError error
	)

	if configuration.Token == "" {
		token := loginScreen.RequestToken(loginMessage)
		session, discordError = discordgo.NewWithToken(
			userSession,
			token)
	} else {
		session, discordError = discordgo.NewWithToken(
			userSession,
			configuration.Token)
	}

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(loginScreen, fmt.Sprintf("Error during last login attempt:\n\n[red]%s\n\n%s", discordError, defaultLoginMessage), app, configuration)
	}

	//When logging in via token, the token isn't ever validated, therefore we do a test lookup here.
	_, discordError = session.UserGuilds(0, "", "")
	if discordError == discordgo.ErrUnauthorized {
		configuration.Token = ""
		return attemptLogin(loginScreen, fmt.Sprintf("Error during last login attempt:\n\n[red]%s\n\n%s", discordError, defaultLoginMessage), app, configuration)
	}

	return session
}
