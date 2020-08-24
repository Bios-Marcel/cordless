package commandimpls

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/Bios-Marcel/discordgo"
)

const friendsDocumentation = `[orange][::u]# friends[white]

The friends command allows you to manage your friends on discord. You can add
new friends by sending or accepting friend-requests. You can also see your
current requests, that goes for the incoming and the outgoing ones.

The friend command currently offers the following subcommands:
  * accept   - accept a friend-request
  * befriend - send a friend-request
  * requests - shows all current requests
  * search   - finds friends by name, name#discriminator or id
  * list     - shows all friends
  * remove   - removes a friend from your friend-list

The following features are currently unsupported:
  * Blocking users
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

	if f.session.State.User.Bot {
		fmt.Fprintln(writer, "[red]This command can't be used by bots due to Discord API restrictions.")
		return
	}

	switch parameters[0] {
	case "list", "show", "which":
		fmt.Fprintln(writer, "Friends:")
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeFriend {
				fmt.Fprintln(writer, "  "+rel.User.Username)
			}
		}
	case "delete", "unfriend", "remove", "decline":
		if len(parameters) != 2 {
			fmt.Fprintln(writer, "Usage: friends remove <Username|Username#NNNN|UserID>")
			return
		}

		input := parameters[1]
		var matches []*discordgo.Relationship
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeFriend ||
				rel.Type == discordgo.RelationTypeOutgoingRequest ||
				rel.Type == discordgo.RelationTypeIncommingRequest {
				if rel.User.ID == input || rel.User.Username == input || rel.User.String() == input {
					matches = append(matches, rel)
				}
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(writer, "No matches for '%s' found.\n", input)
		} else if len(matches) == 1 {
			user := matches[0]
			fmt.Fprintln(writer, "Removing friend "+user.User.String())
			acceptErr := f.session.RelationshipDelete(user.User.ID)
			if acceptErr != nil {
				fmt.Fprintf(writer, "Error removing friend (%s).\n", acceptErr.Error())
			} else {
				fmt.Fprintln(writer, user.User.String()+" has been removed as your friend.")
			}
		} else {
			fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.User.String())
			}
		}
	case "requests", "invites", "outstanding", "unanswered":
		var incoming, outgoing string
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeIncommingRequest {
				incoming += "  " + rel.User.String() + "\n"
			} else if rel.Type == discordgo.RelationTypeOutgoingRequest {
				outgoing += "  " + rel.User.String() + "\n"
			}
		}

		fmt.Fprintln(writer, "Incoming requests:")
		if incoming != "" {
			fmt.Fprintln(writer, incoming)
		} else {
			fmt.Fprintln(writer, "No incoming requests.")
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

		input := parameters[1]
		var matches []*discordgo.Relationship
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeIncommingRequest {
				if rel.User.ID == input || rel.User.Username == input || rel.User.String() == input {
					matches = append(matches, rel)
				}
			}
		}
		if len(matches) == 0 {
			fmt.Fprintf(writer, "No matches for '%s' found.\n", input)
		} else if len(matches) == 1 {
			fmt.Fprintln(writer, "Accepting friend request of "+matches[0].User.String())
			acceptErr := f.session.RelationshipFriendRequestAccept(matches[0].User.ID)
			if acceptErr != nil {
				fmt.Fprintf(writer, "Error accepting friend-request (%s).\n", acceptErr.Error())
			} else {
				fmt.Fprintln(writer, matches[0].User.String()+" is now your friend.")
			}
		} else {
			fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.User.String())
			}
		}
	case "search", "find":
		if len(parameters) != 2 {
			fmt.Fprintln(writer, "Usage: friends find <Username|Username#NNNN|UserID")
			return
		}

		input := parameters[1]

		var matches []*discordgo.Relationship
		for _, rel := range f.session.State.Relationships {
			if rel.Type == discordgo.RelationTypeFriend {
				if strings.Contains(rel.User.ID, input) ||
					strings.Contains(rel.User.Username, input) ||
					strings.Contains(rel.User.String(), input) {
					matches = append(matches, rel)
				}
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(writer, "No matches were found for '%s'.\n", input)
		} else {
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.User.String())
			}
		}

	case "befriend", "add", "send", "ask", "invite", "request":
		if len(parameters) != 2 {
			fmt.Fprintln(writer, "Usage: friends befriend <Username|Username#NNNN|UserID")
			return
		}

		//Iterate over all available users and find one that fits, if we were
		//successful, we send a friendsrequest. Otherwise we check if the input
		//might have been a user idea, lookup the user and do a request.
		input := parameters[1]

		users, err := f.session.State.Users()
		if err != nil {
			fmt.Fprintf(writer, "An error occured during commandexecution (%s).\n", err.Error())
			return
		}

		var matches []*discordgo.User
		for _, user := range users {
			if user.ID == input || user.Username == input || user.String() == input {
				matches = append(matches, user)
			}
		}

		if len(matches) == 0 {
			//Send friendsrequest via username and discriminator if possible
			parts := strings.Split(input, "#")
			if len(parts) == 2 {
				discriminator, _ := strconv.ParseInt(parts[1], 10, 32)
				requestError := f.session.RelationshipFriendRequestSendByNameAndDiscriminator(parts[0], int(discriminator))
				if requestError != nil {
					fmt.Fprintf(writer, "Error sending friend-request to '%s'.\n\t%s\n", input, requestError.Error())
					return
				}

				fmt.Fprintf(writer, "A friend-request has been sent to '%s'.\n", input)
				return
			}

			//If no match was found, try sending a friendsrequest if the input is a snowflake.
			for _, char := range input {
				if !unicode.IsNumber(char) {
					fmt.Fprintf(writer, "No matches for '%s' found. Please ask that person to add you or find out the UserID.\n", input)
					return
				}
			}

			requestError := f.session.RelationshipFriendRequestSend(input)
			if requestError != nil {
				fmt.Fprintf(writer, "Error sending friends-request (%s).\n", requestError)
			} else {
				fmt.Fprintln(writer, "Friend-request has been sent.")
			}
		} else if len(matches) == 1 {
			user := matches[0]

			requestError := f.session.RelationshipFriendRequestSend(user.ID)
			if requestError != nil {
				fmt.Fprintf(writer, "Error sending friend-request (%s).\n", requestError)
			} else {
				fmt.Fprintf(writer, "A friend-request has been sent to '%s'.\n", user.String())
			}
		} else {
			fmt.Fprintf(writer, "Multiple matches were found for '%s'. Please be more precise.\n", input)
			fmt.Fprintln(writer, "The following matches were found:")
			for _, match := range matches {
				fmt.Fprintln(writer, "  "+match.String())
			}
		}
	default:
		f.PrintHelp(writer)
	}
}

// Name returns the name of the command.
func (f *Friends) Name() string {
	return "friends"
}

// Aliases returns availabe aliases for this command
func (f *Friends) Aliases() []string {
	return []string{"friend"}
}

// PrintHelp prints the general help page for the friends commands.
func (f *Friends) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, friendsDocumentation)
}
