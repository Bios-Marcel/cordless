package internal

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/Bios-Marcel/cordless/internal/config"
)

//Run launches the whole application and might abort in case it encounters an
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

	if configuration.Token == "" {
		log.Println("The Discord token could not be found, please input your token.")
		configuration.Token = askForToken()
		persistError := config.PersistConfig()
		if persistError != nil {
			log.Fatalf("Error persisting configuration (%s).\n", persistError.Error())
		}
	}
}

func askForToken() string {
	reader := bufio.NewReader(os.Stdin)
	token, inputError := reader.ReadString('\n')

	if inputError != nil {
		log.Fatalf("Error reading your token (%s).\n", inputError.Error())
	}

	if token == "" {
		log.Println("An empty token is not valid, please try again.")
		return askForToken()
	}

	token = strings.TrimSuffix(token, "\n")

	return token
}
