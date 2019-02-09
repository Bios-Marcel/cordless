package commands

import (
	"io"
)

// Command represents a command that is executable by the user.
type Command interface {
	// Name represents this commands indentifier.
	Name() string

	// Execute runs the command piping its output into the supplied writer.
	Execute(writer io.Writer, parameters []string)

	// PrintHelp prints a static help page for this command
	PrintHelp(writer io.Writer)
}
