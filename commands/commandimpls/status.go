package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/discordgo"
)

const statusDocumentation = `[orange][::u]# status[white]

Current status: %s

The status command always setting or reading the current status of your user.

If not supplied with any parameters, it will just print the help page and the current status.

If supplied with a valid status value, that value will be set as the new status.

Valid values:
    * online
	* dnd
	* idle
    * invisible
`

// Status represents the command that allows the user to read and write his
// user status on discord.
type Status struct {
	session *discordgo.Session
}

// NewStatusCommand creates a new ready to use status command.
func NewStatusCommand(session *discordgo.Session) *Status {
	return &Status{
		session: session,
	}
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
	}

	return "Unknown status"
}

// Execute runs the command piping its output into the supplied writer.
func (status *Status) Execute(writer io.Writer, parameters []string) {
	var settingStatusError error
	var updatedSettings *discordgo.Settings

	if len(parameters) == 0 {
		fmt.Fprintf(writer, "Current status: %s \n", statusToString(status.session.State.Settings.Status))
	} else if len(parameters) == 1 {
		switch strings.ToLower(parameters[0]) {
		case "online", "available":
			updatedSettings, settingStatusError = status.session.UserUpdateStatus(discordgo.StatusOnline)
		case "dnd", "donotdisturb", "busy":
			updatedSettings, settingStatusError = status.session.UserUpdateStatus(discordgo.StatusDoNotDisturb)
		case "idle":
			updatedSettings, settingStatusError = status.session.UserUpdateStatus(discordgo.StatusIdle)
		case "invisible":
			updatedSettings, settingStatusError = status.session.UserUpdateStatus(discordgo.StatusInvisible)
		default:
			status.PrintHelp(writer)
		}
	} else {
		status.PrintHelp(writer)
	}

	if settingStatusError != nil {
		fmt.Fprintf(writer, "Error setting status: '%s'\n", settingStatusError.Error())
	} else if updatedSettings != nil {
		fmt.Fprintf(writer, "Setting status to '%s'\n", statusToString(updatedSettings.Status))
		status.session.State.Settings = updatedSettings
	}
}

// Name represents this commands indentifier.
func (status *Status) Name() string {
	return "status"
}

// PrintHelp the help page for this command. On top of the usual static content
// this help page also contains the current user state.
func (status *Status) PrintHelp(writer io.Writer) {
	fmt.Fprintf(writer, statusDocumentation, statusToString(status.session.State.Settings.Status))
}
