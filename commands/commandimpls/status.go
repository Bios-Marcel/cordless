package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/discordgo"
)

const (
	statusHelpPage = `[::b]NAME
	status - view your or others status or update your own

[::b]SYNPOSIS
	[::b]status [subcommand]

[::b]DESCRPTION
	This command allows to either update your status or view a users status.
	For more information check the help pages of the subcommands.

[::]SUBCOMMANDS
	[::b]status-get (default)
		prints the status of the given user or yourself
	[::b]set-set
		updates your current status`

	statusSetHelpPage = `[::b]NAME
	status-set - allows updating your own status

[::b]SYNPOSIS
	[::b]status-set[::-] <online|idle|dnd|invisible>

[::b]DESCRPTION
	This command can be used to set your current online status to the
	value passed as the first parameter. Other users will immediately
	see your status update.

[::b]EXAMPLES
	[gray]$ status-set invisible`

	statusGetHelpPage = `[::b]NAME
	status-get - prints your current status or the status of the given user

[::b]SYNPOSIS
	[::b]status-get[::-] [Username|Username#NNNN|UserID[]

[::b]DESCRPTION
	This command prints either your current status of no value was passed
	or the status of the passed user, if the presence for that user could
	be found. Due to a problem with the presences, this command might randomly
	fail when trying to query specific users.

[::b]EXAMPLES
	[gray]$ status-get
	[yellow]idle

	[gray]$ status-get Marcel#7299
	[red]Do not disturb`
)

type StatusCmd struct {
	statusGetCmd *StatusGetCmd
	statusSetCmd *StatusSetCmd
}

type StatusGetCmd struct {
	session *discordgo.Session
}

type StatusSetCmd struct {
	session *discordgo.Session
}

func NewStatusCommand(statusGetCmd *StatusGetCmd, statusSetCmd *StatusSetCmd) *StatusCmd {
	return &StatusCmd{
		statusGetCmd: statusGetCmd,
		statusSetCmd: statusSetCmd,
	}
}

func NewStatusGetCommand(session *discordgo.Session) *StatusGetCmd {
	return &StatusGetCmd{session}
}

func NewStatusSetCommand(session *discordgo.Session) *StatusSetCmd {
	return &StatusSetCmd{session}
}

func statusToString(status discordgo.Status) string {
	switch status {
	case discordgo.StatusOnline:
		return "[green]Online[white]"
	case discordgo.StatusDoNotDisturb:
		return "[red]Do not disturb[white]"
	case discordgo.StatusIdle:
		return "[yellow]Idle[white]"
	case discordgo.StatusInvisible:
		return "[gray]Invisible[white]"
	case discordgo.StatusOffline:
		return "[gray]Offline[white]"
	default:
		return "Unknown status"
	}
}

func (cmd *StatusGetCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) > 1 {
		fmt.Fprintln(writer, "[red]Invalid parameters")
		cmd.PrintHelp(writer)
		return
	}

	if len(parameters) == 0 {
		fmt.Fprintln(writer, statusToString(cmd.session.State.Settings.Status))
		return
	}

	input := parameters[0]
	var matches []*discordgo.Presence
	for _, presence := range cmd.session.State.Presences {
		user := presence.User
		if user.ID == input || user.Username == input || user.String() == input {
			matches = append(matches, presence)
		}
	}

	if len(matches) == 0 {
		fmt.Fprintf(writer, "[red]No match for '%s'.\n", input)
	} else if len(matches) > 1 {
		fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
		fmt.Fprintln(writer, "The following matches were found:")
		for _, match := range matches {
			fmt.Fprintf(writer, "\t%s\n", match.User.String())
		}
	} else {
		fmt.Fprintln(writer, statusToString(matches[0].Status))
	}
}

func (cmd *StatusSetCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 || len(parameters) > 1 {
		fmt.Fprintln(writer, "[red]Invalid parameters")
		cmd.PrintHelp(writer)
		return
	}

	var settingStatusError error
	var updatedSettings *discordgo.Settings

	switch strings.ToLower(parameters[0]) {
	case "online", "available":
		updatedSettings, settingStatusError = cmd.session.UserUpdateStatus(discordgo.StatusOnline)
	case "dnd", "donotdisturb", "busy":
		updatedSettings, settingStatusError = cmd.session.UserUpdateStatus(discordgo.StatusDoNotDisturb)
	case "idle":
		updatedSettings, settingStatusError = cmd.session.UserUpdateStatus(discordgo.StatusIdle)
	case "invisible":
		updatedSettings, settingStatusError = cmd.session.UserUpdateStatus(discordgo.StatusInvisible)
	default:
		fmt.Fprintf(writer, "[red]Invalid status: '%s'\n", strings.ToLower(parameters[0]))
		cmd.PrintHelp(writer)
	}

	if settingStatusError != nil {
		fmt.Fprintf(writer, "[red]Error setting status:\n\t[red]'%s'\n", settingStatusError.Error())
	} else if updatedSettings != nil {
		cmd.session.State.Settings = updatedSettings
	}
}

func (cmd *StatusCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) >= 1 {
		if parameters[0] == "set" || parameters[0] == "update" {
			cmd.statusSetCmd.Execute(writer, parameters[1:])
		} else if parameters[0] == "get" {
			cmd.statusGetCmd.Execute(writer, parameters[1:])
		} else {
			cmd.statusGetCmd.Execute(writer, parameters)
		}
	} else {
		cmd.statusGetCmd.Execute(writer, parameters)
	}
}

func (cmd *StatusCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, statusHelpPage)
}

func (cmd *StatusSetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, statusSetHelpPage)
}

func (cmd *StatusGetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, statusGetHelpPage)
}
