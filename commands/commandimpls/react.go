package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/discordgo"
)

const (
	reactionHelp = `[::b]NAME
	reaction - show existing message reactions or add new ones

[::b]SYNOPSIS
	[::b]React to a message with		react channelID messageID emoji
	You can get the first two in chat view by pressing 'w' above a message while in vim mode.
	Emojis can be seen by issuing		react list
	Emoji can either be an unicode emoji or a string.
	Show reactions of a message			react show channelID messageID



[::b]DESCRIPTION
	React to messages and see reactions.
`
)

type Reaction struct {
	session *discordgo.Session
}

func NewReaction (s *discordgo.Session) *Reaction{
	return &Reaction{session: s}
}

// PrintHelp prints a static help page for this command
func (r Reaction) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, reactionHelp)
}

func (r Reaction) Execute(writer io.Writer, parameters []string) {
	if len(parameters) < 1 {
		fmt.Fprintf(writer, "Not enough parameters. Issue man reaction to see help.")
		return
	}
	switch parameters[0] {
	case "list":
		fmt.Fprintf(writer,"TODO")
	case "show":
		if len(parameters) < 3 {
			fmt.Fprintf(writer, "Not enough arguments provided.")
			return
		}
		message, err := r.session.ChannelMessage(parameters[1], parameters[2])
		if err != nil {
			fmt.Fprintf(writer, "There was an error obtaining the message.\n%e",err)
			return
		}
		reactions := message.Reactions
		for _, reaction := range reactions {
			fmt.Fprintf(writer, "%s\n",reaction.Emoji.Name)
		}
		return

	default:
		if len(parameters) < 3 {
			fmt.Fprintf(writer, "Not enough arguments provided for adding reaction.")
			return
		}
		err := r.session.MessageReactionAdd(parameters[0], parameters[1], parameters[2])
		if err != nil {
			fmt.Fprintf(writer, "Something wrong ocurred while adding reaction. \n")
			return
		}
		fmt.Fprintf(writer, "Added reaction successfully.")

	}
}

// Name returns the primary name for this command. This name will also be
// used for listing the command in the commandlist.
func (r Reaction) Name() string {
	return "reaction"
}

// Aliases are a list of aliases for this command. There might be none.
func (r Reaction) Aliases() []string {
	return []string{"react", "reaction"}
}
