package discordgoplus

import (
	"fmt"

	"github.com/Bios-Marcel/discordgo"
)

// GetPrivateChannelName generates a name for a private channel.
func GetPrivateChannelName(channel *discordgo.Channel) string {
	var channelName string
	if channel.Type == discordgo.ChannelTypeDM {
		channelName = channel.Recipients[0].Username
	} else if channel.Type == discordgo.ChannelTypeGroupDM {
		if channel.Name != "" {
			channelName = channel.Name
		} else {
			for index, recipient := range channel.Recipients {
				if index == 0 {
					channelName = recipient.Username
				} else {
					channelName = fmt.Sprintf("%s, %s", channelName, recipient.Username)
				}
			}
		}
	}

	if channelName == "" {
		channelName = "Unnamed"
	}

	return channelName
}
