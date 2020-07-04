package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/version"
)

const versionDocumentation = `[::b]NAME
	version - Displays the version of cordless you are running

[::b]SYNOPSIS
	[::b]version

[::b]DESCRIPTION
	This command displays the version of cordless that you are running.
	The version is the in the format Year-Month-Day and is the date on which
	the version you are using was built. If you have manually built cordless,
	your installation might contain commits which aren't part of the stated
	version.`

type VersionCmd struct{}

func NewVersionCommand() *VersionCmd {
	return &VersionCmd{}
}

func (cmd VersionCmd) Execute(writer io.Writer, parameters []string) {
	fmt.Fprintf(writer, "You are running cordless version %s\n", version.Version)
}

func (cmd VersionCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, versionDocumentation)
}

func (cmd VersionCmd) Name() string {
	return "version"
}

func (cmd VersionCmd) Aliases() []string {
	return nil
}
