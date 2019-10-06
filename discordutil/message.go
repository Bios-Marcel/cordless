package discordutil

import "github.com/Bios-Marcel/discordgo"

func MentionsCurrentUserExplicitly(state *discordgo.State, message *discordgo.Message) bool {
	for _, user := range message.Mentions {
		if user.ID == state.User.ID {
				return true
			break
		}
	}

	return false
}
