package ui

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/Bios-Marcel/cordless/internal/config"
	//Blank import for initializing the tview formatter
	_ "github.com/Bios-Marcel/cordless/internal/syntax"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var (
	codeBlockRegex      = regexp.MustCompile("(?s)\x60\x60\x60(.+?)\n(.+?)\x60\x60\x60(?:$|\n)")
	channelMentionRegex = regexp.MustCompile("<#\\d*>")
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

		var messageText string

		if message.Type == discordgo.MessageTypeDefault {
			messageText = message.Content
			for _, user := range message.Mentions {
				messageText = strings.NewReplacer(
					"<@"+user.ID+">", "[blue]@"+user.Username+"[white]",
					"<@!"+user.ID+">", "[blue]@"+user.Username+"[white]",
				).Replace(messageText)
			}
		} else if message.Type == discordgo.MessageTypeGuildMemberJoin {
			messageText = "[gray]joined the server."
		} else if message.Type == discordgo.MessageTypeCall {
			messageText = "[gray]is calling you."
		} else if message.Type == discordgo.MessageTypeChannelIconChange {
			messageText = "[gray]changed the channel icon."
		} else if message.Type == discordgo.MessageTypeChannelNameChange {
			messageText = "[gray]changed the channel name."
		} else if message.Type == discordgo.MessageTypeChannelPinnedMessage {
			messageText = "[gray]pinned a message."
		} else if message.Type == discordgo.MessageTypeRecipientAdd {
			messageText = "[gray]added " + message.Mentions[0].Username + " to the group."
		} else if message.Type == discordgo.MessageTypeRecipientRemove {
			messageText = "[gray]removed " + message.Mentions[0].Username + " from the group."
		}

		messageText = channelMentionRegex.
			ReplaceAllStringFunc(messageText, func(data string) string {
				channelID := strings.TrimSuffix(strings.TrimPrefix(data, "<#"), ">")
				channel, cacheError := chatView.session.State.Channel(channelID)
				if cacheError != nil {
					return data
				}

				return "[blue]#" + channel.Name + "[white]"
			})

		groupValues := codeBlockRegex.
			//Magicnumber, cuz u ain't gonna such a long message anyway.
			FindAllStringSubmatch(messageText, 1000)

		for _, values := range groupValues {
			language := ""
			for index, value := range values {
				if index == 0 {
					continue
				}

				if index == 1 {
					language = value
				} else if index == 2 {
					// Determine lexer.
					l := lexers.Get(language)
					if l == nil {
						l = lexers.Analyse(value)
					}
					if l == nil {
						l = lexers.Fallback
					}
					l = chroma.Coalesce(l)

					// Determine formatter.
					f := formatters.Get("tview-8bit")
					if f == nil {
						f = formatters.Fallback
					}

					// Determine style.
					s := styles.Get("monokai")
					if s == nil {
						s = styles.Fallback
					}

					it, tokeniseError := l.Tokenise(nil, value)
					if tokeniseError != nil {
						continue
					}

					writer := bytes.NewBufferString("")

					formatError := f.Format(writer, s, it)
					if formatError != nil {
						continue
					}

					messageText = strings.Replace(messageText, value, writer.String(), 1)
				}
			}
		}

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
