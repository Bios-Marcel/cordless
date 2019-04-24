package commandimpls

import (
	"fmt"
	"io"
	"unicode"

	"github.com/Bios-Marcel/discordgo"
)

const friendsDocumentation = `[orange][::u]# friends[white]

The friends command allows you to manage your friends on discord. You can add
new friends by sending or accept friendsrequests. You can also see your current
requets, that goes for the incomming and the outgoing ones.

The friend currently command offers 3 subcommands:
  * accept   - accept a friends-request
  * befriend - send a friends-request
  * requests - shows all current requests

The following features are currently unsupported:
  * Deleting friends
  * Bocking users
  * Unblocking users
`

// Friends is the command for managing discord friends.
type Friends struct {
	session *discordgo.Session
}

// NewFriendsCommand creates a new ready to use friends command instance.
func NewFriendsCommand(session *discordgo.Session) *Friends {
	return &Friends{
		session: session,
	}
}

// Execute handles all input for the friends command.
func (f *Friends) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		f.PrintHelp(writer)
		return
	}

	switch parameters[0] {
	case "requests", "invites", "outstanding", "unanswered":
		var incomming, outgoing string
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeIncommingRequest {
				incomming += "  " + rel.User.String() + "\n"
			} else if rel.Type == discordgo.RelationTypeOutgoingRequest {
				outgoing += "  " + rel.User.String() + "\n"
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
		if len(parameters) != 2 {
			fmt.Fprintln(writer, "Usage: friends accept <Username|Username#NNNN|UserID")
			return
		}

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
			fmt.Fprintf(writer, "No matches for '%s' found.\n", accept)
		} else if len(matches) == 1 {
			fmt.Fprintln(writer, "Accepting friends request of "+matches[0].User.String())
			acceptErr := f.session.RelationshipFriendRequestAccept(matches[0].User.ID)
			if acceptErr != nil {
				fmt.Fprintf(writer, "Error accepting friendsrequest (%s).\n", acceptErr.Error())
			} else {
				fmt.Fprintln(writer, matches[0].User.String()+" is now your friend.")
			}
		} else {
			fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", accept)
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.User.String())
			}
		}
	case "befriend", "send", "ask", "invite", "request":
		if len(parameters) != 2 {
			fmt.Fprintln(writer, "Usage: friends befriend <Username|Username#NNNN|UserID")
		}

		//Iterate over all available users and find one that fits, if we were
		//successful, we send a friendsrequest. Otherwise we check if the input
		//might have been a user idea, lookup the user and do a request.
		input := parameters[1]
		users, err := f.session.State.Users()
		var matches []*discordgo.User

		if err != nil {
			fmt.Fprintf(writer, "An error occured during commandexecution (%s).\n", err.Error())
			return
		}

		for _, user := range users {
			if user.ID == input || user.Username == input || user.String() == input {
				matches = append(matches, user)
			}
		}

		if len(matches) == 0 {
			for _, char := range input {
				if !unicode.IsNumber(char) {
					fmt.Fprintf(writer, "No matches for '%s' found.\n", input)
					return
				}
			}

			requestError := f.session.RelationshipFriendRequestSend(input)
			if requestError != nil {
				fmt.Fprintf(writer, "Error sending friends-request (%s).\n", requestError)
			} else {
				fmt.Fprintln(writer, "Friends-request has been sent.")
			}
		} else if len(matches) == 1 {
			user := matches[0]
			fmt.Fprintln(writer, "Sending friendsrequest to "+user.String())
			requestError := f.session.RelationshipFriendRequestSend(user.ID)
			if requestError != nil {
				fmt.Fprintf(writer, "Error sending friends-request (%s).\n", requestError)
			} else {
				fmt.Fprintln(writer, "Friends-request has been sent.")
			}
		} else {
			fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.String())
			}
		}
	}
}

// Name returns the name of the command.
func (f *Friends) Name() string {
	return "friends"
}

// PrintHelp prints the general help page for the friends commands.
func (f *Friends) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, friendsDocumentation)
}
