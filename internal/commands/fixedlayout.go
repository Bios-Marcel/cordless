package commands

import (
	"errors"
	"strconv"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
)

func FixedLayout(window *ui.Window, parameters []string) error {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			return errors.New("the given input was incorrect, there has to be only one parameter, which can only be of the value 'true' or 'false'")
		}

		config.GetConfig().UseFixedLayout = choice
		window.RefreshLayout()

		persistError := config.PersistConfig()
		if persistError != nil {
			return persistError
		}

		return nil
	}

	if len(parameters) == 2 {
		size, parseError := strconv.ParseInt(parameters[1], 10, 64)
		if parseError != nil {
			return errors.New("the given input was invalid, it has to be an integral number greater than -1")
		}

		if size < 0 {
			return errors.New("the given input was out of bounds, it has to be bigger than -1")
		}

		//TODO Check for upper limit?

		subCommand := parameters[0]
		if subCommand == "left" {
			config.GetConfig().FixedSizeLeft = int(size)
		} else if subCommand == "right" {
			config.GetConfig().FixedSizeRight = int(size)
		} else {
			return errors.New("the subcommand" + subCommand + " does not exist")
		}

		window.RefreshLayout()

		persistError := config.PersistConfig()
		if persistError != nil {
			return persistError
		}
	}

	//TODO Print help

	return nil
}
