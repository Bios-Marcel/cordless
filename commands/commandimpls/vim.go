package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/util/vim"
)

const (
	vimOpenHelpPage = `[::b]NAME
	vim-mode - minor vim mode for cordless

[::b]SYNOPSIS
	[::b]Normal Mode: Navigate around the containers with hjkl.
	Perform some usual commands inside text box input, as pasting, moving cursor or such.
	Press ESC anywhere to return to normal mode.

	[::b]Insert Mode: Type inside box input, perform actions inside chatview.
	Insert without any key restriction inside the text box using insert mode.
	Inside chat view, perform useful commands such as editing message "i" or replying to user "a"

	[::b]Visual Mode: Move around everywhere with vim keys.
	This mode allows to use hjkl pretty much anywhere inside the app. Due to some restrictions, this is
	the only mode that officially supports using hjkl anywhere.
	Also allows using same commands as insert mode inside chat view, or selecting text inside text input.


[::b]DESCRIPTION
	This is a minor mode for vim. See all the shorcuts with Ctrl K, and edit them inside shortcuts/shortcuts.go
	Toggle vim with vim on/off/enable/disable
`
)

type VimHelp struct {
	Config *config.Config
}

func NewVimCmd(c *config.Config) *VimHelp {
	return &VimHelp{Config: c}
}

// PrintHelp prints a static help page for this command
func (v VimHelp) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, vimOpenHelpPage)
}

func (v VimHelp) Execute(writer io.Writer, parameters []string) {
	if len(parameters) < 1 {
		fmt.Fprintf(writer, "You did not specify any parameter for this command.")
		return
	}
	switch parameters[0] {
	case "on","enable":
		v.Config.VimMode.CurrentMode = vim.NormalMode
		fmt.Fprintf(writer, "Vim mode has been enabled.")
	case "off","disable":
		v.Config.VimMode.CurrentMode = vim.NormalMode
		v.Config.VimMode.CurrentMode = vim.Disabled
		fmt.Fprintf(writer, "Vim mode has been disabled.")
	default:
		fmt.Fprintf(writer, "Parameter %s not recognized.", parameters[0])
	}
}

// Name returns the primary name for this command. This name will also be
// used for listing the command in the commandlist.
func (v VimHelp) Name() string {
	return "vim"
}

// Aliases are a list of aliases for this command. There might be none.
func (v VimHelp) Aliases() []string {
	return []string{"vim", "vim-mode"}
}
