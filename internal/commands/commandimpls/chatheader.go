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

type ChatHeader struct {
	window *ui.Window
}

func NewChatHeaderCommand(window *ui.Window) *ChatHeader {
	return &ChatHeader{
		window: window,
	}
}

func (chatHeader *ChatHeader) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			chatHeader.PrintHelp(writer)
			return
		}

		config.GetConfig().ShowChatHeader = choice
		chatHeader.window.RefreshLayout()

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

func (_ *ChatHeader) Name() string {
	return "chatheader"
}

func (_ *ChatHeader) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, chatHeaderDocumentation)
}
