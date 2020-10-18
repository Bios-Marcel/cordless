package commandimpls

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

var fixLayoutDocumentation = `[orange]# fixlayout[white]

The fixlayout command allows adjusting the layout of the application to a certain degree. By default most components take a flexible amount of space. By activating the fixlayout, those components will instead use a fixed amount of space.

You can [green]enable[white] or [` + tviewutil.ColorToHex(config.GetTheme().ErrorColor) + `]disable[white] the fixlayout by using this:
    [-]fixlayout <[green]true[-]/[` + tviewutil.ColorToHex(config.GetTheme().ErrorColor) + `]false[-]>

[white]In order to specify the width of a component, use this:
    [-]fixlayout <left/right> <[blue]N[-]>
[white]where [blue]N[white] is the width of the component.
`

// FixLayout is the command that allows the user to change the applications
// layout.
type FixLayout struct {
	window *ui.Window
}

// NewFixLayoutCommand creates a ready-to-use FixLayout command.
func NewFixLayoutCommand(window *ui.Window) *FixLayout {
	return &FixLayout{
		window: window,
	}
}

// Execute runs the command piping its output into the supplied writer.
func (fixLayout *FixLayout) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 1 {
		choice, parseError := strconv.ParseBool(parameters[0])
		if parseError != nil {
			fmt.Fprintln(writer, "The given input was incorrect, there has to be only one parameter, which can only be of the value 'true' or 'false'")
			return
		}

		config.Current.UseFixedLayout = choice
		fixLayout.window.ApplyFixedLayoutSettings()

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
			config.Current.FixedSizeLeft = int(size)
			successOutput = fmt.Sprintf("The left side of the layout was set to %d", int(size))
		} else if subCommand == "right" {
			config.Current.FixedSizeRight = int(size)
			successOutput = fmt.Sprintf("The right side of the layout was set to %d", int(size))
		} else {
			fmt.Fprintf(writer, "The subcommand '%s' does not exist\n", subCommand)
			return
		}

		fixLayout.window.ApplyFixedLayoutSettings()

		persistError := config.PersistConfig()
		if persistError != nil {
			fmt.Fprintf(writer, "Error saving configuration: %s\n", persistError.Error())
			return
		}

		if successOutput != "" {
			fmt.Fprintln(writer, successOutput)
		}
	} else {
		fixLayout.PrintHelp(writer)
	}
}

func (fixLayout *FixLayout) Name() string {
	return "fixlayout"
}

func (fixLayout *FixLayout) Aliases() []string {
	return []string{"fix-layout"}
}

// PrintHelp prints a static help page for this command
func (fixLayout *FixLayout) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, fixLayoutDocumentation)
}
