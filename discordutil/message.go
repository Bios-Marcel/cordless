package discordutil

import "github.com/Bios-Marcel/discordgo"

// MentionsCurrentUserExplicitly checks whether the message contains any
// explicit mentions for the user associated with the currently logged in user.
func MentionsCurrentUserExplicitly(state *discordgo.State, message *discordgo.Message) bool {
	for _, user := range message.Mentions {
		if user.ID == state.User.ID {
			return true
		}
	}

	return false
}
