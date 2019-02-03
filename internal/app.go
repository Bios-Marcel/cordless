package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/princebot/getpass"

	"github.com/Bios-Marcel/cordless/internal/commands"
	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
	"github.com/Bios-Marcel/discordgo"
)

// Run launches the whole application and might abort in case it encounters an
//error.
func Run() {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	log.Printf("Configuration lies at: %s\n", configDir)

	configuration, configLoadError := config.LoadConfig()

	if configLoadError != nil {
		log.Fatalf("Error loading configuration file (%s).\n", configLoadError.Error())
	}

	var (
		discord      *discordgo.Session
		discordError error
	)
	if configuration.Token == "" {
		discord, discordError = login()
		config.GetConfig().Token = discord.Token
	} else {
		discord, discordError = discordgo.New(configuration.Token)
	}

	if discordError != nil {
		//TODO Handle better
		log.Fatalln("Error logging into Discord", discordError)
	}

	persistError := config.PersistConfig()
	if persistError != nil {
		log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
	}

	discordError = discord.Open()
	if discordError != nil {
		//TODO Handle better
		log.Fatalln("Error establishing web socket connection", discordError)
	}

	window, createError := ui.NewWindow(discord)

	if createError != nil {
		log.Fatalf("Error constructing window (%s).\n", createError.Error())
	}

	window.RegisterCommand("fixlayout", commands.FixLayout)
	window.RegisterCommand("chatheader", commands.ChatHeader)

	runError := window.Run()
	if runError != nil {
		log.Fatalf("Error launching View (%s).\n", runError.Error())
	}

}

func login() (*discordgo.Session, error) {
	log.Println("Please choose wether to login via username and password (1) or authentication token (2).")
	var choice int

	_, err := fmt.Scanf("%d", &choice)

	if err != nil {
		log.Println("Invalid input, please try again.")
		return login()
	}

	if choice == 1 {
		return askForUsernameAndPassword()
	} else if choice == 2 {
		return askForToken()
	} else {
		log.Println("Invalid choice, please try again.")
		return login()
	}
}

func askForUsernameAndPassword() (*discordgo.Session, error) {
	log.Println("Please input your username.")
	reader := bufio.NewReader(os.Stdin)

	nameAsBytes, _, inputError := reader.ReadLine()
	if inputError != nil {
		log.Fatalf("Error reading your username (%s).\n", inputError.Error())
	}
	name := string(nameAsBytes[:])

	passwordAsBytes, inputError := getpass.Get("Please input your password.\n")
	if inputError != nil {
		log.Fatalf("Error reading your username (%s).\n", inputError.Error())
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
