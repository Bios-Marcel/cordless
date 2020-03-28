package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

const serverHelpPage = `[::b]NAME
	server - allows you to join or leave a server

[::b]SYNOPSIS
	[::b]server[::-] <subcommand <args>>

[::b]DESCRIPTION
	The server command allows you to join a new server or leave one that you
	are already a part of. What this command can't do is administrating a
	server in any way.

[::b]SUBCOMMANDS
	[::b]server-join
		joins the server using the given invitation
	[::b]server-leave
		leaves the given server`

const serverJoinHelpPage = `[::b]NAME
	server-join - allows you to join a server

[::b]SYNOPSIS
	[::b]server-join[::-] <InviteCode|InviteURL>

[::b]DESCRIPTION
	This command will take a invite code or an invite URl and attempt joining
	the server behind it.

[::b]EXAMPLES
	[gray]$ server-join https://discord.gg/JDScUK
	[gray]$ server-join discord.gg/JDScUK
	[gray]$ server-join JDScUK`

const serverLeaveHelpPage = `[::b]NAME
	server-leave - allows you to leave a server

[::b]SYNOPSIS
	[::b]server-leave[::-] <ID|Name>

[::b]DESCRIPTION
	This command will take a server ID or it's name and leave that server.

[::b]EXAMPLES
	[gray]$ server-leave 118456055842734083
	[gray]$ server-leave "Discord Gophers"
	[gray]$ server-leave Nirvana`

const serverCreateHelpPage = `[::b]NAME
	server-create - allows you to create a new server

[::b]SYNOPSIS
	[::b]server-create[::-] <Name>

[::b]DESCRIPTION
	This command will take a name and create a server using that name. You'll
	be told the name and the ID on successful creation.

[::b]EXAMPLES
	[gray]$ server-create "Hello world"
	[gray]$ server-create MyServerName`

type ServerCmd struct {
	serverJoinCmd   *ServerJoinCmd
	serverLeaveCmd  *ServerLeaveCmd
	serverCreateCmd *ServerCreateCmd
}

type ServerJoinCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

type ServerLeaveCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

type ServerCreateCmd struct {
	session *discordgo.Session
}

func NewServerCommand(serverJoinCmd *ServerJoinCmd, serverLeaveCmd *ServerLeaveCmd, serverCreateCmd *ServerCreateCmd) *ServerCmd {
	return &ServerCmd{serverJoinCmd, serverLeaveCmd, serverCreateCmd}
}

func NewServerJoinCommand(window *ui.Window, session *discordgo.Session) *ServerJoinCmd {
	return &ServerJoinCmd{window, session}
}

func NewServerLeaveCommand(window *ui.Window, session *discordgo.Session) *ServerLeaveCmd {
	return &ServerLeaveCmd{window, session}
}

func NewServerCreateCommand(session *discordgo.Session) *ServerCreateCmd {
	return &ServerCreateCmd{session}
}
func (cmd *ServerCreateCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}

	newGuild, createError := cmd.session.GuildCreate(parameters[0])
	if createError != nil {
		commands.PrintError(writer, "Couldn't create server", createError.Error())
	} else {
		fmt.Fprintf(writer, "New server '%s' with ID '%s' has been created.", newGuild.Name, newGuild.ID)
	}
}

func (cmd *ServerCreateCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, serverCreateHelpPage)
}

func (cmd *ServerCreateCmd) Name() string {
	return "server-create"
}

func (cmd *ServerCreateCmd) Aliases() []string {
	return []string{"guild-create", "guild-new", "server-new"}
}

func (cmd *ServerCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, serverHelpPage)
}

func (cmd *ServerCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		cmd.PrintHelp(writer)
	} else {
		combinedName := cmd.Name() + "-" + parameters[0]
		if commands.CommandEquals(cmd.serverJoinCmd, combinedName) {
			cmd.serverJoinCmd.Execute(writer, parameters[1:])
		} else if commands.CommandEquals(cmd.serverLeaveCmd, combinedName) {
			cmd.serverLeaveCmd.Execute(writer, parameters[1:])
		} else if commands.CommandEquals(cmd.serverCreateCmd, combinedName) {
			cmd.serverCreateCmd.Execute(writer, parameters[1:])
		} else {
			cmd.PrintHelp(writer)
		}
	}
}

func (cmd *ServerJoinCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, serverJoinHelpPage)
}

func (cmd *ServerJoinCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}

	if cmd.session.State.User.Bot {
		fmt.Fprintln(writer, "[red]This command can't be used by bots due to Discord API restrictions.")
		return
	}

	input := parameters[0]
	lastSlash := strings.LastIndex(input, "/")
	var inviteID string
	if lastSlash == -1 {
		inviteID = input
	} else {
		inviteID = input[lastSlash+1:]
	}

	invite, err := cmd.session.InviteAccept(inviteID)
	if err != nil {
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error accepting invite with ID '%s':\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", inviteID, err)
	} else {
		fmt.Fprintf(writer, "Joined server '%s'\n", invite.Guild.Name)
	}
}

func (cmd *ServerLeaveCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, serverLeaveHelpPage)
}

func (cmd *ServerLeaveCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}

	input := parameters[0]
	matches := make([]*discordgo.Guild, 0)
	for _, guild := range cmd.session.State.Guilds {
		if guild.ID == input || guild.Name == input {
			matches = append(matches, guild)
		}
	}

	if len(matches) == 1 {
		foundGuild := matches[0]
		if foundGuild.OwnerID == cmd.session.State.User.ID {
			fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]You can't leave a server you own.")
			return
		}

		err := cmd.session.GuildLeave(foundGuild.ID)
		if err != nil {
			fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error leaving server '%s':\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", matches[0].Name, err)
		} else {
			fmt.Fprintf(writer, "Left server '%s'.\n", matches[0].Name)
		}
	} else if len(matches) == 0 {
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]No server with the ID or Name '%s' was found.\n", input)
	} else {
		fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
		fmt.Fprintln(writer, "The following matches were found:")
		for _, match := range matches {
			fmt.Fprintf(writer, "ID: %s\tName: %s\n", match.ID, match.Name)
		}
	}
}

func (cmd *ServerCmd) Name() string {
	return "server"
}

func (cmd *ServerJoinCmd) Name() string {
	return "server-join"
}

func (cmd *ServerLeaveCmd) Name() string {
	return "server-leave"
}

func (cmd *ServerCmd) Aliases() []string {
	return []string{"guild"}
}

func (cmd *ServerJoinCmd) Aliases() []string {
	return []string{"guild-join", "guild-accept", "guild-enter", "server-accept", "server-enter"}
}

func (cmd *ServerLeaveCmd) Aliases() []string {
	return []string{"guild-leave", "guild-exit", "guild-quit", "server-exit", "server-quit"}
}
