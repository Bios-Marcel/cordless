package ui

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/skratchdot/open-golang/open"

	"github.com/Bios-Marcel/cordless/fileopen"
	"github.com/Bios-Marcel/cordless/logging"
	"github.com/Bios-Marcel/cordless/util/files"
	"github.com/Bios-Marcel/cordless/util/fuzzy"
	"github.com/Bios-Marcel/cordless/util/text"
	"github.com/Bios-Marcel/cordless/version"

	"github.com/Bios-Marcel/discordemojimap"
	"github.com/Bios-Marcel/goclipimg"

	"github.com/atotto/clipboard"

	"github.com/Bios-Marcel/discordgo"
	"github.com/gdamore/tcell"
	"github.com/gen2brain/beeep"

	"github.com/Bios-Marcel/cordless/tview"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/scripting"
	"github.com/Bios-Marcel/cordless/scripting/js"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui/shortcutdialog"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/maths"
)

var (
	shortcutsDialogShortcut = tcell.NewEventKey(tcell.KeyCtrlK, rune(tcell.KeyCtrlK), tcell.ModCtrl)
)

// Window is basically the whole application, as it contains all the
// components and the necessary global state.
type Window struct {
	app               *tview.Application
	middleContainer   *tview.Flex
	rootContainer     *tview.Flex
	dialogReplacement *tview.Flex
	dialogButtonBar   *tview.Flex
	dialogTextView    *tview.TextView

	leftArea    *tview.Flex
	guildList   *GuildList
	guildPage   *tview.Flex
	channelTree *ChannelTree
	privateList *PrivateChatList

	chatArea         *tview.Flex
	chatView         *ChatView
	messageContainer tview.Primitive
	messageInput     *Editor

	editingMessageID *string
	messageLoader    *discordutil.MessageLoader

	userList *UserTree

	session *discordgo.Session

	selectedGuild   *discordgo.Guild
	previousGuild   *discordgo.Guild
	selectedChannel *discordgo.Channel
	previousChannel *discordgo.Channel

	extensionEngines []scripting.Engine

	commandMode bool
	commandView *CommandView
	commands    []commands.Command

	userActive      bool
	userActiveTimer *time.Timer

	doRestart chan bool

	bareChat   bool
	activeView ActiveView
}

type ActiveView bool

const Guilds ActiveView = true
const Dms ActiveView = false

//NewWindow constructs the whole application window and also registers all
//necessary handlers and functions. If this function returns an error, we can't
//start the application.
func NewWindow(doRestart chan bool, app *tview.Application, session *discordgo.Session, readyEvent *discordgo.Ready) (*Window, error) {
	window := &Window{
		doRestart:        doRestart,
		session:          session,
		app:              app,
		activeView:       Guilds,
		extensionEngines: []scripting.Engine{js.New()},
		messageLoader:    discordutil.CreateMessageLoader(session),
	}

	if config.Current.DesktopNotificationsUserInactivityThreshold > 0 {
		window.userActiveTimer = time.NewTimer(time.Duration(config.Current.DesktopNotificationsUserInactivityThreshold) * time.Second)
		go func() {
			for {
				<-window.userActiveTimer.C
				window.userActive = false
			}
		}()
	}

	window.commandView = NewCommandView(window.ExecuteCommand)
	logging.SetAdditionalOutput(window.commandView)

	for _, engine := range window.extensionEngines {
		initError := window.initExtensionEngine(engine)
		if initError != nil {
			return nil, initError
		}
	}

	guilds := readyEvent.Guilds

	mentionWindowRootNode := tview.NewTreeNode("")
	autocompleteView := tview.NewTreeView().
		SetVimBindingsEnabled(false).
		SetRoot(mentionWindowRootNode).
		SetTopLevel(1).
		SetCycleSelection(true)
	autocompleteView.SetBorder(true)
	autocompleteView.SetBorderSides(false, true, false, true)

	window.leftArea = tview.NewFlex().SetDirection(tview.FlexRow)

	window.guildPage = tview.NewFlex()
	window.guildPage.SetDirection(tview.FlexRow)

	channelTree := NewChannelTree(window.session.State)
	window.channelTree = channelTree
	channelTree.SetOnChannelSelect(func(channelID string) {
		channel, cacheError := window.session.State.Channel(channelID)
		if cacheError == nil && channel.Type != discordgo.ChannelTypeGuildCategory {
			go func() {
				window.chatView.Lock()
				defer window.chatView.Unlock()
				window.QueueUpdateDrawSynchronized(func() {
					loadError := window.LoadChannel(channel)
					if loadError == nil {
						channelTree.MarkChannelAsLoaded(channelID)
					}
				})
			}()
		}
	})
	window.registerGuildChannelHandler()

	discordutil.SortGuilds(window.session.State.Settings, guilds)
	guildList := NewGuildList(guilds)
	window.guildList = guildList
	window.guildList.UpdateUnreadGuildCount()
	guildList.SetOnGuildSelect(func(guildID string) {
		//Update previously selected guild.
		if window.selectedGuild != nil {
			window.updateServerReadStatus(window.selectedGuild.ID, false)
		}

		guild, cacheError := window.session.Guild(guildID)
		if cacheError != nil {
			window.ShowErrorDialog(cacheError.Error())
			return
		}

		// previousGuild and previousGuildNode should be set initially.
		// If going from first guild -> private chat, SwitchToPreviousChannel would crash.
		if window.selectedGuild == nil {
			window.previousGuild = guild
		} else if window.previousGuild != window.selectedGuild {
			window.previousGuild = window.selectedGuild
		}

		window.selectedGuild = guild

		window.updateServerReadStatus(guild.ID, true)

		//FIXME Request presences as soon as that stuff remotely works?
		requestError := session.RequestGuildMembers(guildID, "", 0, false)
		if requestError != nil {
			fmt.Fprintln(window.commandView, "Error retrieving all guild members.")
		}

		channelLoadError := window.channelTree.LoadGuild(guildID)
		if channelLoadError != nil {
			window.ShowErrorDialog(channelLoadError.Error())
		} else {
			if config.Current.FocusChannelAfterGuildSelection {
				app.SetFocus(window.channelTree)
			}
		}

		//Currently has to happen before userlist loading, as it might not load otherwise
		window.RefreshLayout()

		window.userList.Clear()
		if window.userList.internalTreeView.IsVisible() {
			userLoadError := window.userList.LoadGuild(guildID)
			if userLoadError != nil {
				window.ShowErrorDialog(userLoadError.Error())
			}
		}

	})

	window.registerGuildHandlers()
	window.registerGuildMemberHandlers()

	window.guildPage.AddItem(guildList, 0, 1, true)
	window.guildPage.AddItem(channelTree, 0, 2, false)

	window.privateList = NewPrivateChatList(window.session.State)
	window.privateList.Load()
	window.registerPrivateChatsHandler()

	window.leftArea.AddItem(window.privateList.GetComponent(), 1, 0, false)
	window.leftArea.AddItem(window.guildPage, 0, 1, false)

	window.privateList.SetOnChannelSelect(func(channelID string) {
		channel, stateError := window.session.State.Channel(channelID)
		if stateError != nil {
			window.ShowErrorDialog(fmt.Sprintf("Error loading chat: %s", stateError.Error()))
			return
		}

		go func() {
			window.chatView.Lock()
			defer window.chatView.Unlock()
			window.QueueUpdateDrawSynchronized(func() {
				window.userList.Clear()
				window.RefreshLayout()

				if channel.Type == discordgo.ChannelTypeGroupDM {

					if window.userList.internalTreeView.IsVisible() {
						loadError := window.userList.LoadGroup(channel.ID)
						if loadError != nil {
							fmt.Fprintln(window.commandView.commandOutput, "Error loading users for channel.")
						}
					}
				}

			})
			window.QueueUpdateDrawSynchronized(func() {
				window.LoadChannel(channel)
			})
		}()
	})

	window.privateList.SetOnFriendSelect(func(userID string) {
		go func() {
			window.chatView.Lock()
			defer window.chatView.Unlock()
			userChannels, _ := window.session.UserChannels()
			for _, userChannel := range userChannels {
				if userChannel.Type == discordgo.ChannelTypeDM && userChannel.Recipients[0].ID == userID {
					window.QueueUpdateDrawSynchronized(func() {
						window.loadPrivateChannel(userChannel)
					})
					return
				}
			}

			newChannel, discordError := window.session.UserChannelCreate(userID)
			if discordError == nil {
				messages, discordError := window.session.ChannelMessages(newChannel.ID, 100, "", "", "")
				if discordError == nil {
					for _, message := range messages {
						window.session.State.MessageAdd(message)
					}
				}
				window.QueueUpdateDrawSynchronized(func() {
					window.loadPrivateChannel(newChannel)
				})
			}
		}()
	})

	window.chatArea = tview.NewFlex().
		SetDirection(tview.FlexRow)

	window.chatView = NewChatView(window.session.State, window.session.State.User.ID)
	window.chatView.SetOnMessageAction(func(message *discordgo.Message, event *tcell.EventKey) *tcell.EventKey {
		if shortcuts.QuoteSelectedMessage.Equals(event) {
			window.insertQuoteOfMessage(message)
			return nil
		}

		if shortcuts.NewDirectMessage.Equals(event) {
			//Can't message yourself, goon!
			if message.Author.ID == window.session.State.User.ID {
				return nil
			}

			//If there's an existing channel, we use that and avoid unnecessary traffic.
			existingChannel := discordutil.FindDMChannelWithUser(window.session.State, message.Author.ID)
			if existingChannel != nil {
				window.SwitchToPrivateChannel(existingChannel)
				return nil
			}

			newChannel, createError := window.session.UserChannelCreate(message.Author.ID)
			if createError != nil {
				window.ShowErrorDialog(createError.Error())
			} else {
				window.SwitchToPrivateChannel(newChannel)
			}
			return nil
		}

		if shortcuts.ReplySelectedMessage.Equals(event) {
			window.messageInput.SetText("@" + message.Author.Username + "#" + message.Author.Discriminator + " " + window.messageInput.GetText())
			app.SetFocus(window.messageInput.GetPrimitive())
			return nil
		}

		if shortcuts.CopySelectedMessageLink.Equals(event) {
			copyError := clipboard.WriteAll(fmt.Sprintf("<https://discordapp.com/channels/@me/%s/%s>", message.ChannelID, message.ID))
			if copyError != nil {
				window.ShowErrorDialog(fmt.Sprintf("Error copying message link: %s", copyError.Error()))
			}
			return nil
		}

		if shortcuts.DeleteSelectedMessage.Equals(event) {
			if message.Author.ID == window.session.State.User.ID {
				window.askForMessageDeletion(message.ID, true)
			}
			return nil
		}

		if shortcuts.EditSelectedMessage.Equals(event) {
			window.startEditingMessage(message)
			return nil
		}

		if shortcuts.CopySelectedMessage.Equals(event) {
			copyError := clipboard.WriteAll(discordutil.MessageToPlainText(message))
			if copyError != nil {
				window.ShowErrorDialog(fmt.Sprintf("Error copying message: %s", copyError.Error()))
			}
			return nil
		}

		if shortcuts.ViewSelectedMessageImages.Equals(event) {
			var targetFolder string

			if config.Current.FileOpenSaveFilesPermanently {
				absolutePath, pathError := files.ToAbsolutePath(config.Current.FileDownloadSaveLocation)
				if pathError == nil {
					targetFolder = absolutePath
				}
			}

			if targetFolder == "" {
				cacheDir, osError := os.UserCacheDir()
				if osError == nil && cacheDir != "" {
					//Own subdirectory to avoid nuking foreing files by accident.
					targetFolder = filepath.Join(cacheDir, "cordless")
					makeDirError := os.MkdirAll(targetFolder, 0766)
					if makeDirError != nil {
						window.ShowCustomErrorDialog("Couldn't open file", "Can't create cache subdirectory.")
						return nil
					}
				}
			}

			if targetFolder == "" {
				window.ShowCustomErrorDialog("Couldn't open file", "Can't find cache directory.")
			} else {
				for _, file := range message.Attachments {
					openError := fileopen.OpenFile(targetFolder, file.ID, file.URL)
					if openError != nil {
						window.ShowCustomErrorDialog("Couldn't open file", openError.Error())
					}
				}

				urlMatches := urlRegex.FindAllString(message.Content, 1000)
				for _, url := range urlMatches {
					header, _ := http.Head(url)

					//A website! Any other text/ could be a file, like .txt, .css or whatever.
					//Is there a more bulletproof way to doing this?
					if strings.Contains(header.Header.Get("Content-Type"), "text/html") {
						//We hope to just open this with the users browser ;)
						open.Run(url)
						continue
					}

					openError := fileopen.OpenFile(targetFolder, "file", url)
					if openError != nil {
						window.ShowCustomErrorDialog("Couldn't open file", openError.Error())
					}
				}
			}

			//If permanent saving isn't disabled, we clear files older
			//than two weeks whenever something is opened. Since this
			//will happen in a background thread, it won't cause
			//application blocking.
			if !config.Current.FileOpenSaveFilesPermanently && targetFolder != "" {
				fileopen.LaunchCacheCleaner(targetFolder, time.Hour*(24*14))
			}

			return nil
		}

		if shortcuts.DownloadMessageFiles.Equals(event) {
			absolutePath, pathError := files.ToAbsolutePath(config.Current.FileDownloadSaveLocation)
			if pathError != nil || absolutePath == "" {
				window.ShowErrorDialog("Please specify a valid path in 'FileOpenSaveFolder' of your configuration.")
			} else {
				downloadFunction := func(savePath, fileURL string) {
					_, statErr := os.Stat(savePath)
					//If the file exists already, we needn't do anything.
					if statErr == nil {
						return
					}

					downloadError := files.DownloadFile(savePath, fileURL)
					if downloadError != nil {
						window.app.QueueUpdateDraw(func() {
							window.ShowErrorDialog("Error download file: " + downloadError.Error())
						})
					}
				}

				for _, file := range message.Attachments {
					extension := strings.TrimPrefix(filepath.Ext(file.URL), ".")
					targetFile := filepath.Join(absolutePath, file.ID+"."+extension)

					//All files are downloaded separately in order to not
					//block the UI and not download for ages if one or more
					//page has a slow download speed.
					go downloadFunction(targetFile, file.URL)
				}

				urlMatches := urlRegex.FindAllString(message.Content, 1000)
				for _, url := range urlMatches {
					baseName := filepath.Base(url)
					if baseName == "" {
						continue
					}

					targetFile := filepath.Join(absolutePath, filepath.Base(url))

					//All files are downloaded separately in order to not
					//block the UI and not download for ages if one or more
					//page has a slow download speed.
					go downloadFunction(targetFile, url)
				}
			}
			return nil
		}

		return event
	})
	window.messageContainer = window.chatView.GetPrimitive()

	window.messageInput = NewEditor()
	window.messageInput.internalTextView.SetIndicateOverflow(true)
	window.messageInput.SetOnHeightChangeRequest(func(height int) {
		_, _, _, chatViewHeight := window.chatView.internalTextView.GetRect()
		newHeight := maths.Min(height, chatViewHeight/2)

		window.chatArea.ResizeItem(window.messageInput.GetPrimitive(), newHeight, 0)
	})

	window.messageInput.SetAutocompleteValuesUpdateHandler(func(values []*AutocompleteValue) {
		autocompleteView.GetRoot().ClearChildren()
		if len(values) == 0 {
			autocompleteView.SetVisible(false)
			window.app.SetFocus(window.messageInput.GetPrimitive())
		} else {
			rootNode := autocompleteView.GetRoot()
			for _, value := range values {
				newNode := tview.NewTreeNode(value.RenderValue)
				newNode.SetReference(value)
				rootNode.AddChild(newNode)
			}
			autocompleteView.SetCurrentNode(rootNode)
			autocompleteView.SetVisible(true)
			window.app.SetFocus(autocompleteView)
			window.chatArea.ResizeItem(autocompleteView, maths.Min(10, len(values)), 0)
		}
	})

	autocompleteView.SetSelectedFunc(func(node *tview.TreeNode) {
		value := node.GetReference().(*AutocompleteValue)
		window.messageInput.Autocomplete(value.InsertValue)
		window.app.SetFocus(window.messageInput.GetPrimitive())
		autocompleteView.SetVisible(false)
	})

	window.messageInput.RegisterAutocomplete('#', false, func(value string) []*AutocompleteValue {
		if window.selectedChannel != nil && window.selectedChannel.GuildID != "" {
			guild, stateError := session.State.Guild(window.selectedChannel.GuildID)
			if stateError != nil {
				return nil
			}

			filtered := fuzzy.ScoreAndSortChannels(value, guild.Channels)
			var autocompleteValues []*AutocompleteValue
			for _, channel := range filtered {
				if channel.Type != discordgo.ChannelTypeGuildText {
					continue
				}

				autocompleteValues = append(autocompleteValues, &AutocompleteValue{
					RenderValue: channel.Name,
					InsertValue: "<#" + channel.ID + ">",
				})
			}

			return autocompleteValues
		}

		return nil
	})

	window.messageInput.RegisterAutocomplete('@', true, func(value string) []*AutocompleteValue {
		if window.selectedChannel != nil {
			guildID := window.selectedChannel.GuildID
			var autocompleteValues []*AutocompleteValue
			if guildID != "" {
				guild, stateError := session.State.Guild(guildID)
				if stateError == nil {
					filteredRoles := fuzzy.ScoreAndSortRoles(value, guild.Roles)
					for _, role := range filteredRoles {
						//Workaround for discord just having a default role called "@everyone"
						if role.Name == "@everyone" {
							autocompleteValues = append(autocompleteValues, &AutocompleteValue{
								RenderValue: role.Name,
								InsertValue: "@everyone",
							})
						} else {
							autocompleteValues = append(autocompleteValues, &AutocompleteValue{
								RenderValue: role.Name,
								InsertValue: "<@&" + role.ID + ">",
							})

						}
					}

					filteredMembers := fuzzy.ScoreAndSortMembers(value, guild.Members)
					for _, member := range filteredMembers {
						insertValue := member.User.Username + "#" + member.User.Discriminator
						var renderValue string
						if member.Nick != "" {
							renderValue = insertValue + " | " + member.Nick
						} else {
							renderValue = insertValue
						}
						autocompleteValues = append(autocompleteValues, &AutocompleteValue{
							RenderValue: renderValue,
							InsertValue: "@" + insertValue,
						})
					}

					return autocompleteValues
				}
			}

			filtered := fuzzy.ScoreAndSortUsers(value, window.selectedChannel.Recipients)
			for _, user := range filtered {
				insertValue := user.Username + "#" + user.Discriminator
				autocompleteValues = append(autocompleteValues, &AutocompleteValue{
					RenderValue: insertValue,
					InsertValue: "@" + insertValue,
				})
			}

			return autocompleteValues
		}

		return nil
	})

	emojisAsArray := make([]string, 0, len(discordemojimap.EmojiMap))
	for emoji := range discordemojimap.EmojiMap {
		emojisAsArray = append(emojisAsArray, emoji)
	}

	var globallyUsableCustomEmoji []*discordgo.Emoji
	if window.session.State.User.PremiumType == discordgo.UserPremiumTypeNone {
		for _, guild := range window.session.State.Guilds {
			for _, emoji := range guild.Emojis {
				if emoji.Animated {
					continue
				}

				if !strings.HasPrefix(emoji.Name, "GW") {
					continue
				}

				globallyUsableCustomEmoji = append(globallyUsableCustomEmoji, emoji)
			}
		}
	} else {
		for _, guild := range window.session.State.Guilds {
			for _, emoji := range guild.Emojis {
				globallyUsableCustomEmoji = append(globallyUsableCustomEmoji, emoji)
			}
		}
	}

	emojisByGuild := make(map[string][]*discordgo.Emoji)

	window.messageInput.RegisterAutocomplete(':', false, func(value string) []*AutocompleteValue {
		var autocompleteValues []*AutocompleteValue

		var customEmojiUsableInContext []*discordgo.Emoji
		if window.session.State.User.PremiumType == discordgo.UserPremiumTypeNone {
			if window.selectedChannel != nil && window.selectedChannel.GuildID != "" {
				guildID := window.selectedChannel.GuildID
				var cached bool
				customEmojiUsableInContext, cached = emojisByGuild[guildID]
				if cached {
					goto EVALUATE_EMOJIS
				}

				//Non premium users can only use the non-animated guildemojis
				guild, stateError := window.session.State.Guild(guildID)
				if stateError == nil {
					for _, emoji := range guild.Emojis {
						if emoji.Animated {
							continue
						}

						customEmojiUsableInContext = append(customEmojiUsableInContext, emoji)
					}

					customEmojiUsableInContext = append(customEmojiUsableInContext, globallyUsableCustomEmoji...)
					emojisByGuild[window.selectedChannel.GuildID] = customEmojiUsableInContext
				}
			} else {
				//If not in any guild channel, we can only use the global ones
				customEmojiUsableInContext = globallyUsableCustomEmoji
			}
		} else {
			//For non-nitro users everything's available anyway
			customEmojiUsableInContext = globallyUsableCustomEmoji
		}

	EVALUATE_EMOJIS:
		filteredEmoji := fuzzy.ScoreAndSortEmoji(value, emojisAsArray, customEmojiUsableInContext)
		for _, emoji := range filteredEmoji {
			unicodeSymbol := discordemojimap.GetEmoji(emoji)
			var renderValue string
			var insertValue string
			if unicodeSymbol != "" {
				renderValue = unicodeSymbol + " | " + emoji
				insertValue = unicodeSymbol
			} else {
				trimmed := emoji[1:]
				renderValue = "? | " + trimmed + " (custom emoji)"
				insertValue = ":!" + trimmed + ":"
			}
			autocompleteValues = append(autocompleteValues, &AutocompleteValue{
				RenderValue: renderValue,
				InsertValue: insertValue,
			})
		}

		return autocompleteValues
	})

	captureFunc := func(event *tcell.EventKey) *tcell.EventKey {
		messageToSend := window.messageInput.GetText()

		if event.Modifiers() == tcell.ModCtrl {
			if event.Key() == tcell.KeyUp {
				window.chatView.internalTextView.ScrollUp()
				return nil
			}

			if event.Key() == tcell.KeyDown {
				window.chatView.internalTextView.ScrollDown()
				return nil
			}
		}

		if event.Key() == tcell.KeyPgUp {
			handler := window.chatView.internalTextView.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone), nil)
			return nil
		}

		if event.Key() == tcell.KeyPgDn {
			handler := window.chatView.internalTextView.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), nil)
			return nil
		}

		chooseNextMessageToEdit := func(loopStart, loopEnd int, iterNext func(int) int) *discordgo.Message {
			if len(window.chatView.data) == 0 {
				return nil
			}

			window.chatView.Lock()
			defer window.chatView.Unlock()

			var chooseNextMatch bool
			for i := loopStart; i != loopEnd; i = iterNext(i) {
				message := window.chatView.data[i]
				if message.Author.ID == window.session.State.User.ID {
					if !chooseNextMatch && window.editingMessageID != nil && *window.editingMessageID == message.ID {
						chooseNextMatch = true
						continue
					}

					if window.editingMessageID == nil || chooseNextMatch {
						return message
					}
				}
			}

			return nil
		}

		//When you are already typing a message, you probably don't want to risk loosing it.
		if event.Key() == tcell.KeyUp && (messageToSend == "" || window.editingMessageID != nil) {
			messageToEdit := chooseNextMessageToEdit(len(window.chatView.data)-1, -1, func(i int) int { return i - 1 })
			if messageToEdit != nil {
				window.startEditingMessage(messageToEdit)
			}

			return nil
		}

		if event.Key() == tcell.KeyDown && window.editingMessageID != nil {
			messageToEdit := chooseNextMessageToEdit(0, len(window.chatView.data), func(i int) int { return i + 1 })
			if messageToEdit != nil {
				window.startEditingMessage(messageToEdit)
			}

			return nil
		}

		if event.Key() == tcell.KeyEsc {
			window.exitMessageEditMode()
			return nil
		}

		if event.Key() == tcell.KeyCtrlV && window.selectedChannel != nil {
			data, clipError := goclipimg.GetImageFromClipboard()

			if clipError == goclipimg.ErrNoImageInClipboard {
				return event
			}

			if clipError == nil {
				dataChannel := bytes.NewReader(data)
				targetChannel := window.selectedChannel
				currentText := window.prepareMessage(targetChannel, strings.TrimSpace(window.messageInput.GetText()))
				if currentText == "" {
					go window.session.ChannelFileSend(targetChannel.ID, "img.png", dataChannel)
				} else {
					messageData := &discordgo.MessageSend{
						Content: currentText,
						File: &discordgo.File{
							Name:        "img.png",
							ContentType: "image/png",
							Reader:      dataChannel,
						},
					}
					go window.session.ChannelMessageSendComplex(targetChannel.ID, messageData)
					window.messageInput.SetText("")
				}
			} else {
				window.ShowErrorDialog(fmt.Sprintf("Error pasting image: %s", clipError.Error()))
			}

			return nil
		}

		if shortcuts.AddNewLineInCodeBlock.Equals(event) && window.IsCursorInsideCodeBlock() {
			window.insertNewLineAtCursor()
			return nil
		} else if shortcuts.SendMessage.Equals(event) {
			if window.selectedChannel != nil {
				window.TrySendMessage(window.selectedChannel, messageToSend)
			}
			return nil
		}

		return event
	}
	window.messageInput.SetInputCapture(captureFunc)

	//FIXME Buffering might just be retarded, as the event handlers are launched in separate routines either way.
	messageInputChan := make(chan *discordgo.Message)
	messageDeleteChan := make(chan *discordgo.Message)
	messageEditChan := make(chan *discordgo.Message)
	messageBulkDeleteChan := make(chan *discordgo.MessageDeleteBulk)

	window.registerMessageEventHandler(messageInputChan, messageEditChan, messageDeleteChan, messageBulkDeleteChan)
	window.startMessageHandlerRoutines(messageInputChan, messageEditChan, messageDeleteChan, messageBulkDeleteChan)

	window.userList = NewUserTree(window.session.State)

	if config.Current.OnTypeInListBehaviour == config.SearchOnTypeInList {
		guildList.SetSearchOnTypeEnabled(true)
		channelTree.SetSearchOnTypeEnabled(true)
		window.userList.internalTreeView.SetSearchOnTypeEnabled(true)
		window.privateList.internalTreeView.SetSearchOnTypeEnabled(true)
	} else if config.Current.OnTypeInListBehaviour == config.FocusMessageInputOnTypeInList {
		focusTextViewOnTypeInputHandler := tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView)
		guildList.SetInputCapture(focusTextViewOnTypeInputHandler)
		channelTree.SetInputCapture(focusTextViewOnTypeInputHandler)
		window.userList.SetInputCapture(focusTextViewOnTypeInputHandler)
		window.privateList.SetInputCapture(focusTextViewOnTypeInputHandler)
		window.chatView.internalTextView.SetInputCapture(focusTextViewOnTypeInputHandler)
	}

	newGuildHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if shortcuts.GuildListMarkRead.Equals(event) {
			selectedGuildNode := guildList.GetCurrentNode()
			if selectedGuildNode != nil && !readstate.HasGuildBeenRead(selectedGuildNode.GetReference().(string)) {
				ackError := window.session.GuildMessageAck(selectedGuildNode.GetReference().(string))
				if ackError != nil {
					window.ShowErrorDialog(ackError.Error())
				}
			}
			return nil
		}

		return event
	}

	oldGuildListHandler := guildList.GetInputCapture()
	if oldGuildListHandler == nil {
		guildList.SetInputCapture(newGuildHandler)
	} else {
		guildList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newGuildHandler(event)
			if handledEvent != nil {
				return oldGuildListHandler(event)
			}

			return event
		})
	}

	newChannelListHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if shortcuts.ChannelTreeMarkRead.Equals(event) {
			selectedChannelNode := channelTree.GetCurrentNode()
			if selectedChannelNode != nil {
				ackError := discordutil.AcknowledgeChannel(window.session, selectedChannelNode.GetReference().(string))
				if ackError != nil {
					window.ShowErrorDialog(ackError.Error())
				}
			}
			return nil
		}

		return event
	}

	oldChannelListHandler := channelTree.GetInputCapture()
	if oldChannelListHandler == nil {
		channelTree.SetInputCapture(newChannelListHandler)
	} else {
		channelTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newChannelListHandler(event)
			if handledEvent != nil {
				return oldChannelListHandler(event)
			}

			return event
		})
	}

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.MessageAck) {
		window.app.QueueUpdateDraw(func() {
			if readstate.UpdateReadLocal(event.ChannelID, event.MessageID) {
				channel, stateError := s.State.Channel(event.ChannelID)
				if stateError == nil && event.MessageID == channel.LastMessageID {
					if channel.GuildID == "" {
						window.privateList.MarkChannelAsRead(channel.ID)
					} else {
						if window.selectedGuild != nil && channel.GuildID == window.selectedGuild.ID {
							window.channelTree.MarkChannelAsRead(channel.ID)
						} else {
							window.updateServerReadStatus(channel.GuildID, false)
						}
					}
				}
			}
		})
	})

	window.middleContainer = tview.NewFlex().
		SetDirection(tview.FlexColumn)

	window.rootContainer = tview.NewFlex().
		SetDirection(tview.FlexRow)
	window.rootContainer.SetTitleAlign(tview.AlignCenter)
	window.rootContainer.AddItem(window.middleContainer, 0, 1, false)

	window.dialogReplacement = tview.NewFlex().
		SetDirection(tview.FlexRow)

	window.dialogTextView = tview.NewTextView()
	window.dialogReplacement.AddItem(window.dialogTextView, 0, 1, false)

	window.dialogButtonBar = tview.NewFlex().
		SetDirection(tview.FlexColumn)

	window.dialogReplacement.AddItem(window.dialogButtonBar, 1, 0, false)
	window.dialogReplacement.SetVisible(false)

	window.rootContainer.AddItem(window.dialogReplacement, 2, 0, false)

	if config.Current.ShowBottomBar {
		bottomBar := NewBottomBar()
		bottomBar.AddItem(fmt.Sprintf("Logged in as: '%s'", tviewutil.Escape(session.State.User.Username)))
		bottomBar.AddItem(fmt.Sprintf("View / Change shortcuts: %s", shortcutdialog.EventToString(shortcutsDialogShortcut)))
		window.rootContainer.AddItem(bottomBar, 1, 0, false)
	}

	app.SetRoot(window.rootContainer, true)
	app.SetInputCapture(window.handleGlobalShortcuts)

	if config.Current.UseFixedLayout {
		window.middleContainer.AddItem(window.leftArea, config.Current.FixedSizeLeft, 0, true)
		window.middleContainer.AddItem(window.chatArea, 0, 1, false)
		window.middleContainer.AddItem(window.userList.internalTreeView, config.Current.FixedSizeRight, 0, false)
	} else {
		window.middleContainer.AddItem(window.leftArea, 0, 7, true)
		window.middleContainer.AddItem(window.chatArea, 0, 20, false)
		window.middleContainer.AddItem(window.userList.internalTreeView, 0, 6, false)
	}

	autocompleteView.SetVisible(false)
	autocompleteView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Key(); key {
		case tcell.KeyRune, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft, tcell.KeyRight, tcell.KeyCtrlA, tcell.KeyCtrlV:
			window.messageInput.internalTextView.GetInputCapture()(event)
			return nil
		}
		return event
	})

	window.chatArea.AddItem(window.messageContainer, 0, 1, false)
	window.chatArea.AddItem(autocompleteView, 2, 2, true)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), window.messageInput.GetRequestedHeight(), 0, false)

	window.commandView.commandOutput.SetVisible(false)
	window.commandView.commandInput.internalTextView.SetVisible(false)

	window.chatArea.AddItem(window.commandView.commandOutput, 0, 1, false)
	window.chatArea.AddItem(window.commandView.commandInput.internalTextView, 3, 0, false)

	window.SwitchToGuildsPage()

	app.SetFocusDirectionHandler(tview.Up, shortcuts.FocusUp.Equals)
	app.SetFocusDirectionHandler(tview.Down, shortcuts.FocusDown.Equals)
	app.SetFocusDirectionHandler(tview.Left, shortcuts.FocusLeft.Equals)
	app.SetFocusDirectionHandler(tview.Right, shortcuts.FocusRight.Equals)

	window.messageInput.internalTextView.SetNextFocusableComponents(tview.Up, window.chatView.internalTextView)
	window.messageInput.internalTextView.SetNextFocusableComponents(tview.Down, window.commandView.commandOutput, window.chatView.internalTextView)
	window.messageInput.internalTextView.SetNextFocusableComponents(tview.Right, window.userList.internalTreeView, window.channelTree, window.privateList.internalTreeView)
	window.messageInput.internalTextView.SetNextFocusableComponents(tview.Left, window.channelTree, window.privateList.internalTreeView)

	window.channelTree.SetNextFocusableComponents(tview.Up, window.guildList)
	window.channelTree.SetNextFocusableComponents(tview.Down, window.guildList)
	window.channelTree.SetNextFocusableComponents(tview.Left, window.userList.internalTreeView, window.commandView.commandOutput, window.messageInput.GetPrimitive())
	window.channelTree.SetNextFocusableComponents(tview.Right, window.commandView.commandOutput, window.messageInput.GetPrimitive())

	window.guildList.SetNextFocusableComponents(tview.Up, window.channelTree)
	window.guildList.SetNextFocusableComponents(tview.Down, window.channelTree)
	window.guildList.SetNextFocusableComponents(tview.Left, window.userList.internalTreeView, window.chatView.GetPrimitive())
	window.guildList.SetNextFocusableComponents(tview.Right, window.chatView.GetPrimitive(), window.userList.internalTreeView)

	window.privateList.internalTreeView.SetNextFocusableComponents(tview.Right, window.chatView.GetPrimitive())
	window.privateList.internalTreeView.SetNextFocusableComponents(tview.Left, window.userList.internalTreeView, window.chatView.GetPrimitive())

	window.userList.internalTreeView.SetNextFocusableComponents(tview.Left, window.chatView.GetPrimitive())

	window.chatView.internalTextView.SetNextFocusableComponents(tview.Down, window.messageInput.GetPrimitive())
	window.chatView.internalTextView.SetNextFocusableComponents(tview.Up, window.commandView.commandInput.internalTextView, window.messageInput.GetPrimitive())

	window.commandView.commandInput.internalTextView.SetNextFocusableComponents(tview.Up, window.commandView.commandOutput)
	window.commandView.commandInput.internalTextView.SetNextFocusableComponents(tview.Down, window.chatView.GetPrimitive())
	window.commandView.commandInput.internalTextView.SetNextFocusableComponents(tview.Right, window.userList.internalTreeView, window.channelTree, window.privateList.internalTreeView)
	window.commandView.commandInput.internalTextView.SetNextFocusableComponents(tview.Left, window.channelTree, window.privateList.internalTreeView)

	window.commandView.commandOutput.SetNextFocusableComponents(tview.Up, window.messageInput.GetPrimitive())
	window.commandView.commandOutput.SetNextFocusableComponents(tview.Down, window.commandView.commandInput.GetPrimitive())
	window.commandView.commandOutput.SetNextFocusableComponents(tview.Right, window.userList.internalTreeView, window.channelTree, window.privateList.internalTreeView)
	window.commandView.commandOutput.SetNextFocusableComponents(tview.Left, window.channelTree, window.privateList.internalTreeView)

	app.SetFocus(guildList)

	if config.Current.MouseEnabled {
		window.registerMouseFocusListeners()
	}

	window.chatView.internalTextView.SetText(getWelcomeText())

	return window, nil
}

func getWelcomeText() string {
	return fmt.Sprintf(splashText+`

Welcome to version %s of Cordless. Below you can see the most
important changes of the last two versions officially released.

[::b]THIS VERSION
	- Features
		- DM people via "p" in the chatview or use the dm-open command
		- Mark guilds as read
		- Mark guild channels as read
		- Write to logfile by setting "--log"
		- Mentions are now displayed in the guild list
		- You can now bulk send folders and files
	- Changes
		- There's now a double-colon to separate author and messages
		- There's more customizable shortcuts now
	- Bugfixes
		- Muted guilds, channels and categories shouldn't be displayed as
		  unread anymore
		- @everyone works again, so you can piss of others again
		- Messages containing links won't disappear anymore after sending
		- Messages from blocked users won't trigger notifications anymore
		- No more spammed empty error messages when receiving notifications
[::b]2020-08-30
	- Features
		- Nicknames can now be disabled via the configuration
		- Files from messages can now be downloaded (key d) or opened (key o)
		- New parameter "--account" to start cordless with a certain account
	- Changes
		- The "friends" command now has "friend" as an alias
		- "logout" is now a separate command, but "account logout" still works
		- Currently active account is now highlight in "account list" output
		- Password input dialog now uses the configured shortcut for paste
		- Baremode
			- Now includes the message input
			- The command view will hide when entering baremode
	- Bugfixes
		- Fix crash due to race condition in readmarker feature
		- Embed-Edits won't be ignored anymore
		- Names with role colors now respect their role order
		- Unread message numbers now always update when loading a channel instead of when leaving it
		- UTF-8 disabling wasn't taken into account when rendering the channel tree
[::b]2020-08-11 - 2020-06-30
	- Features
		- Notifications for servers and DMs are now displayed in the containers header row 
		- Embeds can now be rendered
		- Usernames can now be rendered with their respective role color.
		  Bots however can't have colors, to avoid confusion with real users.
		  The default is set to "single", meaning it uses the default user
		  color from the specified theme. The setting "UseRandomUserColors" has
		  been removed.
	- Changes
		- The button to switch between DMs and servers is gone. Instead you can
		  click the containers, since the header row is always visible now
		- Token input now ingores surrounding spaces
		- Bot token syntax is more lenient now
	-  Bugfixes
		- Bot login works again
		- Holding down your left mouse and moving it on the chatview won't
		  cause lags anymore
		- No more false positives for unread dm-channels
[::b]20-06-26
	- Features
		- you can now define a custom status
		- shortened URLs optionally can display a file suffix (extension)
		- You can now cycle through message in edit-mode by repeatedly hitting KeyUp/Down
	- Bugfixes
		- config directory path now read from "XDF_CONFIG_HOME" instead of "XDG_CONFIG_DIR"
		- the delete message shortcut was pointing to the same value as "show spoilered message"
		- the lack of the config directory would cause a crash
		- nitro users couldn't use emojis anymore
		- several typos have been corrected
		- the "version" command printed it's help output to stdout
		- the "man" command now searches through the content of pages and suggests those
[::b]2020-01-05
	- Features
		- VT320 terminals are now supported
		- quoted messages now preserve attachment URLs
		- Ctrl-W now deletes the word to the left
		- announcement channels are now shown as well
		- Cordless now has an amazing autocompletion
		- support for TFA
		- user-set command allows supplying emojis
		- custom emojis are now rendered as links
		- login now navigable via arrow keys
		- Ctrl-B now toggles the so called "bare mode", giving all space to the chat
		- configuration path is now customizable via parameters
	- Bugfixes
		- emoji sequences with underscores now work
		- text channels sometimes didn't show up'
		- Cordless doesn't crash anymore when sending a message into an empty channel
		- attachment links are now copied as well
	- Performance improvements
		- the usertree will now load lazily
		- dummycall to validate session token has been removed
	- Changes
		- login button has been removed ... just hit enter ;)
		- tokeninput on login is now masked
		- Docs have been improved
	- JS API
		- there's now an "init" function that gets called on script load
`, version.Version)
}

// initExtensionEngine injections necessary functions into the engine.
// those functions can be called by each script inside of an engine.
func (window *Window) initExtensionEngine(engine scripting.Engine) error {
	engine.SetErrorOutput(window.commandView.commandOutput)
	if err := engine.LoadScripts(config.GetScriptDirectory()); err != nil {
		return err
	}

	engine.SetTriggerNotificationFunction(func(title, text string) {
		notifyError := beeep.Notify("Cordless - "+title, text, "assets/information.png")
		if notifyError != nil {
			log.Printf("["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error sending notification:\n\t[%s]%s\n", tviewutil.ColorToHex(config.GetTheme().ErrorColor), notifyError)
		}
	})

	engine.SetGetCurrentGuildFunction(func() string {
		if window.selectedGuild != nil {
			return window.selectedGuild.ID
		}
		return ""
	})

	engine.SetGetCurrentChannelFunction(func() string {
		if window.selectedChannel != nil {
			return window.selectedChannel.ID
		}
		return ""
	})

	// Even though scripts might already have functions for logging, like the
	// JS engine already has console.log, this is sadly hardcoded to print to
	// stdout instead of a custom specified IO writer.

	engine.SetPrintToConsoleFunction(func(text string) {
		fmt.Fprint(window.commandView, text)
	})

	engine.SetPrintLineToConsoleFunction(func(text string) {
		fmt.Fprintln(window.commandView, text)
	})

	return nil
}

func (window *Window) loadPrivateChannel(channel *discordgo.Channel) {
	window.LoadChannel(channel)
	window.RefreshLayout()
}

// SwitchToPrivateChannel switches to the friends page, loads the given channel
// and then focuses the input primitive.
func (window *Window) SwitchToPrivateChannel(channel *discordgo.Channel) {
	window.SwitchToFriendsPage()
	window.app.SetFocus(window.messageInput.GetPrimitive())
	window.loadPrivateChannel(channel)
}

func (window *Window) insertNewLineAtCursor() {
	window.messageInput.InsertCharacter('\n')
	window.app.QueueUpdateDraw(func() {
		window.messageInput.TriggerHeightRequestIfNecessary()
		window.messageInput.internalTextView.ScrollToHighlight()
	})
}

// IsCursorInsideCodeBlock checks if the cursor comes after three backticks
// that don't have another 3 backticks following after them.
func (window *Window) IsCursorInsideCodeBlock() bool {
	left := window.messageInput.GetTextLeftOfSelection()
	leftSplit := strings.Split(left, "```")
	return len(leftSplit)%2 == 0
}

func getUsernameForQuote(state *discordgo.State, message *discordgo.Message) string {
	if message.GuildID != "" {
		//The error handling here is rather lax, since not being able to show
		// a nickname isn't really a problem worth crashing over.
		guild, stateError := state.Guild(message.GuildID)
		if stateError == nil {
			member, stateError := state.Member(guild.ID, message.Author.ID)
			if stateError == nil && member.Nick != "" {
				return member.Nick
			}
		}
	}

	//Fallback if no respective member can be found, the cache couldn't be
	//accessed or we are in a private chat.
	return message.Author.Username
}

func (window *Window) insertQuoteOfMessage(message *discordgo.Message) {
	username := getUsernameForQuote(window.session.State, message)
	quotedMessage, generateError := discordutil.GenerateQuote(message.ContentWithMentionsReplaced(), username, message.Timestamp, message.Attachments, window.messageInput.GetText())
	if generateError == nil {
		window.messageInput.SetText(quotedMessage)
		window.app.SetFocus(window.messageInput.GetPrimitive())
	} else {
		window.ShowErrorDialog(fmt.Sprintf("Error quoting message:\n\t%s", generateError.Error()))
	}
}

func (window *Window) TrySendMessage(targetChannel *discordgo.Channel, message string) {
	if targetChannel == nil {
		return
	}

	if len(message) == 0 {
		if window.editingMessageID != nil {
			msgIDCopy := *window.editingMessageID
			window.askForMessageDeletion(msgIDCopy, true)
		}
		return
	}

	message = strings.TrimSpace(message)
	if len(message) == 0 {
		window.app.QueueUpdateDraw(func() {
			window.messageInput.SetText("")
		})
		return
	}

	if window.editingMessageID != nil {
		window.editMessage(targetChannel.ID, *window.editingMessageID, message)
		return
	}

	if strings.HasPrefix(message, "file://") {
		window.app.QueueUpdateDraw(func() {
			yesButton := "Yes"
			window.ShowDialog(config.GetTheme().PrimitiveBackgroundColor, "Resolve filepath and send a file instead?", func(button string) {
				if button == yesButton {
					window.messageInput.SetText("")
					go func() {
						path, pathError := files.ToAbsolutePath(message)
						if pathError != nil {
							window.app.QueueUpdateDraw(func() {
								window.ShowErrorDialog(pathError.Error())
							})
							return
						}
						data, readError := ioutil.ReadFile(path)
						if readError != nil {
							window.app.QueueUpdateDraw(func() {
								window.ShowErrorDialog(readError.Error())
							})
							return
						}
						reader := bytes.NewBuffer(data)
						_, sendError := window.session.ChannelFileSend(targetChannel.ID, filepath.Base(message), reader)
						if sendError != nil {
							window.app.QueueUpdateDraw(func() {
								window.ShowErrorDialog(sendError.Error())
							})
						}
					}()
				} else {
					window.sendMessageWithLengthCheck(targetChannel, message)
				}
			}, yesButton, "No")
		})
		return
	}

	window.sendMessageWithLengthCheck(targetChannel, message)
}

func (window *Window) sendMessageWithLengthCheck(targetChannel *discordgo.Channel, message string) {
	message = window.prepareMessage(targetChannel, message)
	overlength := len(message) - 2000
	if overlength > 0 {
		window.app.QueueUpdateDraw(func() {
			sendAsFile := "Send as file"
			window.ShowDialog(config.GetTheme().PrimitiveBackgroundColor, fmt.Sprintf("Your message is %d characters too long, what do you want to do?", overlength),
				func(button string) {
					if button == sendAsFile {
						window.messageInput.SetText("")
						go window.sendMessageAsFile(message, targetChannel.ID)
					}
				}, sendAsFile, "Nothing")
		})
		return
	}

	go window.sendMessage(targetChannel.ID, message)
}

func (window *Window) sendMessageAsFile(message string, channel string) {
	discordutil.SendMessageAsFile(window.session, message, channel, func(sendError error) {
		retry := "Retry sending"
		edit := "Edit"
		window.app.QueueUpdateDraw(func() {
			window.ShowDialog(config.GetTheme().ErrorColor,
				fmt.Sprintf("Error sending message: %s.\n\nWhat do you want to do?", sendError),
				func(button string) {
					switch button {
					case retry:
						go window.sendMessageAsFile(channel, message)
					case edit:
						window.messageInput.SetText(message)
					}
				}, retry, edit, "Cancel")
		})
	})
}

func (window *Window) sendMessage(targetChannelID, message string) {
	window.app.QueueUpdateDraw(func() {
		window.messageInput.SetText("")
		window.chatView.internalTextView.ScrollToEnd()
	})
	_, sendError := window.session.ChannelMessageSend(targetChannelID, message)
	if sendError != nil {
		window.app.QueueUpdateDraw(func() {
			retry := "Retry sending"
			edit := "Edit"
			cancel := "Cancel"
			window.ShowDialog(config.GetTheme().ErrorColor,
				fmt.Sprintf("Error sending message: %s.\n\nWhat do you want to do?", sendError),
				func(button string) {
					switch button {
					case retry:
						go window.sendMessage(targetChannelID, message)
					case edit:
						window.messageInput.SetText(message)
					}
				}, retry, edit, cancel)
		})
	}
}

func (window *Window) updateServerReadStatus(guildID string, isSelected bool) {
	guild, cacheError := window.session.State.Guild(guildID)
	if cacheError == nil {
		window.guildList.UpdateNodeStateByGuild(guild, isSelected)
		window.guildList.UpdateUnreadGuildCount()
	}
}

// prepareMessage prepares a message for being sent to the discord API.
// This will do all necessary escaping and resolving of channel-mentions,
// user-mentions, emojis and the likes.
//
// The input is expected to be a string without surrounding whitespace.
func (window *Window) prepareMessage(targetChannel *discordgo.Channel, inputText string) string {
	message := codeBlockRegex.ReplaceAllStringFunc(inputText, func(input string) string {
		return strings.ReplaceAll(input, ":", "\\:")
	})

	for _, engine := range window.extensionEngines {
		message = engine.OnMessageSend(message)
	}

	if targetChannel.GuildID != "" {
		channelGuild, discordError := window.session.State.Guild(targetChannel.GuildID)
		if discordError == nil {
			//Those could be optimized by searching the string for patterns.
			for _, channel := range channelGuild.Channels {
				if channel.Type == discordgo.ChannelTypeGuildText {
					message = strings.ReplaceAll(message, "#"+channel.Name, "<#"+channel.ID+">")
				}
			}

			message = window.replaceEmojiSequences(channelGuild, message)
		}
	} else {
		message = window.replaceEmojiSequences(nil, message)
	}

	message = strings.Replace(message, "\\:", ":", -1)

	if targetChannel.GuildID == "" {
		for _, user := range targetChannel.Recipients {
			message = strings.ReplaceAll(message, "@"+user.Username+"#"+user.Discriminator, "<@"+user.ID+">")
		}
	} else {
		members, discordError := window.session.State.Members(targetChannel.GuildID)
		if discordError == nil {
			for _, member := range members {
				message = strings.ReplaceAll(message, "@"+member.User.Username+"#"+member.User.Discriminator, "<@"+member.User.ID+">")
			}
		}
	}

	return message
}

// mergeRuneSlices copies the passed rune arrays into a new rune array of the
// correct size.
func mergeRuneSlices(a, b, c []rune) *[]rune {
	length := len(a) + len(b) + len(c)
	result := make([]rune, length, length)
	copy(result[:len(a)], a)
	copy(result[len(a):len(a)+len(b)], b)
	copy(result[len(a)+len(b):], c)
	return &result
}

// replaceEmojiSequences replaces all emoji codes for custom emojis and unicode
// emojis alike. The matching is case-insensitive. It can't differentiate
// between different custom emojis. Forcing the usage of a custom emoji can be
// done by adding a '!' being the first ':'.
// For private channels, the channelGuild may be nil.
func (window *Window) replaceEmojiSequences(channelGuild *discordgo.Guild, message string) string {
	asRunes := []rune(message)
	indexes := text.FindEmojiIndices(asRunes)
INDEX_LOOP:
	for i := 0; i < len(indexes); i += 2 {
		startIndex := indexes[i]
		endIndex := indexes[i+1]
		emojiSequence := strings.ToLower(string(asRunes[startIndex+1 : endIndex]))
		if !strings.HasPrefix(emojiSequence, "!") {
			emoji := discordemojimap.GetEmoji(emojiSequence)
			if emoji != "" {
				asRunes = *mergeRuneSlices(asRunes[:startIndex], []rune(emoji), asRunes[endIndex+1:])
				continue INDEX_LOOP
			}
		}

		emojiSequence = strings.TrimPrefix(emojiSequence, "!")

		if window.session.State.User.PremiumType == discordgo.UserPremiumTypeNitroClassic ||
			window.session.State.User.PremiumType == discordgo.UserPremiumTypeNitro {
			for _, guild := range window.session.State.Guilds {
				for _, emoji := range guild.Emojis {
					if strings.EqualFold(emoji.Name, emojiSequence) {
						var emojiRunes []rune
						if emoji.Animated {
							emojiRunes = []rune("<a:" + emoji.Name + ":" + emoji.ID + ">")
						} else {
							emojiRunes = []rune("<:" + emoji.Name + ":" + emoji.ID + ">")
						}
						asRunes = *mergeRuneSlices(asRunes[:startIndex], emojiRunes, asRunes[endIndex+1:])
						continue INDEX_LOOP
					}
				}
			}
		} else {
			//Local guild emoji take priority
			if channelGuild != nil {
				emoji := discordutil.FindEmojiInGuild(window.session, channelGuild, true, emojiSequence)
				if emoji != "" {
					asRunes = *mergeRuneSlices(asRunes[:startIndex], []rune(emoji), asRunes[endIndex+1:])
					continue INDEX_LOOP
				}
			}

			//Check for global emotes
			for _, guild := range window.session.State.Guilds {
				emoji := discordutil.FindEmojiInGuild(window.session, guild, false, emojiSequence)
				if emoji != "" {
					asRunes = *mergeRuneSlices(asRunes[:startIndex], []rune(emoji), asRunes[endIndex+1:])
					continue INDEX_LOOP
				}
			}
		}
	}

	return string(asRunes)
}

// ShowDialog shows a dialog at the bottom of the window. It doesn't surrender
// its focus and requires action before allowing the user to proceed. The
// buttons are handled depending on their text.
func (window *Window) ShowDialog(color tcell.Color, text string, buttonHandler func(button string), buttons ...string) {
	window.dialogButtonBar.RemoveAllItems()

	if len(buttons) == 0 {
		return
	}

	previousFocus := window.app.GetFocus()

	buttonWidgets := make([]*tview.Button, 0)
	for index, button := range buttons {
		newButton := tview.NewButton(button)
		newButton.SetSelectedFunc(func() {
			buttonHandler(newButton.GetLabel())
			window.dialogReplacement.SetVisible(false)
			window.app.SetFocus(previousFocus)
		})
		buttonWidgets = append(buttonWidgets, newButton)

		indexCopy := index
		newButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyRight {
				if len(buttonWidgets) <= indexCopy+1 {
					window.app.SetFocus(buttonWidgets[0])
				} else {
					window.app.SetFocus(buttonWidgets[indexCopy+1])
				}
				return nil
			}

			if event.Key() == tcell.KeyLeft {
				if indexCopy == 0 {
					window.app.SetFocus(buttonWidgets[len(buttonWidgets)-1])
				} else {
					window.app.SetFocus(buttonWidgets[indexCopy-1])
				}
				return nil
			}

			return event
		})

		window.dialogButtonBar.AddItem(newButton, len(button)+2, 0, false)
		window.dialogButtonBar.AddItem(tview.NewBox(), 1, 0, false)
	}
	window.dialogButtonBar.AddItem(tview.NewBox(), 0, 1, false)

	window.dialogTextView.SetText(text)
	window.dialogTextView.SetBackgroundColor(color)
	window.dialogReplacement.SetVisible(true)
	window.app.SetFocus(buttonWidgets[0])

	_, _, width, _ := window.rootContainer.GetRect()
	height := tviewutil.CalculateNecessaryHeight(width, window.dialogTextView.GetText(true))
	window.rootContainer.ResizeItem(window.dialogReplacement, height+2, 0)
}

func (window *Window) registerMouseFocusListeners() {
	window.chatView.internalTextView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.chatView.internalTextView)
		} else if event.Buttons() == tcell.WheelDown {
			window.chatView.internalTextView.ScrollDown()
		} else if event.Buttons() == tcell.WheelUp {
			window.chatView.internalTextView.ScrollUp()
		} else {
			return false
		}

		return true
	})

	var lastLeftContainerSwitchTimeMillis int64
	window.guildList.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			if window.activeView != Guilds {
				nowMillis := time.Now().UnixNano() / 1000 / 1000
				//Avoid triggering multiple times in a row due to mouse movement during the click
				if nowMillis-lastLeftContainerSwitchTimeMillis > 60 {
					window.SwitchToGuildsPage()
					window.app.SetFocus(window.guildList)
				}
				lastLeftContainerSwitchTimeMillis = nowMillis
			} else {
				window.app.SetFocus(window.guildList)
			}
			return true
		}

		return false
	})
	window.channelTree.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.channelTree)

			return true
		}

		return false
	})

	window.userList.internalTreeView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.userList.internalTreeView)

			return true
		}

		return false
	})

	window.privateList.internalTreeView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			if window.activeView != Dms {
				nowMillis := time.Now().UnixNano() / 1000 / 1000
				//Avoid triggering multiple times in a row due to mouse movement during the click
				if nowMillis-lastLeftContainerSwitchTimeMillis > 60 {
					window.SwitchToFriendsPage()
					window.app.SetFocus(window.privateList.internalTreeView)
				}
				lastLeftContainerSwitchTimeMillis = nowMillis
			} else {
				window.app.SetFocus(window.privateList.internalTreeView)
			}
			return true
		}

		return false
	})

	window.messageInput.internalTextView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.messageInput.internalTextView)

			return true
		}

		return false
	})

	window.commandView.commandInput.internalTextView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.commandView.commandInput.internalTextView)

			return true
		}

		return false
	})

	window.commandView.commandOutput.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.commandView.commandOutput)
		} else if event.Buttons() == tcell.WheelDown {
			window.commandView.commandOutput.ScrollDown()
		} else if event.Buttons() == tcell.WheelUp {
			window.commandView.commandOutput.ScrollUp()
		} else {
			return false
		}

		return true
	})
}

func (window *Window) registerMessageEventHandler(input, edit, delete chan *discordgo.Message, bulkDelete chan *discordgo.MessageDeleteBulk) {
	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		input <- m.Message
	})
	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
		bulkDelete <- m
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		delete <- m.Message
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		edit <- m.Message
	})
}

// QueueUpdateDrawSynchronized is meant to be used by goroutines that aren't
// the main goroutine in order to wait for the UI-Thread to execute the given
// If this method is ever called from the main thread, the application will
// deadlock.
func (window *Window) QueueUpdateDrawSynchronized(runnable func()) {
	blocker := make(chan bool, 1)
	window.app.QueueUpdateDraw(func() {
		runnable()
		blocker <- true
	})
	<-blocker
	close(blocker)
}

// startMessageHandlerRoutines registers the handlers for certain message
// events. It updates the cache and the UI if necessary.
func (window *Window) startMessageHandlerRoutines(input, edit, delete chan *discordgo.Message, bulkDelete chan *discordgo.MessageDeleteBulk) {
	go func() {
		for tempMessage := range input {
			message := tempMessage
			if len(window.extensionEngines) > 0 {
				go func() {
					for _, engine := range window.extensionEngines {
						engine.OnMessageReceive(message)
					}
				}()
			}

			channel, stateError := window.session.State.Channel(message.ChannelID)
			if stateError != nil {
				continue
			}

			window.chatView.Lock()
			if window.selectedChannel != nil && message.ChannelID == window.selectedChannel.ID {
				if message.Author.ID != window.session.State.User.ID {
					readstate.UpdateReadBuffered(window.session, channel, message.ID)
				}

				window.QueueUpdateDrawSynchronized(func() {
					window.chatView.AddMessage(message)
				})
			}
			window.chatView.Unlock()

			if channel.Type == discordgo.ChannelTypeGuildText && (window.selectedGuild == nil ||
				window.selectedGuild.ID != channel.GuildID) {
				window.app.QueueUpdateDraw(func() {
					window.updateServerReadStatus(channel.GuildID, false)
				})
			}

			// TODO,HACK.FIXME Since the cache is inconsistent, I have to
			// update it myself. This should be moved over into the
			// discordgo code ASAP.
			channel.LastMessageID = message.ID

			if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
				//Avoid unnecessary drawing if the updates wouldn't be visible either way.
				//FIXME Useful to use locking here?
				if window.activeView == Dms {
					window.app.QueueUpdateDraw(func() {
						window.privateList.ReorderChannelList()
					})
				} else {
					window.privateList.ReorderChannelList()
				}
			}

			if message.Author.ID == window.session.State.User.ID {
				readstate.UpdateReadLocal(message.ChannelID, message.ID)
				continue
			}

			if config.Current.DesktopNotifications {
				notifyError := window.handleNotification(message, channel)
				if notifyError != nil {
					log.Printf("["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error sending notification:\n\t[%s]%s\n", tviewutil.ColorToHex(config.GetTheme().ErrorColor), notifyError)
				}
			}

			if window.selectedChannel == nil || message.ChannelID != window.selectedChannel.ID {
				if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
					if !readstate.IsPrivateChannelMuted(channel) {
						window.app.QueueUpdateDraw(func() {
							window.privateList.MarkChannelAsUnread(channel)
						})
					}
				} else if channel.Type == discordgo.ChannelTypeGuildText {
					if discordutil.MentionsCurrentUserExplicitly(window.session.State, message) {
						readstate.MarkAsMentioned(channel.ID)
						window.app.QueueUpdateDraw(func() {
							isCurrentGuild := window.selectedGuild != nil && window.selectedGuild.ID == channel.GuildID
							window.updateServerReadStatus(channel.GuildID, isCurrentGuild)
							window.channelTree.MarkChannelAsMentioned(channel.ID)
						})
					} else if !readstate.IsGuildChannelMuted(channel) {
						window.app.QueueUpdateDraw(func() {
							window.channelTree.MarkChannelAsUnread(channel.ID)
						})
					}
				}
			}
		}
	}()

	go func() {
		for messageDeleted := range delete {
			tempMessageDeleted := messageDeleted

			if len(window.extensionEngines) > 0 {
				go func() {
					for _, engine := range window.extensionEngines {
						engine.OnMessageDelete(tempMessageDeleted)
					}
				}()
			}
			window.chatView.Lock()
			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessageDeleted.ChannelID {
				window.QueueUpdateDrawSynchronized(func() {
					window.chatView.DeleteMessage(tempMessageDeleted)
				})
			}
			window.chatView.Unlock()
		}
	}()

	go func() {
		for messagesDeleted := range bulkDelete {
			tempMessagesDeleted := messagesDeleted
			window.chatView.Lock()
			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessagesDeleted.ChannelID {
				window.QueueUpdateDrawSynchronized(func() {
					window.chatView.DeleteMessages(tempMessagesDeleted.Messages)
				})
			}
			window.chatView.Unlock()
		}
	}()

	go func() {
	MESSAGE_EDIT_LOOP:
		for messageEdited := range edit {
			tempMessageEdited := messageEdited
			if len(window.extensionEngines) > 0 {
				go func() {
					for _, engine := range window.extensionEngines {
						engine.OnMessageEdit(tempMessageEdited)
					}
				}()
			}
			window.chatView.Lock()
			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessageEdited.ChannelID {
				for _, message := range window.chatView.data {
					if message.ID == tempMessageEdited.ID {
						//FIXME Workaround for the fact that discordgo doesn't update already filled fields.

						//FIXME Workaround for the workaround, since discord appears to not send the content
						//again for messages that have only had an embed added. In that situation, the
						//timestamp for editing will also not be set, therefore we can circumvent this issue.
						if tempMessageEdited.EditedTimestamp != "" && tempMessageEdited.Content != "" {
							message.Content = tempMessageEdited.Content
						}

						message.Mentions = tempMessageEdited.Mentions
						message.MentionRoles = tempMessageEdited.MentionRoles
						message.MentionEveryone = tempMessageEdited.MentionEveryone

						window.QueueUpdateDrawSynchronized(func() {
							defer window.chatView.Unlock()
							window.chatView.UpdateMessage(message)
						})
						continue MESSAGE_EDIT_LOOP
					}
				}
			}
			window.chatView.Unlock()
		}
	}()
}

func (window *Window) registerGuildHandlers() {
	//Using buffered channels with a size of three, since this shouldn't really happen often

	guildCreateChannel := make(chan *discordgo.GuildCreate, 3)
	window.session.AddHandler(func(s *discordgo.Session, guildCreate *discordgo.GuildCreate) {
		guildCreateChannel <- guildCreate
	})

	guildRemoveChannel := make(chan *discordgo.GuildDelete, 3)
	window.session.AddHandler(func(s *discordgo.Session, guildRemove *discordgo.GuildDelete) {
		guildRemoveChannel <- guildRemove
	})

	guildUpdateChannel := make(chan *discordgo.GuildUpdate, 3)
	window.session.AddHandler(func(s *discordgo.Session, guildUpdate *discordgo.GuildUpdate) {
		guildUpdateChannel <- guildUpdate
	})

	go func() {
		for guildCreate := range guildCreateChannel {
			guild := guildCreate
			if window.guildList.GetCurrentNode() == nil {
				window.app.QueueUpdateDraw(func() {
					window.guildList.AddGuild(guild.ID, guild.Name)
					window.guildList.SetCurrentNode(window.guildList.GetRoot())
				})
			} else {
				window.app.QueueUpdateDraw(func() {
					window.guildList.AddGuild(guild.ID, guild.Name)
				})
			}
		}
	}()

	go func() {
		for guildUpdate := range guildUpdateChannel {
			guild := guildUpdate
			window.app.QueueUpdateDraw(func() {
				window.guildList.UpdateName(guild.ID, guild.Name)
			})
		}
	}()

	go func() {
		for guildRemove := range guildRemoveChannel {
			if window.selectedGuild == nil {
				continue
			}

			if window.previousGuild != nil && window.previousGuild.ID == guildRemove.ID {
				window.previousGuild = nil
				window.previousChannel = nil
			}

			if window.selectedGuild.ID == guildRemove.ID {
				guildID := guildRemove.ID
				window.app.QueueUpdateDraw(func() {
					if window.selectedChannel != nil && window.selectedChannel.GuildID == guildID {
						window.chatView.ClearViewAndCache()
						window.selectedChannel = nil
					}

					window.channelTree.Clear()
					window.userList.Clear()
					window.guildList.RemoveGuild(guildID)
					window.selectedGuild = nil
				})
			}
		}
	}()
}

func (window *Window) registerGuildMemberHandlers() {
	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildMembersChunk) {
		if window.selectedGuild != nil && window.selectedGuild.ID == event.GuildID {
			window.app.QueueUpdateDraw(func() {
				window.userList.AddOrUpdateMembers(event.Members)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildMemberRemove) {
		if window.selectedGuild != nil && window.selectedGuild.ID == event.GuildID {
			window.app.QueueUpdateDraw(func() {
				window.userList.RemoveMember(event.Member)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
		if window.selectedGuild != nil && window.selectedGuild.ID == event.GuildID {
			window.app.QueueUpdateDraw(func() {
				window.userList.AddOrUpdateMember(event.Member)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildMemberUpdate) {
		if window.selectedGuild != nil && window.selectedGuild.ID == event.GuildID {
			window.app.QueueUpdateDraw(func() {
				window.userList.AddOrUpdateMember(event.Member)
			})
		}
	})
}

func (window *Window) registerPrivateChatsHandler() {
	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelCreate) {
		if event.Type == discordgo.ChannelTypeDM || event.Type == discordgo.ChannelTypeGroupDM {
			window.app.QueueUpdateDraw(func() {
				window.privateList.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelDelete) {
		if event.Type == discordgo.ChannelTypeDM || event.Type == discordgo.ChannelTypeGroupDM {
			window.app.QueueUpdateDraw(func() {
				window.privateList.RemoveChannel(event.Channel)
			})
			readstate.ClearReadStateFor(event.ID)
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelUpdate) {
		if event.Type == discordgo.ChannelTypeDM || event.Type == discordgo.ChannelTypeGroupDM {
			window.app.QueueUpdateDraw(func() {
				window.privateList.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.RelationshipAdd) {
		if event.Relationship.Type == discordgo.RelationTypeFriend {
			window.app.QueueUpdateDraw(func() {
				window.privateList.addFriend(event.User)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.RelationshipRemove) {
		if event.Relationship.Type == discordgo.RelationTypeFriend {
			for _, relationship := range window.session.State.Relationships {
				if relationship.ID == event.ID {
					window.app.QueueUpdateDraw(func() {
						window.privateList.RemoveFriend(relationship.User.ID)
					})
					break
				}
			}
		}
	})
}

func (window *Window) isChannelEventRelevant(channelEvent *discordgo.Channel) bool {
	if window.selectedGuild == nil {
		return false
	}

	if channelEvent.Type != discordgo.ChannelTypeGuildText && channelEvent.Type != discordgo.ChannelTypeGuildCategory {
		return false
	}

	if window.selectedGuild.ID != channelEvent.GuildID {
		return false
	}

	return true
}

func (window *Window) registerGuildChannelHandler() {
	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelCreate) {
		if window.isChannelEventRelevant(event.Channel) {
			window.channelTree.Lock()
			window.QueueUpdateDrawSynchronized(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
			window.channelTree.Unlock()
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelUpdate) {
		if window.isChannelEventRelevant(event.Channel) {
			window.channelTree.Lock()
			window.QueueUpdateDrawSynchronized(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
			window.channelTree.Unlock()
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelDelete) {
		if window.isChannelEventRelevant(event.Channel) {
			if window.previousChannel != nil && window.previousChannel.ID == event.ID {
				window.previousGuild = nil
				window.previousChannel = nil
			}

			if window.selectedChannel != nil && window.selectedChannel.ID == event.ID {
				window.selectedChannel = nil
				window.app.QueueUpdateDraw(func() {
					window.chatView.ClearViewAndCache()
				})
			}

			window.messageLoader.DeleteFromCache(event.Channel.ID)
			//On purpose, since we don't care much about removing the channel timely.
			window.app.QueueUpdateDraw(func() {
				window.channelTree.Lock()
				window.channelTree.RemoveChannel(event.Channel)
				window.channelTree.Unlock()
			})
		}
	})
}

func (window *Window) isElligibleForNotification(message *discordgo.Message, channel *discordgo.Channel) bool {
	if discordutil.IsBlocked(window.session.State, message.Author) {
		return false
	}

	isCurrentChannel := window.selectedChannel == nil || message.ChannelID != window.selectedChannel.ID
	//Client is not in a state elligible for notifications.
	if isCurrentChannel && (window.userActive || !config.Current.DesktopNotificationsForLoadedChannel) {
		return false
	}

	isPrivateChannel := channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM
	mentionsCurrentUser := discordutil.MentionsCurrentUserExplicitly(window.session.State, message)
	//We always show notification for private messages, no matter whether
	//the user was explicitly mentioned.
	if !isPrivateChannel && !mentionsCurrentUser {
		return false
	}

	return true
}

func (window *Window) handleNotification(message *discordgo.Message, channel *discordgo.Channel) error {
	if !window.isElligibleForNotification(message, channel) {
		return nil
	}

	var notificationLocation string
	if channel.Type == discordgo.ChannelTypeDM {
		notificationLocation = message.Author.Username
	} else if channel.Type == discordgo.ChannelTypeGroupDM {
		notificationLocation = channel.Name
		if notificationLocation == "" {
			for index, recipient := range channel.Recipients {
				if index == 0 {
					notificationLocation = recipient.Username
				} else {
					notificationLocation = fmt.Sprintf("%s, %s", notificationLocation, recipient.Username)
				}
			}
		}

		notificationLocation = message.Author.Username + " - " + notificationLocation
	} else if channel.Type == discordgo.ChannelTypeGuildText {
		guild, cacheError := window.session.State.Guild(message.GuildID)
		if guild != nil && cacheError == nil {
			notificationLocation = fmt.Sprintf("%s - %s - %s", guild.Name, channel.Name, message.Author.Username)
		} else {
			notificationLocation = fmt.Sprintf("%s - %s", message.Author.Username, channel.Name)
		}
	}

	return beeep.Notify("Cordless - "+notificationLocation, message.ContentWithMentionsReplaced(), "assets/information.png")
}

func (window *Window) askForMessageDeletion(messageID string, usedWithSelection bool) {
	deleteButtonText := "Delete"
	window.ShowDialog(tview.Styles.PrimitiveBackgroundColor,
		"Do you really want to delete the message?", func(button string) {
			if button == deleteButtonText {
				go window.session.ChannelMessageDelete(window.selectedChannel.ID, messageID)
			}

			window.exitMessageEditMode()
			if usedWithSelection {
				window.chatView.SignalSelectionDeleted()
			}
		}, deleteButtonText, "Abort")
}

// SetCommandModeEnabled hides or shows the command ui elements and toggles
// the commandMode flag.
func (window *Window) SetCommandModeEnabled(enabled bool) {
	if window.commandMode != enabled {
		window.commandMode = enabled
		window.commandView.SetVisible(enabled)
	}
}

func (window *Window) handleGlobalShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if shortcuts.ExitApplication.Equals(event) {
		window.doRestart <- false
		window.app.Stop()
		return nil
	}

	// Maybe compare directly to table?
	if config.Current.DesktopNotificationsUserInactivityThreshold > 0 {
		window.userActive = true
		window.userActiveTimer.Reset(time.Duration(config.Current.DesktopNotificationsUserInactivityThreshold) * time.Second)
	}

	if shortcuts.ToggleBareChat.Equals(event) {
		window.toggleBareChat()
		return nil
	}

	//This two have to work in baremode as well, since otherwise only the mouse
	//can be used for focus switching, which sucks in a terminal app.
	if shortcuts.FocusMessageInput.Equals(event) {
		window.app.SetFocus(window.messageInput.GetPrimitive())
		return nil
	}

	if shortcuts.FocusMessageContainer.Equals(event) {
		window.app.SetFocus(window.chatView.internalTextView)
		return nil
	}

	if window.app.GetRoot() != window.rootContainer {
		return event
	}

	if window.dialogReplacement.IsVisible() {
		return event
	}

	if shortcuts.EventsEqual(event, shortcutsDialogShortcut) {
		shortcutdialog.ShowShortcutsDialog(window.app, func() {
			window.app.SetRoot(window.rootContainer, true)
			window.app.ForceDraw()
		})
	} else if shortcuts.ToggleCommandView.Equals(event) {
		window.SetCommandModeEnabled(!window.commandMode)

		if window.commandMode {
			window.app.SetFocus(window.commandView.commandInput.internalTextView)
		} else {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}
	} else if shortcuts.FocusCommandOutput.Equals(event) {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandOutput)
	} else if shortcuts.FocusCommandInput.Equals(event) {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandInput.internalTextView)
	} else if shortcuts.ToggleUserContainer.Equals(event) {
		window.toggleUserContainer()
	} else if shortcuts.FocusChannelContainer.Equals(event) {
		window.SwitchToGuildsPage()
		window.app.SetFocus(window.channelTree)
	} else if shortcuts.FocusPrivateChatPage.Equals(event) {
		window.SwitchToFriendsPage()
		window.app.SetFocus(window.privateList.GetComponent())
	} else if shortcuts.SwitchToPreviousChannel.Equals(event) {
		err := window.SwitchToPreviousChannel()
		if err != nil {
			window.ShowErrorDialog(err.Error())
		}
	} else if shortcuts.FocusGuildContainer.Equals(event) {
		window.SwitchToGuildsPage()
		window.app.SetFocus(window.guildList)
	} else if shortcuts.FocusUserContainer.Equals(event) {
		if window.activeView == Guilds && window.userList.internalTreeView.IsVisible() {
			window.app.SetFocus(window.userList.internalTreeView)
		}
	} else {
		return event
	}

	return nil
}

func (window *Window) toggleUserContainer() {
	config.Current.ShowUserContainer = !config.Current.ShowUserContainer

	if !config.Current.ShowUserContainer && window.app.GetFocus() == window.userList.internalTreeView {
		window.app.SetFocus(window.messageInput.GetPrimitive())
	}

	if config.Current.ShowUserContainer {
		if !window.userList.IsLoaded() {
			if window.selectedChannel != nil && window.selectedChannel.GuildID == "" {
				window.userList.LoadGroup(window.selectedChannel.ID)
			} else if window.selectedGuild != nil {
				window.userList.LoadGuild(window.selectedGuild.ID)
			}
		}
	} else {
		window.userList.Clear()
	}

	config.PersistConfig()
	window.RefreshLayout()
}

// toggleBareChat will display only the chatview as the fullscreen application
// root. Calling this method again will revert the view to it's normal state.
func (window *Window) toggleBareChat() {
	window.bareChat = !window.bareChat
	if window.bareChat {
		window.chatView.internalTextView.SetBorderSides(true, false, true, false)
		previousFocus := window.app.GetFocus()
		//Initially this should be gone. Maybe we'll allow reacessing it at some point.
		window.SetCommandModeEnabled(false)
		window.app.SetRoot(window.chatArea, true)
		window.app.SetFocus(previousFocus)
	} else {
		window.chatView.internalTextView.SetBorderSides(true, true, true, true)
		window.app.SetRoot(window.rootContainer, true)
		window.app.SetFocus(window.messageInput.GetPrimitive())
	}

	window.app.QueueUpdateDraw(func() {
		window.messageInput.TriggerHeightRequestIfNecessary()
		window.chatView.Reprint()
	})
}

// FindCommand searches through the registered command, whether any of them
// equals the passed name.
func (window *Window) FindCommand(name string) commands.Command {
	for _, cmd := range window.commands {
		if commands.CommandEquals(cmd, name) {
			return cmd
		}
	}

	return nil
}

//ExecuteCommand tries to execute the given input as a command. The first word
//will be passed as the commands name and the rest will be parameters. If a
//command can't be found, that info will be printed onto the command output.
func (window *Window) ExecuteCommand(input string) {
	parts := commands.ParseCommand(input)
	fmt.Fprintf(window.commandView, "[gray]$ %s\n", input)

	if len(parts) > 0 {
		command := window.FindCommand(parts[0])
		if command != nil {
			command.Execute(window.commandView, parts[1:])
		} else {
			fmt.Fprintf(window.commandView, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]The command '%s' doesn't exist[white]\n", parts[0])
		}
	}
}

// ShowTFASetup generates a new TFA-Secret and shows a QR-Code. The QR-Code can
// be scanned and the resulting TFA-Token can be entered into cordless and used
// to enable TFA on this account.
func (window *Window) ShowTFASetup() error {
	tfaSecret, secretError := text.GenerateBase32Key()
	if secretError != nil {
		return secretError
	}

	qrURL := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=Discord", window.session.State.User.Email, tfaSecret)
	qrCodeText := text.GenerateQRCode(qrURL, qrterminal.M)
	qrCodeImage := tview.NewTextView().SetText(qrCodeText).SetTextAlign(tview.AlignCenter)
	qrCodeImage.SetTextColor(tcell.ColorWhite).SetBackgroundColor(tcell.ColorBlack)
	qrCodeView := tview.NewFlex().SetDirection(tview.FlexRow)
	width := len([]rune(strings.TrimSpace(strings.Split(qrCodeText, "\n")[2])))
	qrCodeView.AddItem(tviewutil.CreateCenteredComponent(qrCodeImage, width), strings.Count(qrCodeText, "\n"), 0, false)
	humanReadableSecret := tfaSecret[:4] + " " + tfaSecret[4:8] + " " + tfaSecret[8:12] + " " + tfaSecret[12:16]
	defaultInstructions := "1. Scan the QR-Code with your 2FA application\n   or enter the secret manually:\n     " +
		humanReadableSecret + "\n2. Enter the code generated on your 2FA device\n3. Hit Enter!"
	message := tview.NewTextView().SetText(defaultInstructions).SetDynamicColors(true)
	qrCodeView.AddItem(tviewutil.CreateCenteredComponent(message, 68), 0, 1, false)
	tokenInput := tview.NewInputField()
	tokenInput.SetBorder(true)
	tokenInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			code, codeError := text.ParseTFACode(tokenInput.GetText())
			if codeError != nil {
				message.SetText(fmt.Sprintf("%s\n\n[red]Code invalid:\n\t[red]%s", defaultInstructions, codeError))
				return nil
			}

			//panic(fmt.Sprintf("Secret: %s\nCode: %s", tfaSecret, code))
			backupCodes, tfaError := window.session.TwoFactorEnable(tfaSecret, code)
			if tfaError != nil {
				message.SetText(fmt.Sprintf("%s\n\n[red]Error setting up Two-Factor-Authentication:\n\t[red]%s", defaultInstructions, tfaError))
				return nil
			}

			//The token is being updated internally, therefore we need to update our config.
			config.UpdateCurrentToken(window.session.Token)
			configError := config.PersistConfig()
			if configError != nil {
				log.Println(fmt.Sprintf("Error settings new token: %s\n\t%s", window.session.Token, configError))
			}

			var backupCodesAsString string
			for index, backupCode := range backupCodes {
				if index != 0 {
					backupCodesAsString += "\n"
				}
				backupCodesAsString += backupCode.Code
			}

			clipboard.WriteAll(backupCodesAsString)

			successText := tview.NewTextView().SetTextAlign(tview.AlignCenter)
			successText.SetText("Setting up Two-Factor-Authentication was a success.\n\n" +
				"The backup codes have been put into your clipboard." +
				"If you need to view your backup codes again, just run `tfa backup` in the cordless CLI.\n\n" +
				"Currently cordless doesn't support applying backup codes.")

			successView := tview.NewFlex().SetDirection(tview.FlexRow)

			okayButton := tview.NewButton("Okay")
			okayButton.SetSelectedFunc(func() {
				window.app.SetRoot(window.rootContainer, true)
			})

			successView.AddItem(successText, 0, 1, false)
			successView.AddItem(okayButton, 1, 0, false)
			window.app.SetRoot(tviewutil.CreateCenteredComponent(successView, 68), true)
			window.app.SetFocus(okayButton)

			return nil
		}

		if event.Key() == tcell.KeyESC {
			window.app.SetRoot(window.rootContainer, true)
			window.app.ForceDraw()
			return nil
		}

		return event
	})
	qrCodeView.AddItem(tviewutil.CreateCenteredComponent(tokenInput, 68), 3, 0, false)
	window.app.SetRoot(qrCodeView, true)
	window.app.SetFocus(tokenInput)

	return nil
}

func (window *Window) startEditingMessage(message *discordgo.Message) {
	if message.Author.ID == window.session.State.User.ID {
		window.messageInput.SetText(message.Content)
		window.messageInput.SetBorderColor(tcell.ColorYellow)
		window.messageInput.SetBorderFocusColor(tcell.ColorYellow)
		if vtxxx {
			window.messageInput.SetBorderFocusAttributes(tcell.AttrBlink | tcell.AttrBold)
		}
		window.editingMessageID = &message.ID
		window.app.SetFocus(window.messageInput.GetPrimitive())
	}
}

func (window *Window) exitMessageEditMode() {
	window.exitMessageEditModeAndKeepText()
	window.messageInput.SetText("")
}

func (window *Window) exitMessageEditModeAndKeepText() {
	window.editingMessageID = nil
	window.messageInput.SetBorderColor(tview.Styles.BorderColor)
	window.messageInput.SetBorderFocusColor(tview.Styles.BorderFocusColor)
	if vtxxx {
		window.messageInput.SetBorderFocusAttributes(tcell.AttrBold)
		window.messageInput.SetBorderAttributes(tcell.AttrNone)
	}
}

// ShowErrorDialog shows a simple error dialog that has only an Okay button,
// a generic title and the given text.
func (window *Window) ShowErrorDialog(text string) {
	window.ShowDialog(config.GetTheme().ErrorColor, "An error occurred - "+text, func(_ string) {}, "Okay")
}

// ShowCustomErrorDialog shows a simple error dialog with a custom title
// and text. The button says "Okay" and only closes the dialog.
func (window *Window) ShowCustomErrorDialog(title, text string) {
	window.ShowDialog(config.GetTheme().ErrorColor, title+" - "+text, func(_ string) {}, "Okay")
}

func (window *Window) editMessage(channelID, messageID, messageEdited string) {
	go func() {
		window.app.QueueUpdateDraw(func() {
			window.exitMessageEditMode()
			window.messageInput.SetText("")
		})
		_, discordError := window.session.ChannelMessageEdit(channelID, messageID, messageEdited)
		window.app.QueueUpdateDraw(func() {
			if discordError != nil {
				retry := "Retry sending"
				edit := "Edit"
				cancel := "Cancel"
				window.ShowDialog(config.GetTheme().ErrorColor,
					fmt.Sprintf("Error editing message: %s.\n\nWhat do you want to do?", discordError),
					func(button string) {
						switch button {
						case retry:
							window.editMessage(channelID, messageID, messageEdited)
						case edit:
							window.messageInput.SetText(messageEdited)
						}
					}, retry, edit, cancel)
			}
		})
	}()
}

//SwitchToGuildsPage the left side of the layout over to the view where you can
//see the servers and their channels. In additional to that, it also shows the
//user list in case the user didn't explicitly hide it.
func (window *Window) SwitchToGuildsPage() {
	window.leftArea.RemoveAllItems()
	window.leftArea.AddItem(window.privateList.GetComponent(), 1, 0, false)
	window.leftArea.AddItem(window.guildPage, 0, 1, false)
	window.activeView = Guilds

	window.userList.internalTreeView.SetNextFocusableComponents(tview.Right, window.guildList)
	window.chatView.internalTextView.SetNextFocusableComponents(tview.Left, window.guildList)
	window.chatView.internalTextView.SetNextFocusableComponents(tview.Right, window.userList.internalTreeView, window.guildList)
}

//SwitchToFriendsPage switches the left side of the layout over to the view
//where you can see your private chats and groups. In addition to that it
//hides the user list.
func (window *Window) SwitchToFriendsPage() {
	window.leftArea.RemoveAllItems()
	window.leftArea.AddItem(window.guildList, 1, 0, false)
	window.leftArea.AddItem(window.privateList.GetComponent(), 0, 1, false)
	window.activeView = Dms

	window.userList.internalTreeView.SetNextFocusableComponents(tview.Right, window.privateList.internalTreeView)
	window.chatView.internalTextView.SetNextFocusableComponents(tview.Left, window.privateList.internalTreeView)
	window.chatView.internalTextView.SetNextFocusableComponents(tview.Right, window.userList.internalTreeView, window.privateList.internalTreeView)
}

// SwitchToPreviousChannel loads the previously loaded channel and focuses it
// in it's respective UI primitive.
func (window *Window) SwitchToPreviousChannel() error {
	if window.previousChannel == nil || window.previousChannel == window.selectedChannel {
		// No previous channel.
		return nil
	}

	_, err := window.session.State.Channel(window.previousChannel.ID)
	if err != nil {
		window.previousChannel = nil
		return fmt.Errorf("Channel %s not found", window.previousChannel.Name)
	}

	// Switch to appropriate layout.
	switch window.previousChannel.Type {
	case discordgo.ChannelTypeDM, discordgo.ChannelTypeGroupDM:
		window.SwitchToFriendsPage()
		window.privateList.onChannelSelect(window.previousChannel.ID)
	case discordgo.ChannelTypeGuildText:
		_, err := window.session.State.Guild(window.previousGuild.ID)
		if err != nil {
			window.previousGuild = nil
			return fmt.Errorf("Unable to load guild: %s", window.previousGuild.Name)
		}
		if !discordutil.HasReadMessagesPermission(window.previousChannel.ID, window.session.State) {
			return fmt.Errorf("No read permissions for channel: %s", window.previousChannel.Name)
		}
		window.SwitchToGuildsPage()
		previousGuildNode := tviewutil.GetNodeByReference(window.previousGuild.ID, window.guildList.TreeView)
		previousChannelNode := tviewutil.GetNodeByReference(window.previousChannel.ID, window.channelTree.TreeView)
		window.guildList.SetCurrentNode(previousGuildNode)
		window.guildList.onGuildSelect(window.previousGuild.ID)
		window.channelTree.SetCurrentNode(previousChannelNode)
		window.channelTree.onChannelSelect(window.previousChannel.ID)
	default:
		return fmt.Errorf("Invalid channel type: %v", window.previousChannel.Type)
	}
	window.app.SetFocus(window.messageInput.internalTextView)
	return nil
}

//RefreshLayout removes and adds the main parts of the layout
//so that the ones that are disabled by settings do not show up.
func (window *Window) RefreshLayout() {
	window.userList.internalTreeView.SetVisible(config.Current.ShowUserContainer && (window.selectedGuild != nil ||
		(window.selectedChannel != nil && window.selectedChannel.Type == discordgo.ChannelTypeGroupDM)))

	if config.Current.UseFixedLayout {
		window.middleContainer.ResizeItem(window.leftArea, config.Current.FixedSizeLeft, 0)
		window.middleContainer.ResizeItem(window.chatArea, 0, 1)
		window.middleContainer.ResizeItem(window.userList.internalTreeView, config.Current.FixedSizeRight, 0)
	} else {
		window.middleContainer.ResizeItem(window.leftArea, 0, 7)
		window.middleContainer.ResizeItem(window.chatArea, 0, 20)
		window.middleContainer.ResizeItem(window.userList.internalTreeView, 0, 6)
	}

	window.app.ForceDraw()
}

//LoadChannel eagerly loads the channels messages.
func (window *Window) LoadChannel(channel *discordgo.Channel) error {
	messages, loadError := window.messageLoader.LoadMessages(channel)
	if loadError != nil {
		return loadError
	}

	discordutil.SortMessagesByTimestamp(messages)

	window.chatView.SetMessages(messages)
	window.chatView.ClearSelection()
	window.chatView.internalTextView.ScrollToEnd()

	window.UpdateChatHeader(channel)

	if window.selectedChannel == nil {
		window.previousChannel = channel
	} else if channel != window.selectedChannel {
		window.previousChannel = window.selectedChannel

		// When switching to a channel in the same guild, the previousGuild must be set.
		if window.previousChannel.GuildID == channel.GuildID {
			window.previousGuild = window.selectedGuild
		}
	}

	//If there is a currently loaded guild channel and it isn't the same as
	//the new one we assume it must be read and mark it white.
	if window.selectedChannel != nil && channel.ID != window.selectedChannel.ID {
		//FIXME Designflaw! We need to manually reset the primary text
		//color of the selected channels, this really sucks.
		selectedChannelID := window.selectedChannel.ID
		selectedChannelNode := tviewutil.GetNodeByReference(selectedChannelID, window.channelTree.TreeView)
		if selectedChannelNode == nil {
			selectedChannelNode = tviewutil.GetNodeByReference(selectedChannelID, window.privateList.internalTreeView)
		}

		if selectedChannelNode != nil {
			selectedChannelNode.SetColor(tview.Styles.PrimaryTextColor)
		}
	}

	window.selectedChannel = channel

	//Unlike with the channel, where we can assume it is read, we gotta check
	//whether there is still an unread channel and mark the server accordingly.
	wasSelectedGuild := window.selectedGuild != nil && window.selectedGuild.ID == channel.GuildID

	if channel.GuildID == "" {
		window.selectedGuild = nil
	}

	if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
		window.privateList.MarkChannelAsLoaded(channel)
	}

	window.exitMessageEditModeAndKeepText()

	if config.Current.FocusMessageInputAfterChannelSelection {
		window.app.SetFocus(window.messageInput.internalTextView)
	}

	go func() {
		readstate.UpdateRead(window.session, channel, channel.LastMessageID)
		// Here we make the assumption that the channel we are loading must be part
		// of the currently loaded guild, since we don't allow loading a channel of
		// a guild otherwise.
		if channel.GuildID != "" {
			guild, cacheError := window.session.State.Guild(channel.GuildID)

			window.app.QueueUpdateDraw(func() {
				window.updateServerReadStatus(channel.GuildID, wasSelectedGuild)

				if cacheError == nil {
					window.selectedGuild = guild
					for _, guildNode := range window.guildList.GetRoot().GetChildren() {
						if guildNode.GetReference() == channel.GuildID {
							window.guildList.SetCurrentNode(guildNode)
							if vtxxx {
								guildNode.SetAttributes(tcell.AttrUnderline)
							} else {
								guildNode.SetColor(tview.Styles.ContrastBackgroundColor)
							}
							break
						}
					}
				}
			})
		}
	}()

	return nil
}

// UpdateChatHeader updates the bordertitle of the chatviews container.o
// The title consist of the channel name and its topic for guild channels.
// For private channels it's either the recipient in a dm, or all recipients
// in a group dm channel. If the channel has a nickname, that is chosen.
func (window *Window) UpdateChatHeader(channel *discordgo.Channel) {
	if channel == nil {
		return
	}

	if channel.Type == discordgo.ChannelTypeGuildText {
		if channel.Topic != "" {
			window.chatView.SetTitle(channel.Name + " - " + channel.Topic)
		} else {
			window.chatView.SetTitle(channel.Name)
		}
	} else if channel.Type == discordgo.ChannelTypeDM {
		window.chatView.SetTitle(channel.Recipients[0].Username)
	} else {
		window.chatView.SetTitle(discordutil.GetPrivateChannelName(channel))
	}
}

// RegisterCommand register a command. That makes the command available for
// being called from the message input field, in case the user-defined prefix
// is in front of the input.
func (window *Window) RegisterCommand(command commands.Command) {
	window.commands = append(window.commands, command)
}

// GetRegisteredCommands returns the map of all registered commands.
func (window *Window) GetRegisteredCommands() []commands.Command {
	return window.commands
}

// GetSelectedGuild returns a reference to the currently selected Guild.
func (window *Window) GetSelectedGuild() *discordgo.Guild {
	return window.selectedGuild
}

// GetSelectedChannel returns a reference to the currently selected Channel.
func (window *Window) GetSelectedChannel() *discordgo.Channel {
	return window.selectedChannel
}

// PromptSecretInput shows a fullscreen input dialog that masks the user input.
// The returned value will either be empty or what the user has entered.
func (window *Window) PromptSecretInput(title, message string) string {
	return tviewutil.PrompSecretSingleLineInput(window.app, title, message)
}

// ForceRedraw triggers ForceDraw on the underlying tview application, causing
// it to redraw all currently shown components.
func (window *Window) ForceRedraw() {
	window.app.ForceDraw()
}

//Run Shows the window optionally returning an error.
func (window *Window) Run() error {
	return window.app.Run()
}

// Shutdown disconnects from the discord API and stops the tview application.
func (window *Window) Shutdown() {
	if config.Current.ShortenLinks {
		window.chatView.shortener.Close()
	}
	window.session.Close()
	window.app.Stop()
}
