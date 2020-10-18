package discordutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/times"
	"github.com/Bios-Marcel/cordless/util/files"
)

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
	// ChannelMessages fetches up to 100 messages for a channel.
	// The parameter beforeID defines whether message only older than
	// a specific message should be returned. The parameter afterID does
	// the same but for newer messages. The parameter aroundID is a mix of
	// both.
	ChannelMessages(channelID string, limit int, beforeID string, afterID string, aroundID string) ([]*discordgo.Message, error)
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

// CreateMessageLoader creates a MessageLoader using the given
// MessageDataSupplier. It is empty and can be used right away.
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
	//Empty channels are never marked as cached and needn't be loaded.
	if channel.LastMessageID == "" {
		return nil, nil
	}

	//If it's already cached, we assume that it contains all existing messages.
	if l.IsCached(channel.ID) {
		return channel.Messages, nil
	}

	var beforeID string
	localMessageCount := len(channel.Messages)
	if localMessageCount > 0 {
		beforeID = channel.Messages[0].ID
	}

	//We might not have all messages, as we might have received message due to
	//update events, which doesn't include the previously sent messages. This
	//however only matters if we haven't already reached 100 or more messages
	//via update events.
	messagesToGet := 100 - localMessageCount
	if messagesToGet > 0 {
		messages, discordError := l.messageDateSupplier.ChannelMessages(channel.ID, messagesToGet, beforeID, "", "")
		if discordError != nil {
			return nil, discordError
		}

		//Workaround for a bug where messages were lacking the GuildID.
		if channel.GuildID != "" {
			for _, message := range messages {
				message.GuildID = channel.GuildID
			}
		}

		if localMessageCount == 0 {
			channel.Messages = messages
		} else {
			//There are already messages in cache; However, those came from
			//updates events, meaning those have to be newer than the
			//requested ones.
			channel.Messages = append(messages, channel.Messages...)
		}
	}

	l.requestedChannels[channel.ID] = true

	return channel.Messages, nil
}

// SendMessageAsFile sends the given message into the given channel using the
// passed discord Session. If an error occurs, onFailure gets called.
func SendMessageAsFile(session *discordgo.Session, message string, channel string, onFailure func(error)) {
	reader := bytes.NewBufferString(message)
	messageAsFile := &discordgo.File{
		Name:        "message.txt",
		ContentType: "text",
		Reader:      reader,
	}
	complexMessage := &discordgo.MessageSend{
		Content: "The message was too long, therefore, you get a file:",
		Embed:   nil,
		TTS:     false,
		Files:   nil,
		File:    messageAsFile,
	}
	_, sendError := session.ChannelMessageSendComplex(channel, complexMessage)
	if sendError != nil {
		onFailure(sendError)
	}
}

// GenerateQuote formats a message quote using the given Input. The
// `messageAfterQuote` will be appended after the quote in case it is not
// empty.
func GenerateQuote(message, author string, time discordgo.Timestamp, attachments []*discordgo.MessageAttachment, messageAfterQuote string) (string, error) {
	messageTime, parseError := time.Parse()
	if parseError != nil {
		return "", parseError
	}

	// All quotes should be UTC in order to not confuse quote-readers.
	messageTimeUTC := messageTime.UTC()
	quotedMessage := strings.ReplaceAll(message, "\n", "\n> ")

	if len(attachments) > 0 {
		var attachmentsAsText string
		for index, attachment := range attachments {
			if index == 0 {
				attachmentsAsText += attachment.URL
			} else {
				attachmentsAsText += "\n> " + attachment.URL
			}
		}

		//If the quoted message ends with a "useless" quote-line-prefix
		//we simply "reuse" that line to not add unnecessary newlines.
		if strings.HasSuffix(quotedMessage, "> ") {
			quotedMessage = quotedMessage + attachmentsAsText
		} else {
			quotedMessage = quotedMessage + "\n> " + attachmentsAsText
		}
	}

	return fmt.Sprintf("> **%s** %s UTC:\n> %s\n%s", author,
			times.TimeToString(&messageTimeUTC), quotedMessage,
			strings.TrimSpace(messageAfterQuote)),
		nil
}

// MessageToPlainText converts a discord message to a human readable text.
// Markdown characters are reserved and file attachments are added as URLs.
// Embeds are currently not being handled, nor are other special elements.
func MessageToPlainText(message *discordgo.Message) string {
	content := message.ContentWithMentionsReplaced()
	builder := &strings.Builder{}

	if content != "" {
		builder.Grow(len(content))
		builder.WriteString(content)
	}

	if len(message.Attachments) > 0 {
		builder.Grow(1)
		builder.WriteRune('\n')

		if len(message.Attachments) == 1 {
			builder.Grow(len(message.Attachments[0].URL))
			builder.WriteString(message.Attachments[0].URL)
		} else if len(message.Attachments) > 1 {
			links := make([]string, 0, len(message.Attachments))
			for _, file := range message.Attachments {
				links = append(links, file.URL)
			}

			linksAsText := strings.Join(links, "\n")
			builder.Grow(len(linksAsText))
			builder.WriteString(linksAsText)
		}
	}

	return builder.String()
}

// ResolveFilePathAndSendFile will attempt to resolve the message and see if
// it points to a file on the users harddrive. If so, it's sent to the given
// channel using it's basename as the discord filename.
func ResolveFilePathAndSendFile(session *discordgo.Session, message, targetChannelID string) error {
	path, pathError := files.ToAbsolutePath(message)
	if pathError != nil {
		return pathError
	}
	data, readError := ioutil.ReadFile(path)
	if readError != nil {
		return readError
	}
	reader := bytes.NewBuffer(data)
	_, sendError := session.ChannelFileSend(targetChannelID, filepath.Base(message), reader)
	return sendError
}
