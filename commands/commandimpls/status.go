package commandimpls

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Bios-Marcel/discordemojimap"
	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

const (
	statusHelpPage = `[::b]NAME
	status - view your or others status or update your own

[::b]SYNOPSIS
	[::b]status [subcommand]

[::b]DESCRIPTION
	This command allows to either update your status or view a users status.
	For more information check the help pages of the subcommands.

[::b]SUBCOMMANDS
	[::b]status-get (default)
		prints the status of the given user or yourself
	[::b]status-set
		updates your current status
	[::b]status-set-custom
		set your status to a custom text`

	statusSetHelpPage = `[::b]NAME
	status-set - allows updating your own status

[::b]SYNOPSIS
	[::b]status-set[::-] <online|idle|dnd|invisible>

[::b]DESCRIPTION
	This command can be used to set your current online status to the
	value passed as the first parameter. Other users will immediately
	see your status update.

[::b]EXAMPLES
	[gray]$ status-set invisible`

	statusGetHelpPage = `[::b]NAME
	status-get - prints your current status or the status of the given user

[::b]SYNOPSIS
	[::b]status-get[::-] [Username|Username#NNNN|UserID[]

[::b]DESCRIPTION
	This command prints either your current status of no value was passed
	or the status of the passed user, if the presence for that user could
	be found. Due to a problem with the presences, this command might randomly
	fail when trying to query specific users.

[::b]EXAMPLES
	[gray]$ status-get
	[yellow]idle

	[gray]$ status-get Marcel#7299
	[red]Do not disturb`

	statusSetCustomHelpPage = `[::b]NAME
	status-set-custom - set a custom status text

[::b]SYNOPSIS
	[::b]status-set-custom[::-] [OPTION[]...

[::b]DESCRIPTION
	This command allows you to set a custom status.

[::b]OPTIONS
	[::b]-s, --status
		status message
	[::b]-e, --emoji
		emoji in your status
	[::b]-i, --expire, --expiry <s|m|h>
		time that the status expires after

[::b]EXAMPLES
	[gray]$ status-set-custom -s "shining bright" -e :sun:
	[gray]$ status-set-custom -s "shining bright" -e 🌞
	[gray]$ status-set-custom -s test -i 1h`
)

type StatusCmd struct {
	statusGetCmd       *StatusGetCmd
	statusSetCmd       *StatusSetCmd
	statusSetCustomCmd *StatusSetCustomCmd
}

type StatusGetCmd struct {
	session *discordgo.Session
}

type StatusSetCmd struct {
	session *discordgo.Session
}

type StatusSetCustomCmd struct {
	session *discordgo.Session
}

func NewStatusCommand(statusGetCmd *StatusGetCmd, statusSetCmd *StatusSetCmd, statusSetCustomCmd *StatusSetCustomCmd) *StatusCmd {
	return &StatusCmd{
		statusGetCmd:       statusGetCmd,
		statusSetCmd:       statusSetCmd,
		statusSetCustomCmd: statusSetCustomCmd,
	}
}

func NewStatusGetCommand(session *discordgo.Session) *StatusGetCmd {
	return &StatusGetCmd{session}
}

func NewStatusSetCommand(session *discordgo.Session) *StatusSetCmd {
	return &StatusSetCmd{session}
}

func NewStatusSetCustomCommand(session *discordgo.Session) *StatusSetCustomCmd {
	return &StatusSetCustomCmd{session}
}

func statusColor(status discordgo.Status) string {
	switch status {
	case discordgo.StatusOnline:
		return "green"
	case discordgo.StatusDoNotDisturb:
		return "red"
	case discordgo.StatusIdle:
		return "yellow"
	case discordgo.StatusInvisible, discordgo.StatusOffline:
		return "gray"
	default:
		return "gray"
	}
}

func statusToString(status discordgo.Status) string {
	statusString := "[" + statusColor(status) + "]"
	switch status {
	case discordgo.StatusOnline:
		statusString += "Online"
	case discordgo.StatusDoNotDisturb:
		statusString += "Do not disturb"
	case discordgo.StatusIdle:
		statusString += "Idle"
	case discordgo.StatusInvisible:
		statusString += "Invisible"
	case discordgo.StatusOffline:
		statusString += "Offline"
	default:
		statusString += "Unknown status"
	}
	return statusString
}

func (cmd *StatusGetCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) > 1 {
		fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Invalid parameters")
		cmd.PrintHelp(writer)
		return
	}

	if len(parameters) == 0 {
		fmt.Fprintf(writer, statusToString(cmd.session.State.Settings.Status))

		customStatus := cmd.session.State.Settings.CustomStatus
		if len(customStatus.Text) > 0 || len(customStatus.EmojiName) > 0 {
			fmt.Fprintf(writer, ": %s %s", customStatus.EmojiName, customStatus.Text)
		}
		fmt.Fprintf(writer, "[white]\n")
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
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]No match for '%s'.\n", input)
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
	if cmd.session.State.User.Bot {
		fmt.Fprintln(writer, "[red]This command can't be used by bots due to Discord API restrictions.")
		return
	}

	if len(parameters) == 0 || len(parameters) > 1 {
		fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Invalid parameters")
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
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Invalid status: '%s'\n", strings.ToLower(parameters[0]))
		cmd.PrintHelp(writer)
	}

	if settingStatusError != nil {
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error setting status:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]'%s'\n", settingStatusError.Error())
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
		} else if parameters[0] == "set-custom" || parameters[0] == "custom-set" {
			cmd.statusSetCustomCmd.Execute(writer, parameters[1:])
		} else {
			cmd.statusGetCmd.Execute(writer, parameters)
		}
	} else {
		cmd.statusGetCmd.Execute(writer, parameters)
	}
}

func (cmd *StatusSetCustomCmd) Execute(writer io.Writer, parameters []string) {
	if cmd.session.State.User.Bot {
		fmt.Fprintln(writer, "[red]This command can't be used by bots due to Discord API restrictions.")
		return
	}

	if len(parameters) == 0 {
		fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Invalid parameters")
		cmd.PrintHelp(writer)
		return
	}

	errorColor := tviewutil.ColorToHex(config.GetTheme().ErrorColor)
	var customStatus discordgo.CustomStatus
	for index, param := range parameters {
		switch param {
		case "-s", "--status":
			if index != len(parameters)-1 {
				customStatus.Text = parameters[index+1]
			} else {
				fmt.Fprintf(writer, "[%s]Error, you didn't supply a status\n", errorColor)
				return
			}
		case "-e", "--emoji":
			if index != len(parameters)-1 {
				if discordemojimap.ContainsEmoji(parameters[index+1]) {
					customStatus.EmojiName = parameters[index+1]
				} else if emoji := discordemojimap.Replace(parameters[index+1]); emoji != parameters[index+1] {
					customStatus.EmojiName = emoji
				} else {
					fmt.Fprintf(writer, "[%s]Invalid emoji\n", errorColor)
					return
				}
			}
		case "-i", "--expire", "--expiry":
			if m, _ := regexp.MatchString(`\d+(s|m|h)`, parameters[index+1]); m {
				lastIndex := len(parameters[index+1]) - 1
				num, err := strconv.Atoi(parameters[index+1][:lastIndex])
				if err != nil {
					fmt.Fprintf(writer, "[%s]Invalid expiry\n", errorColor)
					return
				}

				now := time.Now().UTC()
				switch parameters[index+1][lastIndex] {
				case 's':
					now = now.Add(time.Second * time.Duration(num))
				case 'm':
					now = now.Add(time.Minute * time.Duration(num))
				case 'h':
					now = now.Add(time.Hour * time.Duration(num))
				default:
					fmt.Fprintf(writer, "[%s]Invalid time character: %c != <s|m|h>\n", errorColor, parameters[index+1][lastIndex])
					return
				}
				customStatus.ExpiresAt = now.Format(time.RFC3339Nano)
			}
		}
	}

	settings, err := cmd.session.UserUpdateStatusCustom(customStatus)
	if err != nil {
		fmt.Fprintf(writer, "[%s]Error updating custom status:\n\t[%s]'%s'\n", errorColor, errorColor, err.Error())
		return
	} else if settings != nil {
		cmd.session.State.Settings = settings
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

func (cmd *StatusSetCustomCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, statusSetCustomHelpPage)
}

func (cmd *StatusSetCustomCmd) Name() string {
	return "status-set-custom"
}

func (cmd *StatusSetCmd) Name() string {
	return "status-set"
}

func (cmd *StatusGetCmd) Name() string {
	return "status-get"
}

func (cmd *StatusCmd) Name() string {
	return "status"
}

func (cmd *StatusSetCustomCmd) Aliases() []string {
	return []string{"status-custom"}
}

func (cmd *StatusSetCmd) Aliases() []string {
	return []string{"status-update"}
}

func (cmd *StatusGetCmd) Aliases() []string {
	return nil
}

func (cmd *StatusCmd) Aliases() []string {
	return nil
}
