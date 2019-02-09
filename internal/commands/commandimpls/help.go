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
`

// Help is the command that offers help for specific commands in the app.
type Help struct {
	window *ui.Window
}

// NewHelpCommand constructs a new ready-to-use Help command.
func NewHelpCommand(window *ui.Window) *Help {
	return &Help{
		window: window,
	}
}

// Execute runs the command piping its output into the supplied writer.
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
		} else {
			fmt.Fprintf(writer, "[red]The command '%s' doesn't exist.[white]\n", parameters[0])
		}
	} else {
		help.PrintHelp(writer)
	}
}

// Name represents this commands indentifier.
func (help *Help) Name() string {
	return "help"
}

// PrintHelp prints a static help page for this command
func (help *Help) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, helpDocumentation)
}
