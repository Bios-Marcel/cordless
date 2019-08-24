package app

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/shortcuts"

	"github.com/princebot/getpass"

	"github.com/Bios-Marcel/cordless/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tcellutil"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

const (
	splashText = `
 ██████╗ ██████╗ ██████╗ ██████╗ ██╗     ███████╗███████╗███████╗              
██╔════╝██╔═══██╗██╔══██╗██╔══██╗██║     ██╔════╝██╔════╝██╔════╝     /   \    
██║     ██║   ██║██████╔╝██║  ██║██║     █████╗  ███████╗███████╗ ████-   -████
██║     ██║   ██║██╔══██╗██║  ██║██║     ██╔══╝  ╚════██║╚════██║     \   /    
╚██████╗╚██████╔╝██║  ██║██████╔╝███████╗███████╗███████║███████║              
 ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚══════╝╚══════╝              `

	defaultLoginMessage = "Please choose whether to login via authentication token (1) or email and password (2)."
	userSession         = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:66.0) Gecko/20100101 Firefox/66.0"
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
	screen, err := tcellutil.NewScreen()
	if err != nil {
		log.Fatalf("Unable to create TUI screen (%s)\n", err.Error())
	}
	app.SetScreen(screen)
	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tview.Escape(splashText + "\n\nConfig lies at: " + configDir))
	app.SetRoot(splashScreen, true)
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

		discord := attemptLogin(defaultLoginMessage, app, configuration)

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
				log.Fatalf("Error constructing window (%s).\n", createError.Error())
			}

			window.RegisterCommand(commandimpls.NewStatusCommand(discord), "status")
			window.RegisterCommand(commandimpls.NewFileSendCommand(discord, window), "file-send")
			window.RegisterCommand(commandimpls.NewAccount(runNext, window), "account", "profile")
			window.RegisterCommand(commandimpls.NewHelpCommand(window), "help")
			window.RegisterCommand(commandimpls.NewManualCommand(), "manual")
			window.RegisterCommand(commandimpls.NewFixLayoutCommand(window), "fixlayout", "fix-layout")
			window.RegisterCommand(commandimpls.NewChatHeaderCommand(window), "chatheader", "chat-header")
			window.RegisterCommand(commandimpls.NewFriendsCommand(discord), "friends")
			userSetCmd := commandimpls.NewUserSetCommand(window, discord)
			userGetCmd := commandimpls.NewUserGetCommand(window, discord)
			window.RegisterCommand(userSetCmd, "user-set", "user-update")
			window.RegisterCommand(userGetCmd, "user-get")
			window.RegisterCommand(commandimpls.NewUserCommand(userSetCmd, userGetCmd), "user")
			window.RegisterCommand(commandimpls.NewServerCommand(discord, window), "server")
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

func attemptLogin(loginMessage string, app *tview.Application, configuration *config.Config) *discordgo.Session {
	var (
		session      *discordgo.Session
		discordError error
	)

	if configuration.Token == "" {
		if suspended {
			session, discordError = login(loginMessage)
		} else {
			suspended = true
			app.Suspend(func() {
				session, discordError = login(loginMessage)
			})
			suspended = false
		}
	} else {
		session, discordError = discordgo.NewWithToken(
			userSession,
			configuration.Token)
	}

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(fmt.Sprintf("Error during login attempt (%s).\n%s", discordError, defaultLoginMessage), app, configuration)
	}

	//When logging in via token, the token isn't ever validated, therefore we do a test lookup here.
	_, discordError = session.UserGuilds(0, "", "")
	if discordError == discordgo.ErrUnauthorized {
		configuration.Token = ""
		return attemptLogin(fmt.Sprintf("Your password or token was incorrect, please try again.\n%s", defaultLoginMessage), app, configuration)
	}

	return session
}

func login(loginMessage string) (*discordgo.Session, error) {
	log.Println(loginMessage)
	var choice int

	var err error
	if runtime.GOOS == "windows" {
		_, err = fmt.Scanf("%d\n", &choice)
	} else {
		_, err = fmt.Scanf("%d", &choice)
	}

	if err != nil {
		log.Println("Invalid input, please try again.")
		return login(loginMessage)
	}

	if choice == 1 {
		return askForToken()
	} else if choice == 2 {
		return askForEmailAndPassword()
	} else {
		log.Println()
		return login(fmt.Sprintf("Invalid choice, please try again.\n%s", defaultLoginMessage))
	}
}

func askForEmailAndPassword() (*discordgo.Session, error) {
	log.Println("Please input your email.")
	reader := bufio.NewReader(os.Stdin)

	nameAsBytes, _, inputError := reader.ReadLine()
	if inputError != nil {
		log.Fatalf("Error reading your email (%s).\n", inputError.Error())
	}
	name := string(nameAsBytes[:])

	passwordAsBytes, inputError := getpass.Get("Please input your password.\n")
	if inputError != nil {
		log.Fatalf("Error reading your email (%s).\n", inputError.Error())
	}
	password := string(passwordAsBytes[:])

	return discordgo.NewWithPassword(userSession, name, password)
}

func askForToken() (*discordgo.Session, error) {
	log.Println("Please input your token.")
	reader := bufio.NewReader(os.Stdin)
	tokenAsBytes, _, inputError := reader.ReadLine()

	if inputError != nil {
		log.Fatalf("Error reading your token (%s).\n", inputError.Error())
	}

	token := string(tokenAsBytes[:])

	if token == "" {
		log.Println("An empty token is not valid, please try again.")
		return askForToken()
	}

	return discordgo.NewWithToken(userSession, token)
}
