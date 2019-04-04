package app

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Bios-Marcel/cordless/shortcuts"

	"github.com/princebot/getpass"

	"github.com/Bios-Marcel/cordless/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
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
     ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚══════╝╚══════╝╚══════╝              
`
	defaultLoginMessage = "Please choose wether to login via authentication token (1) or email and password (2)."
)

// Run launches the whole application and might abort in case it encounters an
//error.
func Run() {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	defineColorTheme()

	app := tview.NewApplication()
	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tview.Escape(splashText + "\n\nConfig lies at: " + configDir))
	app.SetRoot(splashScreen, true)
	runNext := make(chan bool, 1)

	go func() {
		configuration, configLoadError := config.LoadConfig()

		if configLoadError != nil {
			app.Stop()
			log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
		}

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

		readyChan := make(chan struct{})
		discord.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.Ready) {
			readyChan <- struct{}{}
		})

		discordError := discord.Open()
		if discordError != nil {
			app.Stop()
			log.Fatalln("Error establishing web socket connection", discordError)
		}

		<-readyChan

		app.QueueUpdateDraw(func() {
			window, createError := ui.NewWindow(runNext, app, discord)

			if createError != nil {
				app.Stop()
				log.Fatalf("Error constructing window (%s).\n", createError.Error())
			}

			window.RegisterCommand(commandimpls.NewStatusCommand(discord))
			window.RegisterCommand(commandimpls.NewAccount(runNext, window))
			window.RegisterCommand(commandimpls.NewHelpCommand(window))
			window.RegisterCommand(commandimpls.NewManualCommand())
			window.RegisterCommand(commandimpls.NewFixLayoutCommand(window))
			window.RegisterCommand(commandimpls.NewChatHeaderCommand(window))
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
		app.Suspend(func() {
			session, discordError = login(loginMessage)
		})
	} else {
		session, discordError = discordgo.New(configuration.Token)
	}

	if discordError != nil {
		configuration.Token = ""
		return attemptLogin(fmt.Sprintf("Your password or token was incorrect, please try again.\n%s", defaultLoginMessage), app, configuration)
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

	return discordgo.New(name, password)
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

	return discordgo.New(token)
}
