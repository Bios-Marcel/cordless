package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/internal/ui"
)

const helpDocumentation = `[orange]# help[white]

For a list of commands type
    [-]help commands

[white]For the help page of a specific command type
	[-]help <[blue]command[-]>
[white]where [blue]command[white] is the name of the command.

For a manual type
    [-]help manual
`

type Help struct {
	window *ui.Window
}

func NewHelpCommand(window *ui.Window) *Help {
	return &Help{
		window: window,
	}
}

func (help *Help) Name() string {
	return "help"
}

func (help *Help) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 1 {
		command, contains := help.window.GetRegisteredCommands()[parameters[0]]
		if contains {
			command.PrintHelp(writer)
		} else if parameters[0] == "commands" {
			fmt.Fprintln(writer, "# Available commands")
			for name := range help.window.GetRegisteredCommands() {
				fmt.Fprintln(writer, "    * "+name)
			}
		} else if parameters[0] == "manual" {
			fmt.Fprintln(writer, "Not written yet! Refer to the wiki at:\nhttps://github.com/Bios-Marcel/cordless/wiki")
		} else {
			fmt.Fprintf(writer, "[red]The command '%s' doesn't exist.[white]\n", parameters[0])
		}
	} else {
		help.PrintHelp(writer)
	}
}

func (help *Help) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, helpDocumentation)
}
