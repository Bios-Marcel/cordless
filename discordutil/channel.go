package discordutil

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

// SortMessagesByTimestamp sorts all messages in the given array according to
// their creation date.
func SortMessagesByTimestamp(messages []*discordgo.Message) {
	sort.Slice(messages, func(a, b int) bool {
		timeA, parseError := messages[a].Timestamp.Parse()
		if parseError != nil {
			return false
		}

		timeB, parseError := messages[b].Timestamp.Parse()
		if parseError != nil {
			return true
		}

		return timeA.Before(timeB)
	})
}

// GetPrivateChannelNameUnescaped generates a name for a private channel.
// The name won't be escaped view tviewutil and therefore shouldn't be used
// for displaying it in tview components.
func GetPrivateChannelNameUnescaped(channel *discordgo.Channel) string {
	var channelName string
	if channel.Type == discordgo.ChannelTypeDM {
		//The first recipient should always be us!
		channelName = channel.Recipients[0].Username
		//Since the official client doesn't seem to allow creating nicks for
		//simple DMs, we assume it isn't possible.
	} else if channel.Type == discordgo.ChannelTypeGroupDM {
		//Groups can have custom names.
		if channel.Name != "" {
			channelName = channel.Name
		} else {
			//Channels can have nicknames, but if they don't the default discord
			//client just displays the recipients names sticked together.
			for index, recipient := range channel.Recipients {
				if index == 0 {
					channelName = recipient.Username
				} else {
					channelName = fmt.Sprintf("%s, %s", channelName, recipient.Username)
				}
			}
		}
	}

	//This is a fallback, so we don't have an empty string.
	//This happens sometimes, I am unsure when though.
	if channelName == "" {
		channelName = "Unnamed"
	}

	return channelName
}

// GetPrivateChannelName generates a name for a private channel.
func GetPrivateChannelName(channel *discordgo.Channel) string {
	return tviewutil.Escape(GetPrivateChannelNameUnescaped(channel))
}

// FindDMChannelWithUser tries to find a DM channel with the specified user as
// one of its two recipients. If no channel is found, nil is returned.
func FindDMChannelWithUser(state *discordgo.State, userID string) *discordgo.Channel {
	for _, privateChannel := range state.PrivateChannels {
		if privateChannel.Type == discordgo.ChannelTypeDM {
			for _, recipient := range privateChannel.Recipients {
				if recipient.ID == userID {
					return privateChannel
				}
			}
		}
	}

	return nil
}

// CompareChannels checks which channel is smaller. Smaller meaning it is the
// one with the more recent message.
func CompareChannels(a, b *discordgo.Channel) bool {
	messageA, parseError := strconv.ParseInt(a.LastMessageID, 10, 64)
	if parseError != nil {
		return false
	}

	messageB, parseError := strconv.ParseInt(b.LastMessageID, 10, 64)
	if parseError != nil {
		return true
	}

	return messageA > messageB
}

// SortPrivateChannels sorts private channels depending on their last message.
func SortPrivateChannels(channels []*discordgo.Channel) {
	sort.Slice(channels, func(a, b int) bool {
		return CompareChannels(channels[a], channels[b])
	})
}

// HasReadMessagesPermission checks if the user has permission to view a
// specific channel.
func HasReadMessagesPermission(channelID string, state *discordgo.State) bool {
	userPermissions, err := state.UserChannelPermissions(state.User.ID, channelID)
	if err != nil {
		// Unable to access channel permissions.
		return false
	}
	return (userPermissions & discordgo.PermissionViewChannel) > 0
}

// AcknowledgeChannel acknowledges all messages in the given channel. If the
// channel is a category, all children will be acknowledged.
func AcknowledgeChannel(session *discordgo.Session, channelID string) error {
	channel, stateError := session.State.Channel(channelID)
	if stateError != nil {
		return stateError
	}

	//Bulk-Acknowledge of categories
	if channel.Type == discordgo.ChannelTypeGuildCategory {
		guild, stateError := session.State.Guild(channel.GuildID)
		if stateError != nil {
			return stateError
		}

		var channelsToAck []*discordgo.Channel
		for _, guildChannel := range guild.Channels {
			if guildChannel.ParentID != channel.ID {
				continue
			}

			//These can't have messages. Store is dead anyways, so we needn't handle it.
			if guildChannel.Type == discordgo.ChannelTypeGuildVoice {
				continue
			}

			if guildChannel.LastMessageID == "" || readstate.HasBeenRead(guildChannel, guildChannel.LastMessageID) {
				continue
			}

			channelsToAck = append(channelsToAck, guildChannel)
		}

		if len(channelsToAck) > 0 {
			ackError := session.BulkChannelMessageAck(channelsToAck)
			return ackError
		}
	} else {
		if channel.LastMessageID != "" && !readstate.HasBeenRead(channel, channel.LastMessageID) {
			_, ackError := session.ChannelMessageAck(channel.ID, channel.LastMessageID, "")
			return ackError
		}
	}

	return nil
}
