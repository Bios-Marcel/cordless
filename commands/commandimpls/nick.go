package commandimpls

import (
	"fmt"
	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/discordgo"
	"io"
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
		commands.PrintError(writer, "Error setting nickname", "Usage: nick-set <NAME>")
		return
	}

	selectedGuild := cmd.clientState.GetSelectedGuild()
	if selectedGuild == nil {
		commands.PrintError(writer, "Error setting nickname", "No guild selected")
	} else {
		setError := cmd.session.GuildMemberNickname(selectedGuild.ID, "@me", parameters[0])
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
