package commands

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
)

func FixLayout(writer io.Writer, window *ui.Window, parameters []string) {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			fmt.Fprintln(writer, "The given input was incorrect, there has to be only one parameter, which can only be of the value 'true' or 'false'")
			return
		}

		config.GetConfig().UseFixedLayout = choice
		window.RefreshLayout()

		persistError := config.PersistConfig()
		if persistError != nil {
			fmt.Fprintf(writer, "Error saving configuration: %s\n", persistError.Error())
			return
		}

		if choice {
			fmt.Fprintln(writer, "FixLayout has been enabled")
		} else {
			fmt.Fprintln(writer, "FixLayout has been disabled")
		}
	} else if len(parameters) == 2 {
		size, parseError := strconv.ParseInt(parameters[1], 10, 64)
		if parseError != nil {
			fmt.Fprintln(writer, "The given input was invalid, it has to be an integral number greater than -1")
			return
		}

		if size < 0 {
			fmt.Fprintln(writer, "The given input was out of bounds, it has to be bigger than -1")
			return
		}

		// TODO Check for upper limit?

		var successOutput string
		subCommand := parameters[0]
		if subCommand == "left" {
			config.GetConfig().FixedSizeLeft = int(size)
			successOutput = fmt.Sprintf("The left side of the layout was set to %d", int(size))
		} else if subCommand == "right" {
			config.GetConfig().FixedSizeRight = int(size)
			successOutput = fmt.Sprintf("The right side of the layout was set to %d", int(size))
		} else {
			fmt.Fprintf(writer, "The subcommand '%s' does not exist\n", subCommand)
			return
		}

		window.RefreshLayout()

		persistError := config.PersistConfig()
		if persistError != nil {
			fmt.Fprintf(writer, "Error saving configuration: %s\n", persistError.Error())
			return
		}

		if successOutput != "" {
			fmt.Fprintln(writer, successOutput)
		}
	}
	// TODO Else ... Print help
}
