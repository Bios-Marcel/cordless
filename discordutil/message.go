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

// MessageDataSupplier defines the method that is necessary for requesting
// channels. This is satisfied by the discordgo.Session struct and can be
// used in order to make testing easier.
type MessageDataSupplier interface {
	ChannelMessages(string, int, string, string, string) ([]*discordgo.Message, error)
}

// MessageLoader represents a util object that remember which channels have
// already been cached and which not.
type MessageLoader struct {
	messageDateSupplier MessageDataSupplier
	requestedChannels   map[string]bool
}

// IsCached checks whether the channel has already been requested from the
// backend once.
func (l *MessageLoader) IsCached(channelID string) bool {
	value, cached := l.requestedChannels[channelID]
	return cached && value
}

func CreateMessageLoader(messageDataSupplier MessageDataSupplier) *MessageLoader {
	loader := &MessageLoader{
		requestedChannels:   make(map[string]bool),
		messageDateSupplier: messageDataSupplier,
	}

	return loader
}

// DeleteFromCache deletes the entry that indicates the channel has been
// cached. The next call to LoadMessages with the same ID will ask for data
// from the MessageDataSupplier.
func (l *MessageLoader) DeleteFromCache(channelID string) {
	delete(l.requestedChannels, channelID)
}

// LoadMessages returns the last 100 messages for a channel. If less messages
// were sent, less will be returned. As soon as a channel has been loaded once
// it won't ever be loaded again, instead a global cache will be accessed.
func (l *MessageLoader) LoadMessages(channel *discordgo.Channel) ([]*discordgo.Message, error) {
	var messages []*discordgo.Message

	if channel.LastMessageID != "" {
		if !l.IsCached(channel.ID) {
			l.requestedChannels[channel.ID] = true

			var beforeID string
			localMessageCount := len(channel.Messages)
			if localMessageCount > 0 {
				beforeID = channel.Messages[0].ID
			}

			messagesToGet := 100 - localMessageCount
			if messagesToGet > 0 {
				var discordError error
				messages, discordError = l.messageDateSupplier.ChannelMessages(channel.ID, messagesToGet, beforeID, "", "")
				if discordError != nil {
					return nil, discordError
				}

				if channel.GuildID != "" {
					for _, message := range messages {
						message.GuildID = channel.GuildID
					}
				}
				if localMessageCount == 0 {
					channel.Messages = messages
				} else {
					//There are already messages in cache; However, those came from updates events.
					//Therefore those have to be newer than the newly retrieved ones.
					channel.Messages = append(messages, channel.Messages...)
				}
			}
		}
		messages = channel.Messages
	}

	return messages, nil
}
