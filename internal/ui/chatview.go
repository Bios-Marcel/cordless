package ui

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	linkshortener "github.com/Bios-Marcel/shortnotforlong"

	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/Bios-Marcel/cordless/internal/times"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/internal/config"
	// Blank import for initializing the tview formatter
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
	urlRegex            = regexp.MustCompile("<?(https?://)(.+?)(/.+?)?($|\\s|\\||>)")
	spoilerRegex        = regexp.MustCompile("(?s)\\|\\|(.+?)\\|\\|")

	userColor = "green"
)

// ChatView is using a tview.TextView in order to be able to display messages
// in a simple way. It supports highlighting specific element types and it
// also supports multiline.
type ChatView struct {
	internalTextView *tview.TextView

	shortener *linkshortener.Shortener

	session   *discordgo.Session
	data      []*discordgo.Message
	ownUserID string

	selection     int
	selectionMode bool

	showSpoilerContent map[string]bool

	onMessageAction func(message *discordgo.Message, event *tcell.EventKey) *tcell.EventKey
}

// NewChatView constructs a new ready to use ChatView.
func NewChatView(session *discordgo.Session, ownUserID string) *ChatView {
	chatView := ChatView{
		internalTextView:   tview.NewTextView(),
		session:            session,
		ownUserID:          ownUserID,
		selection:          -1,
		selectionMode:      false,
		showSpoilerContent: make(map[string]bool, 0),
	}

	if config.GetConfig().ShortenLinks {
		chatView.shortener = linkshortener.NewShortener(config.GetConfig().ShortenerPort)
		go func() {
			shortenerError := chatView.shortener.Start()
			if shortenerError != nil {
				log.Fatalln("Error creating shortener:", shortenerError.Error())
			}
		}()
	}

	chatView.internalTextView.SetOnBlur(func() {
		chatView.selectionMode = false
		chatView.ClearSelection()
	})

	chatView.internalTextView.SetOnFocus(func() {
		chatView.selectionMode = true
	})

	chatView.internalTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if chatView.selectionMode && event.Modifiers() == tcell.ModNone {
			if event.Key() == tcell.KeyUp {
				if chatView.selection == -1 {
					chatView.selection = len(chatView.data) - 1
				} else {
					if chatView.selection >= 1 {
						chatView.selection--
					}
				}

				chatView.updateHighlights()

				return nil
			}

			if event.Key() == tcell.KeyDown {
				if chatView.selection == -1 {
					chatView.selection = 0
				} else {
					if chatView.selection <= len(chatView.data)-2 {
						chatView.selection++
					}
				}

				chatView.updateHighlights()

				return nil
			}

			if event.Key() == tcell.KeyHome {
				chatView.selection = 0

				chatView.updateHighlights()

				return nil
			}

			if event.Key() == tcell.KeyEnd {
				chatView.selection = len(chatView.data) - 1

				chatView.updateHighlights()

				return nil
			}

			if chatView.selection > 0 && chatView.selection < len(chatView.data) && event.Rune() == 's' {
				messageID := chatView.data[chatView.selection].ID
				currentValue, contains := chatView.showSpoilerContent[messageID]
				if contains {
					chatView.showSpoilerContent[messageID] = !currentValue
				} else {
					chatView.showSpoilerContent[messageID] = true
				}
				chatView.SetMessages(chatView.data)
				return nil
			}

			if chatView.selection > 0 && chatView.selection < len(chatView.data) && chatView.onMessageAction != nil {
				return chatView.onMessageAction(chatView.data[chatView.selection], event)
			}
		}

		return event
	})

	chatView.internalTextView.SetDynamicColors(true)
	chatView.internalTextView.SetRegions(true)
	chatView.internalTextView.SetWordWrap(true)
	chatView.internalTextView.SetBorder(true)

	return &chatView
}

// SetOnMessageAction sets the handler that will get called if the user tries
// to interact with a selected message.
func (chatView *ChatView) SetOnMessageAction(onMessageAction func(message *discordgo.Message, event *tcell.EventKey) *tcell.EventKey) {
	chatView.onMessageAction = onMessageAction
}

func intToString(value int) string {
	return strconv.FormatInt(int64(value), 10)
}

func (chatView *ChatView) updateHighlights() {
	chatView.internalTextView.Highlight(intToString(chatView.selection))
	chatView.internalTextView.ScrollToHighlight()
}

// GetPrimitive returns the component that can be added to a layout, since
// the ChatView itself is not a component.
func (chatView *ChatView) GetPrimitive() tview.Primitive {
	return chatView.internalTextView
}

//AddMessages adds additional messages to the ChatView.
func (chatView *ChatView) AddMessages(messages []*discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()
	nextIndex := len(chatView.data)
	chatView.data = append(chatView.data, messages...)

	newText := ""

	conf := config.GetConfig()

	for _, message := range messages {

		time, parseError := message.Timestamp.Parse()
		var timeCellText string
		if parseError == nil {
			timeCellText = times.TimeToString(&time)
		}

		var messageText string

		if message.Type == discordgo.MessageTypeDefault {
			messageText = tview.Escape(message.Content)

			for _, roleID := range message.MentionRoles {
				role, cacheError := chatView.session.State.Role(message.GuildID, roleID)
				if cacheError == nil {
					messageText = strings.NewReplacer(
						"<@&"+roleID+">", "[blue]@"+role.Name+"[white]",
					).Replace(messageText)
				}
			}

			for _, user := range message.Mentions {
				var userName string
				if message.GuildID != "" {
					member, cacheError := chatView.session.State.Member(message.GuildID, user.ID)
					if cacheError == nil {
						userName = discordgoplus.GetMemberName(member, nil)
					}
				}

				if userName == "" {
					userName = discordgoplus.GetUserName(user, nil)
				}

				messageText = strings.NewReplacer(
					"<@"+chatView.session.State.User.ID+">", "[orange]@"+userName+"[white]",
					"<@!"+chatView.session.State.User.ID+">", "[orange]@"+userName+"[white]",
				).Replace(messageText)

				messageText = strings.NewReplacer(
					"<@"+user.ID+">", "[blue]@"+userName+"[white]",
					"<@!"+user.ID+">", "[blue]@"+userName+"[white]",
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

		// FIXME Needs improvement, as it wastes space and breaks things
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

		// FIXME Handle Non-embed links nonetheless?
		if conf.ShortenLinks {
			urlMatches := urlRegex.FindAllStringSubmatch(messageText, 1000)

			for _, urlMatch := range urlMatches {
				newURL := urlMatch[1] + urlMatch[2]
				if len(urlMatch) == 5 || (len(urlMatch) == 4 && len(urlMatch[3]) > 1) {
					newURL = newURL + urlMatch[3]
				}
				if (len(urlMatch[2]) + 35) < len(newURL) {
					newURL = fmt.Sprintf("(%s) %s", urlMatch[2], chatView.shortener.Shorten(newURL))
				}
				if len(urlMatch) == 5 {
					newURL = newURL + strings.TrimSuffix(urlMatch[4], ">")
				}
				messageText = strings.Replace(messageText, urlMatch[0], newURL, 1)
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

					escaped := strings.Replace(strings.Replace(writer.String(), "*", "\\*", -1), "_", "\\_", -1)
					messageText = strings.Replace(messageText, value, escaped, 1)
				}
			}
		}

		messageText = strings.Replace(strings.Replace(parseBoldAndUnderline(messageText), "\\*", "*", -1), "\\_", "_", -1)

		messageText = strings.Replace(messageText, "|", "\\|", -1)
		shouldShow, contains := chatView.showSpoilerContent[message.ID]
		if !contains || !shouldShow {
			messageText = spoilerRegex.ReplaceAllString(messageText, "[red]!SPOILER![white]")
		}
		messageText = strings.Replace(messageText, "\\|", "|", -1)

		var messageAuthor string
		if message.GuildID != "" {
			member, cacheError := chatView.session.State.Member(message.GuildID, message.Author.ID)
			if cacheError == nil {
				messageAuthor = discordgoplus.GetMemberName(member, &userColor)
			}
		}

		if messageAuthor == "" {
			messageAuthor = discordgoplus.GetUserName(message.Author, &userColor)
		}

		messageText = fmt.Sprintf("[\"%d\"][red]%s [green]%s [white]%s[\"\"][\"\"]", nextIndex, timeCellText, messageAuthor, messageText)
		newText = fmt.Sprintf("%s\n%s", newText, messageText)

		nextIndex++
	}

	fmt.Fprint(chatView.internalTextView, newText)

	chatView.updateHighlights()

	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

func parseBoldAndUnderline(messageText string) string {
	messageTextTemp := make([]rune, 0)

	firstBoldFound := false
	boldOpen := false
	firstUnderlineFound := false
	underlineOpen := false

	runes := []rune(messageText)
	lastIndex := len(runes) - 1
	for index, character := range runes {
		messageTextTemp = append(messageTextTemp, character)
		if character == '\n' {
			if boldOpen && !underlineOpen {
				messageTextTemp = append(messageTextTemp, '[', ':', ':', 'b', ']')
			} else if !boldOpen && underlineOpen {
				messageTextTemp = append(messageTextTemp, '[', ':', ':', 'u', ']')
			} else if boldOpen && underlineOpen {
				messageTextTemp = append(messageTextTemp, '[', ':', ':', 'u', 'b', ']')
			}
		} else if character == '*' {
			if firstBoldFound {
				firstBoldFound = false
				if boldOpen {
					boldOpen = false
					if underlineOpen {
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':', 'u', ']')
					} else {
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':', '-', ']')
					}
				} else if index != lastIndex {
					doesClosingOneExist := false
					foundFirstClosing := false
					for _, c := range runes[index+1:] {
						if c == '*' {
							if foundFirstClosing {
								doesClosingOneExist = true
								break
							}
							foundFirstClosing = true
						}
					}

					if doesClosingOneExist {
						if runes[index+1] == '*' {
							firstBoldFound = true
							continue
						}

						boldOpen = true
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':')
						if underlineOpen {
							messageTextTemp = append(messageTextTemp, 'u')
						}
						messageTextTemp = append(messageTextTemp, 'b', ']')
					}
				}
			} else {
				firstBoldFound = true
			}
		} else if character == '_' {
			if firstUnderlineFound {
				firstUnderlineFound = false
				if underlineOpen {
					underlineOpen = false
					if boldOpen {
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':', 'b', ']')
					} else {
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':', '-', ']')
					}
				} else if index != lastIndex {
					doesClosingOneExist := false
					foundFirstClosing := false
					for _, c := range runes[index+1:] {
						if c == '_' {
							if foundFirstClosing {
								doesClosingOneExist = true
								break
							}
							foundFirstClosing = true
						}
					}

					if doesClosingOneExist {
						if runes[index+1] == '_' {
							firstUnderlineFound = true
							continue
						}

						underlineOpen = true
						messageTextTemp = append(messageTextTemp[:len(messageTextTemp)-2], '[', ':', ':')
						if boldOpen {
							messageTextTemp = append(messageTextTemp, 'b')
						}
						messageTextTemp = append(messageTextTemp, 'u', ']')
					}
				}
			} else {
				firstUnderlineFound = true
			}
		} else {
			firstBoldFound = false
			firstUnderlineFound = false
		}
	}

	return string(messageTextTemp)
}

// ClearSelection clears the current selection of messages.
func (chatView *ChatView) ClearSelection() {
	chatView.selection = -1
	chatView.updateHighlights()
}

// SignalSelectionDeleted notifies the ChatView that its currently selected
// message doesn't exist anymore, moving the selection up by a row if possible.
func (chatView *ChatView) SignalSelectionDeleted() {
	if chatView.selection > 0 {
		chatView.selection--
	}
}

// SetMessages defines all currently displayed messages. Parsing and
// manipulation of single message elements happens in this function.
func (chatView *ChatView) SetMessages(messages []*discordgo.Message) {
	chatView.data = make([]*discordgo.Message, 0)
	chatView.internalTextView.SetText("")

	chatView.AddMessages(messages)
}
