package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/princebot/getpass"

	"github.com/Bios-Marcel/cordless/internal/commands/commandimpls"
	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
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
)

// Run launches the whole application and might abort in case it encounters an
//error.
func Run() {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	app := tview.NewApplication()
	splashScreen := tview.NewTextView()
	splashScreen.SetTextAlign(tview.AlignCenter)
	splashScreen.SetText(tview.Escape(splashText + "\n\nConfig lies at: " + configDir))
	app.SetRoot(splashScreen, true)

	go func() {
		configuration, configLoadError := config.LoadConfig()

		if configLoadError != nil {
			app.Stop()
			log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
		}

		var (
			discord      *discordgo.Session
			discordError error
		)

		if configuration.Token == "" {
			app.Suspend(func() {
				discord, discordError = login()
			})
			config.GetConfig().Token = discord.Token
		} else {
			discord, discordError = discordgo.New(configuration.Token)
		}

		if discordError != nil {
			app.Stop()
			log.Fatalln("Error logging into Discord", discordError)
		}

		persistError := config.PersistConfig()
		if persistError != nil {
			app.Stop()
			log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
		}

		readyChan := make(chan struct{})
		discord.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.Ready) {
			readyChan <- struct{}{}
		})

		discordError = discord.Open()
		if discordError != nil {
			app.Stop()
			log.Fatalln("Error establishing web socket connection", discordError)
		}

		<-readyChan

		app.QueueUpdateDraw(func() {
			window, createError := ui.NewWindow(app, discord)

			if createError != nil {
				app.Stop()
				log.Fatalf("Error constructing window (%s).\n", createError.Error())
			}

			window.RegisterCommand(commandimpls.NewStatusCommand(discord))
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
}

func login() (*discordgo.Session, error) {
	log.Println("Please choose wether to login via authentication token (1) or email and password (2).")
	var choice int

	var err error
	if runtime.GOOS == "windows" {
		_, err = fmt.Scanf("%d\n", &choice)
	} else {
		_, err = fmt.Scanf("%d", &choice)
	}

	if err != nil {
		log.Println("Invalid input, please try again.")
		return login()
	}

	if choice == 1 {
		return askForToken()
	} else if choice == 2 {
		return askForEmailAndPassword()
	} else {
		log.Println("Invalid choice, please try again.")
		return login()
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
