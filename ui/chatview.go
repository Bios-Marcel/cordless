package ui

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	linkshortener "github.com/Bios-Marcel/shortnotforlong"

	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/times"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/config"
	// Blank import for initializing the tview formatter
	_ "github.com/Bios-Marcel/cordless/syntax"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

var (
	linkColor           = "[#efec1c]"
	codeBlockRegex      = regexp.MustCompile("(?s)(^|.)(\x60\x60\x60(.*?)?\n(.+?)\x60\x60\x60)($|.)")
	colorRegex          = regexp.MustCompile("\\[#.{6}\\]")
	channelMentionRegex = regexp.MustCompile(`<#\d*>`)
	urlRegex            = regexp.MustCompile(`<?(https?://)(.+?)(/.+?)?($|\s|\||>)`)
	spoilerRegex        = regexp.MustCompile(`(?s)\|\|(.+?)\|\|`)
	roleMentionRegex    = regexp.MustCompile(`<@&\d*>`)
)

// ChatView is using a tview.TextView in order to be able to display messages
// in a simple way. It supports highlighting specific element types and it
// also supports multiline.
type ChatView struct {
	internalTextView *tview.TextView

	shortener *linkshortener.Shortener

	state      *discordgo.State
	data       []*discordgo.Message
	bufferSize int
	ownUserID  string

	shortenLinks bool

	selection     int
	selectionMode bool

	showSpoilerContent map[string]bool
	formattedMessages  map[string]string

	onMessageAction func(message *discordgo.Message, event *tcell.EventKey) *tcell.EventKey

	mutex *sync.Mutex
}

// NewChatView constructs a new ready to use ChatView.
func NewChatView(state *discordgo.State, ownUserID string) *ChatView {
	chatView := ChatView{
		internalTextView:   tview.NewTextView(),
		state:              state,
		ownUserID:          ownUserID,
		selection:          -1,
		bufferSize:         100,
		selectionMode:      false,
		showSpoilerContent: make(map[string]bool),
		shortenLinks:       config.GetConfig().ShortenLinks,
		formattedMessages:  make(map[string]string),
		mutex:              &sync.Mutex{},
	}

	if chatView.shortenLinks {
		chatView.shortener = linkshortener.NewShortener(config.GetConfig().ShortenerPort)
		go func() {
			shortenerError := chatView.shortener.Start()
			if shortenerError != nil {
				//Disable shortening in case of start failure.
				chatView.shortenLinks = false
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
				} else if chatView.selection >= 1 {
					chatView.selection--
				} else {
					return nil
				}

				chatView.updateHighlights()

				return nil
			}

			if event.Key() == tcell.KeyDown {
				if chatView.selection == -1 {
					chatView.selection = 0
				} else if chatView.selection <= len(chatView.data)-2 {
					chatView.selection++
				} else {
					return nil
				}

				chatView.updateHighlights()

				return nil
			}

			if event.Key() == tcell.KeyHome {
				if chatView.selection != 0 {
					chatView.selection = 0

					chatView.updateHighlights()
				}

				return nil
			}

			if event.Key() == tcell.KeyEnd {
				if chatView.selection != len(chatView.data)-1 {
					chatView.selection = len(chatView.data) - 1

					chatView.updateHighlights()
				}

				return nil
			}

			if chatView.selection > 0 && chatView.selection < len(chatView.data) && event.Rune() == 's' {
				message := chatView.data[chatView.selection]
				messageID := message.ID
				currentValue, contains := chatView.showSpoilerContent[messageID]
				if contains {
					chatView.showSpoilerContent[messageID] = !currentValue
				} else {
					chatView.showSpoilerContent[messageID] = true
				}
				chatView.formattedMessages[messageID] = chatView.formatMessage(message)
				chatView.Rerender()
				return nil
			}

			if chatView.selection >= 0 && chatView.selection < len(chatView.data) && chatView.onMessageAction != nil {
				return chatView.onMessageAction(chatView.data[chatView.selection], event)
			}
		}

		return event
	})

	chatView.internalTextView.
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetBorder(true).
		SetTitleColor(tcell.ColorGreen)

	return &chatView
}

// SetTitle sets the border text of the chatview.
func (chatView *ChatView) SetTitle(text string) {
	chatView.internalTextView.SetTitle(text)
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
	if chatView.selection != -1 {
		chatView.internalTextView.ScrollToHighlight()
	}
}

// GetPrimitive returns the component that can be added to a layout, since
// the ChatView itself is not a component.
func (chatView *ChatView) GetPrimitive() tview.Primitive {
	return chatView.internalTextView
}

// UpdateMessage reformats the passed message, updates the cache and triggers
// a rerender.
func (chatView *ChatView) UpdateMessage(updatedMessage *discordgo.Message) {
	for _, message := range chatView.data {
		if message.ID == updatedMessage.ID {
			chatView.formattedMessages[updatedMessage.ID] = chatView.formatMessage(updatedMessage)
			chatView.Rerender()
			break
		}
	}
}

// DeleteMessage drops the message from the cache and triggers a rerender
func (chatView *ChatView) DeleteMessage(deletedMessage *discordgo.Message) {
	delete(chatView.showSpoilerContent, deletedMessage.ID)
	delete(chatView.formattedMessages, deletedMessage.ID)
	filteredMessages := make([]*discordgo.Message, 0)
	for _, message := range chatView.data {
		if message.ID != deletedMessage.ID {
			filteredMessages = append(filteredMessages, message)
		}
	}
	chatView.data = filteredMessages
	chatView.Rerender()
}

// DeleteMessages drops the messages from the cache and triggers a rerender
func (chatView *ChatView) DeleteMessages(deletedMessages []string) {
	filteredMessages := make([]*discordgo.Message, 0)

	for _, message := range deletedMessages {
		delete(chatView.showSpoilerContent, message)
		delete(chatView.formattedMessages, message)

	}

OUTER_LOOP:
	for _, message := range chatView.data {
		for _, toDelete := range deletedMessages {
			if toDelete == message.ID {
				continue OUTER_LOOP
			}
		}

		filteredMessages = append(filteredMessages, message)
	}

	chatView.data = filteredMessages
	chatView.Rerender()
}

// ClearViewAndCache clears the TextView buffer and removes all data for
// all messages.
func (chatView *ChatView) ClearViewAndCache() {
	chatView.data = make([]*discordgo.Message, 0)
	chatView.showSpoilerContent = make(map[string]bool)
	chatView.formattedMessages = make(map[string]string)
	chatView.selection = -1
	chatView.internalTextView.SetText("")
	chatView.SetTitle("")
}

func (chatView *ChatView) addMessageInternal(message *discordgo.Message) {
	isBlocked := discordutil.IsBlocked(chatView.state, message.Author)

	if !config.GetConfig().ShowPlaceholderForBlockedMessages && isBlocked {
		return
	}

	var rerender bool
	if len(chatView.data) >= chatView.bufferSize {
		idToDrop := chatView.data[0].ID
		delete(chatView.showSpoilerContent, idToDrop)
		delete(chatView.formattedMessages, idToDrop)
		chatView.data = append(chatView.data[1:], message)
		rerender = true
		if chatView.selection > -1 {
			chatView.selection--
		}
		chatView.updateHighlights()
	} else {
		chatView.data = append(chatView.data, message)
	}

	var newText string
	formattedMessage, messageAlreadyFormatted := chatView.formattedMessages[message.ID]
	if messageAlreadyFormatted {
		newText = formattedMessage
	} else {
		if isBlocked {
			newText = messagePartsToColouredString(message.Timestamp, "Blocked user", "Blocked message")
		} else {
			newText = chatView.formatMessage(message)
		}
		chatView.formattedMessages[message.ID] = newText
	}

	if rerender {
		chatView.Rerender()
	} else {
		fmt.Fprint(chatView.internalTextView, "\n[\""+intToString(len(chatView.data)-1)+"\"]"+newText)
	}
}

//AddMessage add an additional message to the ChatView.
func (chatView *ChatView) AddMessage(message *discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()

	chatView.addMessageInternal(message)

	chatView.updateHighlights()
	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

// AddMessages is the same as AddMessage, but for an array of messages instead
// of a single message. Calling this method will not repeat certain actions and
// therefore be slightly more performant than calling AddMessage multiple
// times.
func (chatView *ChatView) AddMessages(messages []*discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()

	for _, message := range messages {
		chatView.addMessageInternal(message)
	}

	chatView.updateHighlights()
	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

// Rerender clears the text view and fills it again using the current cache.
func (chatView *ChatView) Rerender() {
	chatView.internalTextView.SetText("")
	var newContent string
	for index, message := range chatView.data {
		formattedMessage, contains := chatView.formattedMessages[message.ID]
		//Should always be true, otherwise we got ourselves a bug.
		if contains {
			newContent = newContent + "\n[\"" + intToString(index) + "\"]" + formattedMessage
		} else {
			panic("Bug in chatview, a message could not be found.")
		}
	}
	fmt.Fprint(chatView.internalTextView, newContent)
}

func (chatView *ChatView) formatMessage(message *discordgo.Message) string {
	return messagePartsToColouredString(
		message.Timestamp,
		chatView.formatMessageAuthor(message),
		chatView.formatMessageText(message))
}

func (chatView *ChatView) formatMessageAuthor(message *discordgo.Message) string {
	var messageAuthor string
	if message.GuildID != "" {
		member, cacheError := chatView.state.Member(message.GuildID, message.Author.ID)
		if cacheError == nil {
			messageAuthor = discordutil.GetMemberName(member)
		}
	}

	if messageAuthor == "" {
		messageAuthor = discordutil.GetUserName(message.Author)
	}

	if config.GetConfig().UseRandomUserColors {
		messageAuthor = "[" + discordutil.GetUserColor(message.Author) + "]" + messageAuthor
	} else {
		messageAuthor = "[#44e544]" + messageAuthor
	}

	return messageAuthor
}

func (chatView *ChatView) formatMessageText(message *discordgo.Message) string {
	if message.Type == discordgo.MessageTypeDefault {
		return chatView.formatDefaultMessageText(message)
	} else if message.Type == discordgo.MessageTypeGuildMemberJoin {
		return "[gray]joined the server."
	} else if message.Type == discordgo.MessageTypeCall {
		return "[gray]has started a call."
	} else if message.Type == discordgo.MessageTypeChannelIconChange {
		return "[gray]changed the channel icon."
	} else if message.Type == discordgo.MessageTypeChannelNameChange {
		return "[gray]changed the channel name to " + message.Content + "."
	} else if message.Type == discordgo.MessageTypeChannelPinnedMessage {
		return "[gray]pinned a message."
	} else if message.Type == discordgo.MessageTypeRecipientAdd {
		return "[gray]added " + message.Mentions[0].Username + " to the group."
	} else if message.Type == discordgo.MessageTypeRecipientRemove {
		return "[gray]removed " + message.Mentions[0].Username + " from the group."
	}

	//Might happen when there are unsupported types.
	return "[gray]message couldn't be a rendered."
}

func (chatView *ChatView) formatDefaultMessageText(message *discordgo.Message) string {
	messageText := tview.Escape(message.Content)

	//Message.MentionRoles only contains the mentions for mentionable.
	//Therefore we do it like this, in order to render every mention.
	messageText = roleMentionRegex.
		ReplaceAllStringFunc(messageText, func(data string) string {
			roleID := strings.TrimSuffix(strings.TrimPrefix(data, "<@&"), ">")
			role, cacheError := chatView.state.Role(message.GuildID, roleID)
			if cacheError != nil {
				return data
			}

			return linkColor + "@" + role.Name + "[white]"
		})

	messageText = strings.NewReplacer(
		"@everyone", linkColor+"@everyone[white]",
		"@here", linkColor+"@here[white]",
	).Replace(messageText)

	for _, user := range message.Mentions {
		var userName string
		if message.GuildID != "" {
			member, cacheError := chatView.state.Member(message.GuildID, user.ID)
			if cacheError == nil {
				userName = discordutil.GetMemberName(member)
			}
		}

		if userName == "" {
			userName = discordutil.GetUserName(user)
		}

		var color string
		if chatView.state.User.ID == user.ID {
			color = "[#ef9826]"
		} else {
			color = linkColor
		}

		replacement := color + "@" + userName + "[white]"
		messageText = strings.NewReplacer(
			"<@"+user.ID+">", replacement,
			"<@!"+user.ID+">", replacement,
		).Replace(messageText)
	}

	messageText = channelMentionRegex.
		ReplaceAllStringFunc(messageText, func(data string) string {
			channelID := strings.TrimSuffix(strings.TrimPrefix(data, "<#"), ">")
			channel, cacheError := chatView.state.Channel(channelID)
			if cacheError != nil {
				return data
			}

			return linkColor + "#" + channel.Name + "[white]"
		})

	// FIXME Needs improvement, as it wastes space and breaks things
	if message.Attachments != nil && len(message.Attachments) > 0 {
		var attachments []string
		for _, attachment := range message.Attachments {
			attachments = append(attachments, attachment.URL)
		}
		attachmentsAsText := strings.Join(attachments, " ")

		if messageText != "" {
			messageText = messageText + "\n" + attachmentsAsText
		} else {
			messageText = attachmentsAsText
		}
	}

	// FIXME Handle Non-embed links nonetheless?
	if chatView.shortenLinks {
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

	codeBlocks := codeBlockRegex.
		// Magicnumber, because message aren't gonna be that long anyway.
		FindAllStringSubmatch(messageText, 1000)

	for _, values := range codeBlocks {
		language := values[3]
		code := values[4]

		//Remove last \n on the last line of code, also taking windows
		//line endings into account.
		code = strings.TrimSuffix(strings.TrimSuffix(code, "\n"), "\r")
		code = removeLeadingWhitespaceInCode(code)

		// Determine lexer.
		l := lexers.Get(language)
		if l == nil {
			l = lexers.Analyse(code)
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

		it, tokeniseError := l.Tokenise(nil, code)
		if tokeniseError != nil {
			continue
		}

		writer := bytes.NewBufferString("")

		formatError := f.Format(writer, s, it)
		if formatError != nil {
			continue
		}

		escapedCode := strings.NewReplacer("*", "\\*", "_", "\\_", "|", "\\|").Replace(writer.String())
		var formattedCode string
		lines := strings.Split(escapedCode, "\n")
		var lastColor string
		for index, line := range lines {
			if index != 0 {
				colorCodes := colorRegex.FindAllString(lines[index-1], -1)
				if len(colorCodes) > 0 {
					lastColor = colorCodes[len(colorCodes)-1]
				}

				if lastColor != "" {
					formattedCode += fmt.Sprintf("[#c9dddc]▐ %s%s", lastColor, line)
					if index != len(lines)-1 {
						formattedCode += "\n"
					}
					continue
				}
			}

			formattedCode += "[#c9dddc]▐ " + line
			if index != len(lines)-1 {
				formattedCode += "\n"
			}
		}

		beforeCodeBlock := values[1]
		if beforeCodeBlock != "\n" {
			formattedCode = "\n" + formattedCode
		}
		afterCodeBlock := values[5]
		if len(afterCodeBlock) != 0 && afterCodeBlock != "\n" {
			formattedCode = formattedCode + "\n"
		}

		messageText = strings.Replace(messageText, values[2], formattedCode, 1)
	}

	messageText = strings.Replace(strings.Replace(parseBoldAndUnderline(messageText), "\\*", "*", -1), "\\_", "_", -1)

	shouldShow, contains := chatView.showSpoilerContent[message.ID]
	if !contains || !shouldShow {
		messageText = spoilerRegex.ReplaceAllString(messageText, "[red]!SPOILER![white]")
	}
	messageText = strings.Replace(messageText, "\\|", "|", -1)

	return messageText
}

func trimMinAmountOfCharacterAsPrefix(charToTrim rune, text string) (string, int) {
	lines := strings.Split(text, "\n")
	minAmountOfCharacter := math.MaxInt32
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		amountOfCharacters := 0
		for _, character := range []rune(line) {
			if character != charToTrim {
				break
			}

			amountOfCharacters++
		}

		if amountOfCharacters < minAmountOfCharacter {
			minAmountOfCharacter = amountOfCharacters
		}

		if amountOfCharacters == 0 {
			break
		}
	}

	if minAmountOfCharacter > 0 {
		toTrim := strings.Repeat(string(charToTrim), minAmountOfCharacter)
		for index, line := range lines {
			lines[index] = strings.TrimPrefix(line, toTrim)
		}

		return strings.Join(lines, "\n"), minAmountOfCharacter
	}

	return text, 0
}

func removeLeadingWhitespaceInCode(code string) string {
	spacesTrimmed, amountTrimmed := trimMinAmountOfCharacterAsPrefix(' ', code)
	if amountTrimmed > 0 {
		return spacesTrimmed
	}

	tabsTrimmed, _ := trimMinAmountOfCharacterAsPrefix('	', code)
	return tabsTrimmed
}

func messagePartsToColouredString(timestamp discordgo.Timestamp, author, message string) string {
	time, parseError := timestamp.Parse()
	var timeCellText string
	if parseError == nil {
		timeCellText = times.TimeToLocalString(&time)
	}

	return fmt.Sprintf("[gray]%s %s [white]%s[\"\"][\"\"]", timeCellText, author, message)
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

// Lock will lock the ChatView, allowing other callers to prevent race
// conditions.
func (chatView *ChatView) Lock() {
	chatView.mutex.Lock()
}

// Unlock unlocks the previously locked ChatView.
func (chatView *ChatView) Unlock() {
	chatView.mutex.Unlock()
}
