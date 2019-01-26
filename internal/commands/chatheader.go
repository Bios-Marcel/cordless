package commands

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
)

func ChatHeader(writer io.Writer, window *ui.Window, parameters []string) {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			printHelp(writer)
			return
		}

		config.GetConfig().ShowChatHeader = choice
		window.RefreshLayout()

		persistError := config.PersistConfig()
		if persistError != nil {
			fmt.Fprintf(writer, "Error saving configuration: %s\n", persistError.Error())
			return
		}

		if choice {
			fmt.Fprintln(writer, "Chatheader has been enabled")
		} else {
			fmt.Fprintln(writer, "Chatheader has been disabled")
		}
	} else {
		printHelp(writer)
	}
}

func printHelp(writer io.Writer) {
	fmt.Fprintf(writer, "Usage: chatheader <true/false>\n")
}
