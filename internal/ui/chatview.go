package ui

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

type ChatView struct {
	internalTextView *tview.TextView

	data []*discordgo.Message
}

func NewChatView() *ChatView {
	chatView := ChatView{
		internalTextView: tview.NewTextView(),
	}

	chatView.internalTextView.SetDynamicColors(true)
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

		messageText := message.ContentWithMentionsReplaced()
		if message.Attachments != nil && len(message.Attachments) != 0 {
			if messageText != "" {
				messageText = messageText + " "
			}
			messageText = messageText + message.Attachments[0].URL
		}

		newText = fmt.Sprintf("%s[#FF0000]%s[#00FF00]%s [white]%s\n", newText, timeCellText, message.Author.Username, messageText)
	}

	chatView.internalTextView.SetText("")
	fmt.Fprintf(chatView.internalTextView, "%s", newText)

	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}
