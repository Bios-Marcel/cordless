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
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/times"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"

	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"

	// Blank import for initializing the tview formatter
	_ "github.com/Bios-Marcel/cordless/syntax"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

const dashCharacter = "\u2500"

var (
	successiveCustomEmojiRegex = regexp.MustCompile("<a?:.+?:\\d+(><)a?:.+?:\\d+>")
	customEmojiRegex           = regexp.MustCompile("(?sm)(.?)<(a?):(.+?):(\\d+)>(.?)")
	codeBlockRegex             = regexp.MustCompile("(?sm)(^|.)?(\x60\x60\x60(.*?)?\n(.+?)\x60\x60\x60)($|.)")
	colorRegex                 = regexp.MustCompile("\\[#.{6}\\]")
	channelMentionRegex        = regexp.MustCompile(`<#\d*>`)
	urlRegex                   = regexp.MustCompile(`<?(https?://)(.+?)(/.+?)?($|\s|\||>)`)
	spoilerRegex               = regexp.MustCompile(`(?s)\|\|(.+?)\|\|`)
	roleMentionRegex           = regexp.MustCompile(`<@&\d*>`)
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
	format     string

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
		format:             "2006-01-02",
		selection:          -1,
		bufferSize:         100,
		selectionMode:      false,
		showSpoilerContent: make(map[string]bool),
		shortenLinks:       config.Current.ShortenLinks,
		formattedMessages:  make(map[string]string),
		mutex:              &sync.Mutex{},
	}

	if chatView.shortenLinks {
		chatView.shortener = linkshortener.NewShortener(config.Current.ShortenerPort)
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

				chatView.refreshSelectionAndScrollToSelection()

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

				chatView.refreshSelectionAndScrollToSelection()

				return nil
			}

			if event.Key() == tcell.KeyHome {
				if chatView.selection != 0 {
					chatView.selection = 0

					chatView.refreshSelectionAndScrollToSelection()
				}

				return nil
			}

			if event.Key() == tcell.KeyEnd {
				if chatView.selection != len(chatView.data)-1 {
					chatView.selection = len(chatView.data) - 1

					chatView.refreshSelectionAndScrollToSelection()
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
				chatView.Reprint()
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
		SetIndicateOverflow(true).
		SetBorder(true).
		SetTitleColor(config.GetTheme().InverseTextColor)

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

func (chatView *ChatView) refreshSelectionAndScrollToSelection() {
	if chatView.selection == -1 {
		//Empty basically clears the highlights
		chatView.internalTextView.Highlight("")
	} else {
		chatView.internalTextView.Highlight(intToString(chatView.selection))
		chatView.internalTextView.ScrollToHighlight()
	}
}

// GetPrimitive returns the component that can be added to a layout, since
// the ChatView itself is not a component.
func (chatView *ChatView) GetPrimitive() tview.Primitive {
	return chatView.internalTextView
}

// UpdateMessage reformats the passed message, updates the cache and triggers
// a reprint.
func (chatView *ChatView) UpdateMessage(updatedMessage *discordgo.Message) {
	for _, message := range chatView.data {
		if message.ID == updatedMessage.ID {
			chatView.formattedMessages[updatedMessage.ID] = chatView.formatMessage(updatedMessage)
			chatView.Reprint()
			break
		}
	}
}

// DeleteMessage drops the message from the cache and triggers a reprint
func (chatView *ChatView) DeleteMessage(deletedMessage *discordgo.Message) {
	delete(chatView.showSpoilerContent, deletedMessage.ID)
	delete(chatView.formattedMessages, deletedMessage.ID)

	for index, message := range chatView.data {
		if message.ID == deletedMessage.ID {
			chatView.data = append(chatView.data[:index], chatView.data[index+1:]...)
			chatView.Reprint()
			break
		}
	}
}

// DeleteMessages drops the messages from the cache and triggers a reprint
func (chatView *ChatView) DeleteMessages(deletedMessages []string) {
	for _, message := range deletedMessages {
		delete(chatView.showSpoilerContent, message)
		delete(chatView.formattedMessages, message)
	}

	filteredMessages := make([]*discordgo.Message, 0, len(chatView.data)-len(deletedMessages))
OUTER_LOOP:
	for _, message := range chatView.data {
		for _, toDelete := range deletedMessages {
			if toDelete == message.ID {
				continue OUTER_LOOP
			}
		}

		filteredMessages = append(filteredMessages, message)
	}

	if len(chatView.data) != len(filteredMessages) {
		chatView.data = filteredMessages
		chatView.Reprint()
	}
}

// ClearViewAndCache clears the TextView buffer and removes all data for
// all messages.
func (chatView *ChatView) ClearViewAndCache() {
	chatView.data = make([]*discordgo.Message, 0)
	chatView.showSpoilerContent = make(map[string]bool)
	chatView.formattedMessages = make(map[string]string)
	chatView.selection = -1
	chatView.internalTextView.Clear()
	chatView.SetTitle("")
}

// addMessageInternal prints a new message to the textview or triggers a
// rerender. It also takes the blocked relation into consideration.
func (chatView *ChatView) addMessageInternal(message *discordgo.Message) {
	isBlocked := discordutil.IsBlocked(chatView.state, message.Author)

	if !config.Current.ShowPlaceholderForBlockedMessages && isBlocked {
		return
	}

	var reprint bool
	if len(chatView.data) >= chatView.bufferSize {
		idToDrop := chatView.data[0].ID
		delete(chatView.showSpoilerContent, idToDrop)
		delete(chatView.formattedMessages, idToDrop)
		chatView.data = append(chatView.data[1:], message)
		reprint = true

		//Moving up the selection, since we have removed the first message. If
		//the previously selected message was the first message, then no
		// message will be selected.
		if chatView.selection > -1 {
			chatView.selection--
		}

		chatView.refreshSelectionAndScrollToSelection()
	} else {
		chatView.data = append(chatView.data, message)
	}

	var newText string
	formattedMessage, messageAlreadyFormatted := chatView.formattedMessages[message.ID]
	if messageAlreadyFormatted {
		newText = formattedMessage
	} else {
		if isBlocked {
			newText = chatView.messagePartsToColouredString(message.Timestamp, "Blocked user", "Blocked message")
		} else {
			newText = chatView.formatMessage(message)
		}
		chatView.formattedMessages[message.ID] = newText
	}

	if reprint {
		chatView.Reprint()
	} else {
		fmt.Fprint(chatView.internalTextView, "\n[\""+intToString(len(chatView.data)-1)+"\"]"+newText)
	}
}

//AddMessage add an additional message to the ChatView.
func (chatView *ChatView) AddMessage(message *discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()
	t1, _ := chatView.data[len(chatView.data)-1].Timestamp.Parse()
	t2, _ := message.Timestamp.Parse()
	if !times.AreDatesTheSameDay(t1.Local(), t2.Local()) {
		fmt.Fprint(chatView.internalTextView, chatView.CreateDateDelimiter(t2.Format(chatView.format)))
	}

	chatView.addMessageInternal(message)
	chatView.refreshSelectionAndScrollToSelection()
	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

// CreateDateDelimiter creates a date delimiter between messages to mark the date and returns it
func (chatView *ChatView) CreateDateDelimiter(date string) string {
	_, _, width, _ := chatView.internalTextView.GetInnerRect()
	characterAmountLeftForDashes := width - len(date) - 2 /* Because of the spaces */
	amountDashesLeft := characterAmountLeftForDashes / 2
	dashesLeft := strings.Repeat(dashCharacter, amountDashesLeft)
	dashesRight := strings.Repeat(dashCharacter, characterAmountLeftForDashes-amountDashesLeft)
	return "\n[\"\"]" + dashesLeft + " " + date + " " + dashesRight
}

// ReturnDateDelimiter creates datedelimiters between two messages and returns them
func (chatView *ChatView) ReturnDateDelimiter(messages []*discordgo.Message, index int) string {
	if index == 0 {
		time, _ := messages[index].Timestamp.Parse()
		date := time.Format(chatView.format)
		return chatView.CreateDateDelimiter(date)
	}

	t1, _ := messages[index-1].Timestamp.Parse()
	t2, _ := messages[index].Timestamp.Parse()

	if !times.AreDatesTheSameDay(t1.Local(), t2.Local()) {
		return chatView.CreateDateDelimiter(t2.Local().Format(chatView.format))
	}

	return ""
}

// WriteDateDelimiter runs ReturnDateDelimiter and writes it to chatView.internalTextView
func (chatView *ChatView) WriteDateDelimiter(messages []*discordgo.Message, index int) {
	fmt.Fprint(chatView.internalTextView, chatView.ReturnDateDelimiter(messages, index))
}

// AddMessages is the same as AddMessage, but for an array of messages instead
// of a single message. Calling this method will not repeat certain actions and
// therefore be slightly more performant than calling AddMessage multiple
// times.
func (chatView *ChatView) AddMessages(messages []*discordgo.Message) {
	wasScrolledToTheEnd := chatView.internalTextView.IsScrolledToEnd()

	for index, message := range messages {
		chatView.WriteDateDelimiter(messages, index)
		chatView.addMessageInternal(message)
	}

	chatView.refreshSelectionAndScrollToSelection()
	if wasScrolledToTheEnd {
		chatView.internalTextView.ScrollToEnd()
	}
}

// Reprint clears the internal TextView and prints all currently cached
// messages into the internal TextView again. This will not actually cause a
// redraw in the user interface. This would still only be done by
// ForceDraw ,QueueUpdateDraw or user events. Calling this method is
// necessary if previously added content has changed or has been removed, since
// can only append to the TextViews buffers, but not cut parts out.
func (chatView *ChatView) Reprint() {
	var newContent string
	for index, message := range chatView.data {
		formattedMessage, contains := chatView.formattedMessages[message.ID]
		//Should always be true, otherwise we got ourselves a bug.
		if contains {
			newContent += chatView.ReturnDateDelimiter(chatView.data, index)
			newContent = newContent + "\n[\"" + intToString(index) + "\"]" + formattedMessage
		} else {
			panic("Bug in chatview, a message could not be found.")
		}
	}
	chatView.internalTextView.SetText(newContent)
}

func (chatView *ChatView) formatMessage(message *discordgo.Message) string {
	return chatView.messagePartsToColouredString(
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

	if config.Current.UseRandomUserColors {
		return "[" + discordutil.GetUserColor(message.Author) + "]" + messageAuthor
	}

	return "[" + tviewutil.ColorToHex(config.GetTheme().DefaultUserColor) + "]" + messageAuthor
}

func (chatView *ChatView) formatMessageText(message *discordgo.Message) string {
	if message.Type == discordgo.MessageTypeDefault {
		return chatView.formatDefaultMessageText(message)
	} else if message.Type == discordgo.MessageTypeGuildMemberJoin {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]joined the server."
	} else if message.Type == discordgo.MessageTypeCall {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]has started a call."
	} else if message.Type == discordgo.MessageTypeChannelIconChange {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]changed the channel icon."
	} else if message.Type == discordgo.MessageTypeChannelNameChange {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]changed the channel name to " + message.Content + "."
	} else if message.Type == discordgo.MessageTypeChannelPinnedMessage {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]pinned a message."
	} else if message.Type == discordgo.MessageTypeRecipientAdd {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]added " + message.Mentions[0].Username + " to the group."
	} else if message.Type == discordgo.MessageTypeRecipientRemove {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]removed " + message.Mentions[0].Username + " from the group."
	} else if message.Type == discordgo.MessageTypeChannelFollowAdd {
		return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]has added '" + message.Content + "' to this channel"
	}

	//TODO Support boost messages; Would be handy to see what they look like first.

	//Might happen when there are unsupported types.
	return "[" + tviewutil.ColorToHex(config.GetTheme().InfoMessageColor) + "]message couldn't be rendered."
}

func (chatView *ChatView) formatDefaultMessageText(message *discordgo.Message) string {
	messageText := tviewutil.Escape(message.Content)

	//Message.MentionRoles only contains the mentions for mentionable.
	//Therefore we do it like this, in order to render every mention.
	messageText = roleMentionRegex.
		ReplaceAllStringFunc(messageText, func(data string) string {
			roleID := strings.TrimSuffix(strings.TrimPrefix(data, "<@&"), ">")
			role, cacheError := chatView.state.Role(message.GuildID, roleID)
			if cacheError != nil {
				return data
			}

			return "[" + tviewutil.ColorToHex(config.GetTheme().LinkColor) + "]@" + role.Name + "[" + tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor) + "]"
		})

	messageText = strings.NewReplacer("@everyone", "["+tviewutil.ColorToHex(config.GetTheme().LinkColor)+"]@everyone["+
		tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)+"]", "@here",
		"["+tviewutil.ColorToHex(config.GetTheme().LinkColor)+"]@here["+tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)+"]",
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
			color = "[" + tviewutil.ColorToHex(config.GetTheme().AttentionColor) + "]"
		} else {
			color = "[" + tviewutil.ColorToHex(config.GetTheme().LinkColor) + "]"
		}

		replacement := color + "@" + userName + "[" + tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor) + "]"
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

			return "[" + tviewutil.ColorToHex(config.GetTheme().LinkColor) + "]#" + channel.Name + "[" + tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor) + "]"
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

		//Remove all carriage returns to prevent bugs with windows newlines.
		code = strings.ReplaceAll(code, "\r", "")
		//Remove last newline, as it's usually just the newline that seperates code from markdown notation.
		code = strings.TrimSuffix(code, "\n")
		code = removeLeadingWhitespaceInCode(code)

		// Determine lexer.
		l := lexers.Get(language)
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

		//Remove the last newline, as some formatters behave differently and don't drop it.
		escapedCode := strings.NewReplacer("*", "\\*", "_", "\\_", "|", "\\|").Replace(writer.String())
		newLineDifference := strings.Count(escapedCode, "\n") - strings.Count(code, "\n")
		if newLineDifference > 0 {
			for ; newLineDifference != 0; newLineDifference-- {
				escapedCode = escapedCode[:(strings.LastIndex(escapedCode, "\n"))]
			}
		}
		var formattedCode, lastColor string
		lines := strings.Split(escapedCode, "\n")
		for index, line := range lines {
			if index != 0 {
				formattedCode += "\n"
				colorCodes := colorRegex.FindAllString(lines[index-1], -1)
				if len(colorCodes) > 0 {
					lastColor = colorCodes[len(colorCodes)-1]
				}

				if lastColor != "" {
					formattedCode += fmt.Sprintf("[#c9dddc]▐ %s%s", lastColor, line)
					continue
				}
			}

			formattedCode += "[#c9dddc]▐ " + line
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

	messageText = strings.
		NewReplacer("\\*", "*", "\\_", "_", "\\`", "`").
		Replace(parseBoldAndUnderline(messageText))
	messageText = parseCustomEmojis(messageText)

	shouldShow, contains := chatView.showSpoilerContent[message.ID]
	if !contains || !shouldShow {
		messageText = spoilerRegex.ReplaceAllString(messageText, "["+tviewutil.ColorToHex(config.GetTheme().AttentionColor)+"]!SPOILER!["+tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)+"]")
	}
	messageText = strings.Replace(messageText, "\\|", "|", -1)

	return messageText
}

func parseCustomEmojis(text string) string {
	messageText := text

	//Little hack, since the customEmojiRegex can't handle <:emoji:123><:emoji:123>
	//And we do it this way in order to allow overlapping matches.
	var matches [][]int
	for resume := true; resume; resume = len(matches) > 0 {
		matches = successiveCustomEmojiRegex.FindAllStringSubmatchIndex(messageText, -1)
		//Iterating backwards, since the indexes would be incorrect after
		//the first match otherwise
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			messageText = messageText[:match[2]+1] + " " + messageText[match[3]-1:]
		}
	}

	customEmojiMatches := customEmojiRegex.FindAllStringSubmatch(messageText, -1)
	for _, match := range customEmojiMatches {
		customEmojiCode := match[3]
		if len(match[2]) > 0 {
			customEmojiCode = "a:" + customEmojiCode
		}
		url := tviewutil.Escape("[" + customEmojiCode + "]( https://cdn.discordapp.com/emojis/" + match[4] + " )")
		if match[1] != "" && match[1] != "\n" {
			url = match[1] + "\n" + url
		}
		if match[5] != "" && match[5] != "\n" {
			url = url + "\n" + match[5]
		}
		messageText = strings.Replace(messageText, match[0], url, 1)
	}

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

func (chatView *ChatView) messagePartsToColouredString(timestamp discordgo.Timestamp, author, message string) string {
	time, parseError := timestamp.Parse()
	var timeCellText string
	if parseError == nil {
		timeCellText = times.TimeToLocalString(&time)
	}

	return fmt.Sprintf("["+tviewutil.ColorToHex(config.GetTheme().MessageTimeColor)+"]%s %s ["+tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)+"]%s[\"\"][\"\"]", timeCellText, author, message)
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
	chatView.refreshSelectionAndScrollToSelection()
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
	chatView.data = make([]*discordgo.Message, 0, len(messages))
	chatView.internalTextView.Clear()

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
