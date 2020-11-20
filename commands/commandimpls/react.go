package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

const (
	reactionHelp = `[::b]NAME
	reaction - show existing message reactions or add new ones

[::b]SYNOPSIS
	[::b]React to a message with		react channelID messageID emoji

	Pressing 'w' above a message will show a dialog with the options to
	show the reactions or add a new one from the preset reactions.

	Emojis can be seen by issuing		react list

	Emoji can either be an unicode emoji or a string.

	Show reactions of a message			react show channelID messageID

	Show a dialog with the options		react dialog channelID messageID
	of showing a message's
	reaction or add a new one ('w' does this
	by default). This only allows to use a preset of favorite reactions, which
	can be customized in the config file. A maximum of 20 reactions will be
	randomly chosen from that list. They can be either unicode strings, such as 👍 , or
	custom emoji names, which can be gotten in discord by issuing \\emoji.



[::b]DESCRIPTION
	React to messages and see reactions. Press 'w' for an interactive dialog.
`
)

type Reaction struct {
	session *discordgo.Session
	window *ui.Window
	app *tview.Application
}

func NewReaction (s *discordgo.Session, window *ui.Window, app *tview.Application) *Reaction{
	return &Reaction{session: s, window: window, app: app}
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
		fmt.Fprintf(writer,"%s", r.List())
		return
	case "show":
		if len(parameters) < 3 {
			fmt.Fprintf(writer, "Not enough arguments provided.")
			return
		}
		emojis, l := r.Emojis(parameters[1], parameters[2])
		if l == "" {
			fmt.Fprintf(writer, "%s\n",emojis)
		} else {
			fmt.Fprintf(writer, "%s\n",l)
		}
		return

	case "dialog":
		add := "Add"
		show := "Show"
		text := ""
			r.app.QueueUpdateDraw(func() {
			r.window.ShowDialog(config.GetTheme().PrimitiveBackgroundColor,
				"Show reactions or add?", func(button string) {
					if button == show {
						text = r.Show(parameters[1], parameters[2])
						go func() {
							r.app.QueueUpdateDraw(func() {
								r.Dialog(text)
							})
						}()
						return
					} else if button == add {
						go func() {
						r.app.QueueUpdateDraw(func(){
						r.window.ShowDialog(config.GetTheme().PrimitiveBackgroundColor,
							"Select reaction to add. (Add more in your config)", func(button string) {
								err := r.Add(parameters[1],parameters[2], button)
								if err != nil {
									fmt.Fprintf(writer, "Could not add emoji %s, %e", button, err)
									return
								} else {
									fmt.Fprintf(writer, "Added reaction.")
									return
								}
							}, r.List()...)
							})
						}()
					}
					return
				}, show, add)
			})

	default:
		if len(parameters) < 3 {
			fmt.Fprintf(writer, "Not enough arguments provided for adding reaction.")
			return
		}

	}
}

func (r Reaction) List() []string {
	favEmojis := config.Current.FavoriteReactions
	if len(favEmojis) >= 20 {
		favEmojis = favEmojis[0:20]
	}
	return favEmojis
}

func (r Reaction) Show(c string, m string) string {
	emojis, _ := r.Emojis(c, m)
	return fmt.Sprintf("%s", emojis)
}

func (r Reaction) Add(c, m, emoji string) error {
	perms, _ := r.session.State.UserChannelPermissions(r.session.State.User.ID, c)
	if perms&discordgo.PermissionAddReactions != discordgo.PermissionAddReactions {
		return fmt.Errorf("You can't add reactions here.\n")
	}
	err := r.session.MessageReactionAdd(c, m, emoji)
	if err != nil {
		return fmt.Errorf("Some error ocurred.\n")
	}
	return err
}

func (r Reaction) Emojis(c string, m string) ([]string,string) {
		message, err := r.session.State.Message(c, m)
		msgLog := ""
		if err != nil {
			msgLog = fmt.Sprintf("There was an error obtaining the message.\n")
			return nil, msgLog
		}
		reactions := message.Reactions
		returnedReactions := make([]string, len(reactions))
		for _, reaction := range reactions {
			returnedReactions = append(returnedReactions,reaction.Emoji.Name)
		}
	return returnedReactions, msgLog
}

func (r Reaction) Dialog(emojis string) {
	if emojis == "" {
		return
	}
	r.window.ShowDialog(config.GetTheme().PrimitiveBackgroundColor, fmt.Sprintf("%s",emojis), func(button string) {

	}, "OK")
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
