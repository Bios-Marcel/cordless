package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/discordgo"
)

type Friends struct {
	session *discordgo.Session
}

func NewFriendsCommand(session *discordgo.Session) *Friends {
	return &Friends{
		session: session,
	}
}

func (f *Friends) Execute(writer io.Writer, parameters []string) {
	switch parameters[0] {
	case "invites", "requests":
		var incomming, outgoing string
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeIncommingRequest {
				incomming += "  " + rel.User.String() + "\n"
			} else if rel.Type == discordgo.RelationTypeOutgoingRequest {
				incomming += "  " + rel.User.String() + "\n"
			}
		}

		fmt.Fprintln(writer, "Incomming requests:")
		if incomming != "" {
			fmt.Fprintln(writer, incomming)
		} else {
			fmt.Fprintln(writer, "No incomming requests.")
		}

		fmt.Fprintln(writer, "Outgoing requests:")
		if outgoing != "" {
			fmt.Fprintln(writer, outgoing)
		} else {
			fmt.Fprintln(writer, "No outgoing requests.")
		}
	case "accept", "agree":
		accept := parameters[1]
		var matches []*discordgo.Relationship
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeIncommingRequest {
				if rel.User.ID == accept || rel.User.Username == accept || rel.User.String() == accept {
					matches = append(matches, rel)
				}
			}
		}
		if len(matches) == 0 {
			fmt.Fprintln(writer, "Oopsie woopsie, there was a fuckie wuckie.")
		} else if len(matches) == 1 {
			fmt.Fprintln(writer, "Accepting friends request of "+matches[0].User.String())
			acceptErr := f.session.RelationshipFriendRequestAccept(matches[0].User.ID)
			if acceptErr != nil {
				fmt.Fprintf(writer, "Error accepting friendsrequest (%s).\n", acceptErr.Error())
			} else {
				fmt.Fprintln(writer, matches[0].User.String()+" is now your friend.")
			}
		} else {
			fmt.Fprintln(writer, "Oopsie woopsie, there was a fuckie wuckie.")
		}
	}
}

func (f *Friends) Name() string {
	return "friends"
}

func (f *Friends) PrintHelp(writer io.Writer) {
	panic("not implemented")
}
