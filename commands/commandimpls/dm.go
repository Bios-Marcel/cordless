package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

const (
	dmOpenHelpPage = `[::b]NAME
	dm-open - open or create a dm channel

[::b]SYNOPSIS
	[::b]dm-open <Username|Username#NNNN|User-ID>

[::b]DESCRIPTION
	If the user can be found in your local cache, a new dm channel is created
	or an existing one loaded.`
)

// DMOpenCmd allows to open / create DM channels.
type DMOpenCmd struct {
	session *discordgo.Session
	window  *ui.Window
}

// NewDMOpenCmd creates a ready to use command to open / create DM channels.
func NewDMOpenCmd(session *discordgo.Session, window *ui.Window) *DMOpenCmd {
	return &DMOpenCmd{session, window}
}

// Execute runs the command piping its output into the supplied writer.
func (cmd *DMOpenCmd) Execute(writer io.Writer, parameters []string) {
	//We expect exactly one user
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
	}

	//FIXME Pretty much copied from friends.go. Can i somehow abstract this away?

	users, err := cmd.session.State.Users()
	if err != nil {
		fmt.Fprintf(writer, "An error occured during commandexecution (%s).\n", err.Error())
		return
	}

	input := parameters[0]
	var matches []*discordgo.User
	for _, user := range users {
		if user.ID == input || user.Username == input || user.String() == input {
			matches = append(matches, user)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintln(writer, "No user was found.")
	} else if len(matches) == 1 {
		user := matches[0]

		//Can't message yourself, goon!
		if user.ID == cmd.session.State.User.ID {
			fmt.Fprintln(writer, "You can't message yourself.")
			return
		}

		//If there's an existing channel, we use that and avoid unnecessary traffic.
		existingChannel := discordutil.FindDMChannelWithUser(cmd.session.State, user.ID)
		if existingChannel != nil {
			cmd.window.SwitchToPrivateChannel(existingChannel)
			return
		}

		newChannel, createError := cmd.session.UserChannelCreate(user.ID)
		if createError != nil {
			fmt.Fprintf(writer, "Error opening DM (%s).\n", createError.Error())
		} else {
			cmd.window.SwitchToPrivateChannel(newChannel)
		}
	} else {
		fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
		fmt.Fprintln(writer, "The following matches were found:")
		for _, match := range matches {
			fmt.Fprintln(writer, "  "+match.String())
		}
	}
}

// PrintHelp prints a static help page for this command
func (cmd *DMOpenCmd) PrintHelp(writer io.Writer) {

}

// Name returns the primary name for this command. This name will also be
// used for listing the command in the commandlist.
func (cmd *DMOpenCmd) Name() string {
	return "dm-open"
}

// Aliases are a list of aliases for this command. There might be none.
func (cmd *DMOpenCmd) Aliases() []string {
	return []string{"dm-start", "dm-new", "dm-show"}
}
