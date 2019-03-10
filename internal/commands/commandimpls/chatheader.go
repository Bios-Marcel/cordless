package commandimpls

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/ui"
)

const chatHeaderDocumentation = `[orange]# chatheader[white]

The chatheader command [green]shows[white] or [red]hides[white] the chatheader
and saves the new state to the user configuration.

The chatheader is the textbox above the message container. It contains the
channel name and optionally a topic if the channel has one.

Usage: chatheader <[green]true[white]/[red]false[white]>`

// ChatHeader is the command that allows toggling the chatheader.
type ChatHeader struct {
	window *ui.Window
}

// NewChatHeaderCommand creates a new ready-to-use ChatHeader command.
func NewChatHeaderCommand(window *ui.Window) *ChatHeader {
	return &ChatHeader{
		window: window,
	}
}

// Execute runs the command piping its output into the supplied writer.
func (chatHeader *ChatHeader) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			chatHeader.PrintHelp(writer)
			return
		}

		config.GetConfig().ShowChatHeader = choice
		chatHeader.window.RefreshChannelTitle()

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
		chatHeader.PrintHelp(writer)
	}
}

// Name represents this commands indentifier.
func (chatHeader *ChatHeader) Name() string {
	return "chatheader"
}

// PrintHelp prints a static help page for this command
func (chatHeader *ChatHeader) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, chatHeaderDocumentation)
}
