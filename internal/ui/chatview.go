package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

type ChatView struct {
	internalTextView *tview.TextView

	session   *discordgo.Session
	data      []*discordgo.Message
	ownUserID string
}

func NewChatView(session *discordgo.Session, ownUserID string) *ChatView {
	chatView := ChatView{
		internalTextView: tview.NewTextView(),
		session:          session,
		ownUserID:        ownUserID,
	}

	chatView.internalTextView.SetDynamicColors(true)
	chatView.internalTextView.SetRegions(true)
	chatView.internalTextView.SetWordWrap(true)
	chatView.internalTextView.SetBorder(true)

	return &chatView
}

func (chatView *ChatView) GetPrimitive() tview.Primitive {
	return chatView.internalTextView
}

func (chatView *ChatView) SetMessages(messages []*discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()
	chatView.data = messages

	newText := ""

	for _, message := range messages {

		time, parseError := message.Timestamp.Parse()
		var timeCellText string
		if parseError == nil {
			time := time.Local()
			conf := config.GetConfig()
			if conf.Times == config.HourMinuteAndSeconds {
				timeCellText = fmt.Sprintf("%02d:%02d:%02d ", time.Hour(), time.Minute(), time.Second())
			} else if conf.Times == config.HourAndMinute {
				timeCellText = fmt.Sprintf("%02d:%02d ", time.Hour(), time.Minute())
			}
		}

		mentionsUser := false
		for _, mentionedUser := range message.Mentions {
			if mentionedUser.ID == chatView.ownUserID {
				mentionsUser = true
				break
			}
		}

		messageText := message.Content
		for _, user := range message.Mentions {
			messageText = strings.NewReplacer(
				"<@"+user.ID+">", "[blue]@"+user.Username+"[white]",
				"<@!"+user.ID+">", "[blue]@"+user.Username+"[white]",
			).Replace(messageText)
		}

		messageText = regexp.
			MustCompile("<#\\d*>").
			ReplaceAllStringFunc(messageText, func(data string) string {
				channelID := strings.TrimSuffix(strings.TrimPrefix(data, "<#"), ">")
				channel, cacheError := chatView.session.State.Channel(channelID)
				if cacheError != nil {
					return data
				}

				return "[blue]#" + channel.Name + "[white]"
			})

		//TODO Role mentions

		if message.Attachments != nil && len(message.Attachments) != 0 {
			if messageText != "" {
				messageText = messageText + " "
			}
			messageText = messageText + message.Attachments[0].URL
		}

		messageText = fmt.Sprintf("[\"%s\"][#FF0000]%s[#00FF00]%s [white]%s[\"\"]", message.ID, timeCellText, message.Author.Username, messageText)
		newText = fmt.Sprintf("%s\n%s", newText, messageText)

		if mentionsUser {
			chatView.internalTextView.Highlight(append(chatView.internalTextView.GetHighlights(), message.ID)...)
		}
	}

	chatView.internalTextView.SetText("")
	fmt.Fprint(chatView.internalTextView, newText)

	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}
