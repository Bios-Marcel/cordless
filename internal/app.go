package internal

import (
	"log"

	"github.com/Bios-Marcel/cordless/internal/config"
)

func Run() {
	configDir, configErr := config.GetConfigDirectory()

	if configErr != nil {
		log.Fatalf("Unable to determine configuration directory (%s)\n", configErr.Error())
	}

	log.Printf("Configuration lies at: %s\n", configDir)
}
