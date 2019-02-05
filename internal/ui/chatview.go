package ui

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Bios-Marcel/cordless/internal/config"
	// Blank import for initializing the tview formatter
	_ "github.com/Bios-Marcel/cordless/internal/syntax"
	"github.com/Bios-Marcel/discordgo"
	linkshortener "github.com/Bios-Marcel/shortnotforlong"
	"github.com/Bios-Marcel/tview"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var (
	codeBlockRegex      = regexp.MustCompile("(?s)\x60\x60\x60(.+?)\n(.+?)\x60\x60\x60(?:$|\n)")
	channelMentionRegex = regexp.MustCompile("<#\\d*>")
	urlRegex            = regexp.MustCompile("https?://(.+?)(?:(?:/.*)|\\s|$)")
	shortener           = linkshortener.NewShortener(51726)
)

// ChatView is using a tview.TextView in order to be able to display messages
// in a simple way. It supports highlighting specific element types and it
// also supports multiline.
type ChatView struct {
	internalTextView *tview.TextView

	session   *discordgo.Session
	data      []*discordgo.Message
	ownUserID string
}

// NewChatView constructs a new ready to use ChatView.
func NewChatView(session *discordgo.Session, ownUserID string) *ChatView {
	chatView := ChatView{
		internalTextView: tview.NewTextView(),
		session:          session,
		ownUserID:        ownUserID,
	}

	if config.GetConfig().ShortenLinks {
		go func() {
			shortenerError := shortener.Start()
			if shortenerError != nil {
				log.Fatalln("Error creating shortener:", shortenerError.Error())
			}
		}()
	}

	chatView.internalTextView.SetDynamicColors(true)
	chatView.internalTextView.SetRegions(true)
	chatView.internalTextView.SetWordWrap(true)
	chatView.internalTextView.SetBorder(true)

	return &chatView
}

// GetPrimitive returns the component that can be added to a layout, since
// the ChatView itself is not a component.
func (chatView *ChatView) GetPrimitive() tview.Primitive {
	return chatView.internalTextView
}

//AddMessages adds additional messages to the ChatView.
func (chatView *ChatView) AddMessages(messages []*discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()
	chatView.data = append(chatView.data, messages...)

	newText := ""

	conf := config.GetConfig()

	for _, message := range messages {

		time, parseError := message.Timestamp.Parse()
		var timeCellText string
		if parseError == nil {
			time := time.Local()
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
			messageText = tview.Escape(message.Content)
			for _, user := range message.Mentions {
				escapedUsername := tview.Escape(user.Username)
				messageText = strings.NewReplacer(
					"<@"+user.ID+">", "[blue]@"+escapedUsername+"[white]",
					"<@!"+user.ID+">", "[blue]@"+escapedUsername+"[white]",
				).Replace(messageText)
			}
		} else if message.Type == discordgo.MessageTypeGuildMemberJoin {
			messageText = "[gray]joined the server."
		} else if message.Type == discordgo.MessageTypeCall {
			messageText = "[gray]is calling you."
		} else if message.Type == discordgo.MessageTypeChannelIconChange {
			messageText = "[gray]changed the channel icon."
		} else if message.Type == discordgo.MessageTypeChannelNameChange {
			messageText = "[gray]changed the channel name to " + message.Content + "."
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

		if message.Attachments != nil && len(message.Attachments) > 0 {
			attachmentsAsText := ""
			for attachmentIndex, attachment := range message.Attachments {
				attachmentsAsText = attachmentsAsText + attachment.URL
				if attachmentIndex != len(message.Attachments)-1 {
					attachmentsAsText = attachmentsAsText + "\n"
				}
			}

			if messageText != "" {
				messageText = attachmentsAsText + "\n" + messageText
			} else {
				messageText = attachmentsAsText + messageText
			}
		}

		if conf.ShortenLinks {
			urlMatches := urlRegex.FindAllStringSubmatch(messageText, 1000)

			for _, urlMatch := range urlMatches {
				if (len(urlMatch[1]) + 35) < len(urlMatch[0]) {
					shortenedURL := fmt.Sprintf("(%s) %s", urlMatch[1], shortener.Shorten(urlMatch[0]))
					messageText = strings.Replace(messageText, urlMatch[0], shortenedURL, 1)
				}
			}
		}

		groupValues := codeBlockRegex.
			// Magicnumber, because message aren't gonna be that long anyway.
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

		messageText = fmt.Sprintf("[\"%s\"][#FF0000]%s[#00FF00]%s [white]%s[\"\"]", message.ID, timeCellText, message.Author.Username, messageText)
		newText = fmt.Sprintf("%s\n%s", newText, messageText)

		if mentionsUser {
			chatView.internalTextView.Highlight(append(chatView.internalTextView.GetHighlights(), message.ID)...)
		}
	}

	fmt.Fprint(chatView.internalTextView, newText)

	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

// SetMessages defines all currently displayed messages. Parsing and
// manipulation of single message elements happens in this function.
func (chatView *ChatView) SetMessages(messages []*discordgo.Message) {
	chatView.data = make([]*discordgo.Message, 0)
	chatView.internalTextView.SetText("")

	chatView.AddMessages(messages)
}
