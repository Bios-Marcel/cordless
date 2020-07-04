package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/discordemojimap"
	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/commands"
)

type NickSetCmd struct {
	session     *discordgo.Session
	clientState commands.ClientState
}

func NewNickSetCmd(session *discordgo.Session, clientState commands.ClientState) *NickSetCmd {
	return &NickSetCmd{
		session:     session,
		clientState: clientState,
	}
}

func (cmd NickSetCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		commands.PrintError(writer, "Error setting nickname", "Either use 'nick-set <NAME>' or 'nick-set \"\"' to explicitly reset the nickname.")
		return
	}

	selectedGuild := cmd.clientState.GetSelectedGuild()
	if selectedGuild == nil {
		commands.PrintError(writer, "Error setting nickname", "No guild selected")
	} else {
		newName := discordemojimap.Replace(strings.TrimSpace(parameters[0]))
		setError := cmd.session.GuildMemberNickname(selectedGuild.ID, "@me", newName)
		if setError != nil {
			commands.PrintError(writer, "Error setting nickname", setError.Error())
		}
	}
}

func (cmd NickSetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "TODO")
}

func (cmd NickSetCmd) Name() string {
	return "nick-set"
}

func (cmd NickSetCmd) Aliases() []string {
	return nil
}
