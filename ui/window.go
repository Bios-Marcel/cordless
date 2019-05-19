package ui

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Bios-Marcel/goclipimg"

	"github.com/atotto/clipboard"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/maths"
	"github.com/Bios-Marcel/cordless/scripting"
	"github.com/Bios-Marcel/cordless/scripting/js"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/times"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/discordemojimap"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
	"github.com/gen2brain/beeep"
)

const (
	guildPageName    = "Guilds"
	privatePageName  = "Private"
	userInactiveTime = 10 * time.Second
)

var (
	emojiRegex = regexp.MustCompile(":.+:")
)

// Window is basically the whole application, as it contains all the
// components and the necccessary global state.
type Window struct {
	app               *tview.Application
	middleContainer   *tview.Flex
	rootContainer     *tview.Flex
	dialogReplacement *tview.Flex
	dialogButtonBar   *tview.Flex
	dialogTextView    *tview.TextView
	currentContainer  tview.Primitive

	leftArea    *tview.Pages
	guildList   *tview.TreeView
	channelTree *ChannelTree
	privateList *PrivateChatList

	chatArea         *tview.Flex
	chatView         *ChatView
	messageContainer tview.Primitive
	messageInput     *Editor

	editingMessageID *string

	userList *UserTree

	session *discordgo.Session

	selectedGuildNode   *tview.TreeNode
	selectedGuild       *discordgo.UserGuild
	selectedChannelNode *tview.TreeNode
	selectedChannel     *discordgo.Channel

	jsEngine scripting.Engine

	commandMode bool
	commandView *CommandView
	commands    map[string]commands.Command

	userActive      bool
	userActiveTimer *time.Timer

	doRestart chan bool
}

//NewWindow constructs the whole application window and also registers all
//necessary handlers and functions. If this function returns an error, we can't
//start the application.
func NewWindow(doRestart chan bool, app *tview.Application, session *discordgo.Session) (*Window, error) {
	window := Window{
		doRestart:       doRestart,
		session:         session,
		app:             app,
		commands:        make(map[string]commands.Command),
		jsEngine:        js.New(),
		userActiveTimer: time.NewTimer(userInactiveTime),
	}

	go func() {
		for {
			<-window.userActiveTimer.C
			window.userActive = false
		}
	}()

	window.commandView = NewCommandView(window.ExecuteCommand)

	window.jsEngine.SetErrorOutput(window.commandView.commandOutput)
	if err := window.jsEngine.LoadScripts(config.GetScriptDirectory()); err != nil {
		return nil, err
	}

	guilds, discordError := discordutil.LoadGuilds(window.session)
	if discordError != nil {
		return nil, discordError
	}

	mentionWindowRootNode := tview.NewTreeNode("")
	mentionWindow := tview.NewTreeView().
		SetVimBindingsEnabled(false).
		SetRoot(mentionWindowRootNode).
		SetTopLevel(1).
		SetCycleSelection(true)
	mentionWindow.SetBorder(true)
	mentionWindow.SetBorderSides(false, true, false, true)

	window.leftArea = tview.NewPages()

	guildPage := tview.NewFlex()
	guildPage.SetDirection(tview.FlexRow)

	channelTree := NewChannelTree(window.session.State)
	window.channelTree = channelTree
	channelTree.SetOnChannelSelect(func(channelID string) {
		channel, cacheError := window.session.State.Channel(channelID)
		if cacheError == nil {
			loadError := window.LoadChannel(channel)
			if loadError == nil {
				channelTree.MarkChannelAsLoaded(channelID)
			}
		}
	})
	window.registerGuildChannelHandler()

	guildList := tview.NewTreeView().
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true)
	window.guildList = guildList

	guildRootNode := tview.NewTreeNode("")
	guildList.SetRoot(guildRootNode)
	guildList.SetBorder(true)
	guildList.SetTopLevel(1)

	window.registerGuildMemberHandlers()

	discordutil.SortGuilds(window.session.State.Settings, guilds)

	for _, tempGuild := range guilds {
		guild := tempGuild
		guildNode := tview.NewTreeNode(guild.Name)
		guildRootNode.AddChild(guildNode)
		guildNode.SetSelectable(true)
		guildNode.SetSelectedFunc(func() {
			if window.selectedGuildNode != nil {
				window.selectedGuildNode.SetColor(tcell.ColorWhite)
			}

			window.selectedGuildNode = guildNode
			window.selectedGuildNode.SetColor(tview.Styles.ContrastBackgroundColor)

			window.selectedGuild = guild
			requestError := session.RequestGuildMembers(guild.ID, "", 0)
			if requestError != nil {
				fmt.Fprintln(window.commandView, "Error retrieving all guild members.")
			}

			channelLoadError := window.channelTree.LoadGuild(guild.ID)
			if channelLoadError != nil {
				window.ShowErrorDialog(channelLoadError.Error())
			} else {
				if config.GetConfig().FocusChannelAfterGuildSelection {
					app.SetFocus(window.channelTree.internalTreeView)
				}
			}

			userLoadError := window.userList.LoadGuild(guild.ID)
			if userLoadError != nil {
				window.ShowErrorDialog(userLoadError.Error())
			}

			window.RefreshLayout()
		})
	}

	if len(guildRootNode.GetChildren()) > 0 {
		guildList.SetCurrentNode(guildRootNode)
	}

	guildPage.AddItem(guildList, 0, 1, true)
	guildPage.AddItem(channelTree.internalTreeView, 0, 2, true)

	window.leftArea.AddPage(guildPageName, guildPage, true, false)

	window.privateList = NewPrivateChatList(window.session.State)
	window.privateList.Load()
	window.registerPrivateChatsHandler()

	window.leftArea.AddPage(privatePageName, window.privateList.GetComponent(), true, false)

	window.privateList.SetOnChannelSelect(func(node *tview.TreeNode, channelID string) {
		channel, stateError := window.session.State.Channel(channelID)
		if stateError != nil {
			window.ShowErrorDialog(fmt.Sprintf("Error loading chat: %s", stateError.Error()))
			return
		}

		window.LoadChannel(channel)
		window.UpdateChatHeader(channel)
		if channel.Type == discordgo.ChannelTypeGroupDM {
			loadError := window.userList.LoadGroup(channel.ID)
			if loadError != nil {
				fmt.Fprintln(window.commandView.commandOutput, "Error loading users for channel.")
			}
		}

		window.RefreshLayout()
	})

	window.privateList.SetOnFriendSelect(func(userID string) {
		userChannels, _ := window.session.UserChannels()
		for _, userChannel := range userChannels {
			if userChannel.Type == discordgo.ChannelTypeDM &&
				(userChannel.Recipients[0].ID == userID) {
				window.LoadChannel(userChannel)
				window.UpdateChatHeader(userChannel)
				window.RefreshLayout()
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
			window.LoadChannel(newChannel)
			window.UpdateChatHeader(newChannel)
			window.RefreshLayout()
		}
	})

	window.chatArea = tview.NewFlex().
		SetDirection(tview.FlexRow)

	window.chatView = NewChatView(window.session, window.session.State.User.ID)
	window.chatView.SetOnMessageAction(func(message *discordgo.Message, event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModNone {
			if event.Rune() == 'q' {
				time, parseError := message.Timestamp.Parse()
				if parseError == nil {
					//TODO Username doesn't take Nicknames into consideration.
					window.messageInput.SetText(fmt.Sprintf(">%s %s: %s\n\n", times.TimeToString(&time), message.Author.Username, message.ContentWithMentionsReplaced()))
					app.SetFocus(window.messageInput.GetPrimitive())
				}
				return nil
			}

			if event.Rune() == 'r' {
				window.messageInput.SetText("@" + message.Author.Username + "#" + message.Author.Discriminator + " " + window.messageInput.GetText())
				app.SetFocus(window.messageInput.GetPrimitive())
				return nil
			}

			if event.Rune() == 'l' {
				copyError := clipboard.WriteAll(fmt.Sprintf("<https://discordapp.com/channels/@me/%s/%s>", message.ChannelID, message.ID))
				if copyError != nil {
					window.ShowErrorDialog(fmt.Sprintf("Error copying message link: %s", copyError.Error()))
				}
			}

			if event.Key() == tcell.KeyDelete {
				if message.Author.ID == window.session.State.User.ID {
					window.askForMessageDeletion(message.ID, true)
				}
				return nil
			}

			if event.Rune() == 'e' {
				window.startEditingMessage(message)
				return nil
			}

			if event.Rune() == 'c' {
				copyError := clipboard.WriteAll(message.ContentWithMentionsReplaced())
				if copyError != nil {
					window.ShowErrorDialog(fmt.Sprintf("Error copying message: %s", copyError.Error()))
				}
				return nil
			}
		}

		return event
	})
	window.messageContainer = window.chatView.GetPrimitive()

	window.messageInput = NewEditor()
	window.messageInput.SetOnHeightChangeRequest(func(height int) {
		window.chatArea.ResizeItem(window.messageInput.GetPrimitive(), maths.Min(height, 20), 0)
	})

	window.messageInput.SetMentionShowHandler(func(namePart string) {
		mentionWindow.GetRoot().ClearChildren()
		window.commandView.commandOutput.Clear()

		if window.selectedChannel != nil {
			if window.selectedGuild != nil {
				guild, discordError := window.session.State.Guild(window.selectedGuild.ID)
				if discordError == nil {
					for _, user := range guild.Members {
						if strings.Contains(strings.ToUpper(user.Nick), strings.ToUpper(namePart)) || strings.Contains(strings.ToUpper(user.User.Username)+"#"+user.User.Discriminator, strings.ToUpper(namePart)) {
							userName := user.User.Username + "#" + user.User.Discriminator
							userNodeText := "\t" + userName
							if len(user.Nick) > 0 {
								userNodeText += " | " + user.Nick
							}
							userNode := tview.NewTreeNode(userNodeText)
							userNode.SetReference(userName)
							mentionWindow.GetRoot().AddChild(userNode)
						}
					}

					for _, role := range guild.Roles {
						if strings.Contains(strings.ToUpper(role.Name), strings.ToUpper(namePart)) {
							roleNode := tview.NewTreeNode(role.Name)
							roleNode.SetReference(role)
							mentionWindow.GetRoot().AddChild(roleNode)
						}
					}
				}
			} else {
				for _, user := range window.selectedChannel.Recipients {
					if strings.Contains(strings.ToUpper(user.Username)+"#"+user.Discriminator, strings.ToUpper(namePart)) {
						userName := user.Username + "#" + user.Discriminator
						userNodeText := "\t" + userName
						userNode := tview.NewTreeNode(userNodeText)
						userNode.SetReference(userName)
						mentionWindow.GetRoot().AddChild(userNode)
					}
				}
			}
		}

		if mentionWindow.GetRoot().GetChildren() != nil {
			numChildren := len(mentionWindow.GetRoot().GetChildren())
			if numChildren > 10 {
				numChildren = 10
			}
			window.chatArea.ResizeItem(mentionWindow, numChildren, 0)
			if numChildren > 0 {
				mentionWindow.SetCurrentNode(mentionWindow.GetRoot().GetChildren()[0])
			}
		}
		mentionWindow.SetVisible(mentionWindow.GetRoot().GetChildren() != nil)
		window.app.SetFocus(mentionWindow)
	})

	window.messageInput.SetMentionHideHandler(func() {
		mentionWindow.SetVisible(false)
		window.app.SetFocus(window.messageInput.GetPrimitive())
	})

	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		messageToSend := window.messageInput.GetText()

		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
				window.app.SetFocus(window.chatView.internalTextView)
				return nil
			}

			if event.Key() == tcell.KeyRight {
				if window.userList.internalTreeView.IsVisible() {
					window.app.SetFocus(window.userList.internalTreeView)
				} else {
					if window.leftArea.GetCurrentPage() == guildPageName {
						window.app.SetFocus(window.channelTree.internalTreeView)
						return nil
					} else if window.leftArea.GetCurrentPage() == privatePageName {
						window.app.SetFocus(window.privateList.internalTreeView)
						return nil
					}
				}
				return nil
			}

			if event.Key() == tcell.KeyLeft {
				if window.leftArea.GetCurrentPage() == guildPageName {
					window.app.SetFocus(window.channelTree.internalTreeView)
					return nil
				} else if window.leftArea.GetCurrentPage() == privatePageName {
					window.app.SetFocus(window.privateList.internalTreeView)
					return nil
				}
			}
		}

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

		if event.Key() == tcell.KeyUp && messageToSend == "" {
			for i := len(window.chatView.data) - 1; i >= 0; i-- {
				message := window.chatView.data[i]
				if message.Author.ID == window.session.State.User.ID {
					window.startEditingMessage(message)
					break
				}
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
				currentText := window.messageInput.GetText()
				if currentText == "" {
					go window.session.ChannelFileSend(window.selectedChannel.ID, "img.png", dataChannel)
				} else {
					go window.session.ChannelFileSendWithMessage(window.selectedChannel.ID, currentText, "img.png", dataChannel)
					window.messageInput.SetText("")
				}
			} else {
				window.ShowErrorDialog(fmt.Sprintf("Error pasting image: %s", clipError.Error()))
			}

			return nil
		}

		if event.Key() == tcell.KeyEnter {
			if window.selectedChannel != nil {
				window.messageInput.SetText("")

				if len(messageToSend) != 0 {
					messageToSend = codeBlockRegex.ReplaceAllStringFunc(messageToSend, func(input string) string {
						return strings.Replace(input, ":", "\\:", -1)
					})

					if window.selectedGuild != nil {
						guild, discordError := window.session.State.Guild(window.selectedGuild.ID)
						if discordError == nil {

							//Those could be optimized by searching the string for patterns.
							for _, channel := range guild.Channels {
								if channel.Type == discordgo.ChannelTypeGuildText {
									messageToSend = strings.Replace(messageToSend, "#"+channel.Name, "<#"+channel.ID+">", -1)
								}
							}

							messageToSend = emojiRegex.ReplaceAllStringFunc(messageToSend, func(match string) string {
								matchStripped := strings.TrimPrefix(strings.TrimSuffix(match, ":"), ":")
								for _, emoji := range guild.Emojis {
									if emoji.Name == matchStripped {
										return "<:" + emoji.Name + ":" + emoji.ID + ">"
									}
								}

								return match
							})
						}
					}

					//Replace formatter characters and replace emoji codes.
					messageToSend = discordemojimap.Replace(messageToSend)
					messageToSend = strings.Replace(messageToSend, "\\:", ":", -1)

					if window.selectedGuild != nil {
						members, discordError := window.session.State.Members(window.selectedGuild.ID)
						if discordError == nil {
							for _, member := range members {
								messageToSend = strings.Replace(messageToSend, "@"+member.User.Username+"#"+member.User.Discriminator, "<@"+member.User.ID+">", -1)
							}
						}
					} else if window.selectedChannel != nil {
						for _, user := range window.selectedChannel.Recipients {
							messageToSend = strings.Replace(messageToSend, "@"+user.Username+"#"+user.Discriminator, "<@"+user.ID+">", -1)
						}
					}

					if window.editingMessageID != nil {
						overLength := len(messageToSend) - 2000
						if overLength > 0 {
							window.app.QueueUpdateDraw(func() {
								window.ShowErrorDialog(fmt.Sprintf("The message you are trying to send is %d characters too long.", overLength))
							})
						} else {
							go window.editMessage(window.selectedChannel.ID, *window.editingMessageID, messageToSend)
							window.exitMessageEditMode()
						}
					} else {
						go func() {
							messageText := window.jsEngine.OnMessageSend(messageToSend)
							overLength := len(messageText) - 2000
							if overLength > 0 {
								window.app.QueueUpdateDraw(func() {
									window.ShowErrorDialog(fmt.Sprintf("The message you are trying to send is %d characters too long.", overLength))
								})
								return
							}

							_, sendError := session.ChannelMessageSend(window.selectedChannel.ID, messageText)
							window.chatView.internalTextView.ScrollToEnd()
							if sendError != nil {
								window.app.QueueUpdateDraw(func() {
									window.ShowErrorDialog("Error sending message: " + sendError.Error())
								})
							}
						}()
					}
				} else {
					if window.editingMessageID != nil {
						msgIDCopy := *window.editingMessageID
						window.askForMessageDeletion(msgIDCopy, true)
					}
				}

				return nil
			}
		}

		return event
	})

	messageInputChan := make(chan *discordgo.Message, 200)
	messageDeleteChan := make(chan *discordgo.Message, 50)
	messageEditChan := make(chan *discordgo.Message, 50)
	messageBulkDeleteChan := make(chan *discordgo.MessageDeleteBulk, 50)

	window.registerMessageEventHandler(messageInputChan, messageEditChan, messageDeleteChan, messageBulkDeleteChan)
	window.startMessageHandlerRoutines(messageInputChan, messageEditChan, messageDeleteChan, messageBulkDeleteChan)

	window.userList = NewUserTree(window.session.State)

	if config.GetConfig().OnTypeInListBehaviour == config.SearchOnTypeInList {
		guildList.SetSearchOnTypeEnabled(true)
		channelTree.internalTreeView.SetSearchOnTypeEnabled(true)
		window.userList.internalTreeView.SetSearchOnTypeEnabled(true)
		window.privateList.internalTreeView.SetSearchOnTypeEnabled(true)
	} else if config.GetConfig().OnTypeInListBehaviour == config.FocusMessageInputOnTypeInList {
		guildList.SetInputCapture(tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView))
		channelTree.internalTreeView.SetInputCapture(tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView))
		window.userList.SetInputCapture(tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView))
		window.privateList.SetInputCapture(tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView))
		window.chatView.internalTextView.SetInputCapture(tviewutil.CreateFocusTextViewOnTypeInputHandler(
			window.app, window.messageInput.internalTextView))
	}

	//Guild Container arrow key navigation. Please end my life.
	oldGuildListHandler := guildList.GetInputCapture()
	newGuildHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
				window.app.SetFocus(window.channelTree.internalTreeView)
				return nil
			}

			if event.Key() == tcell.KeyLeft {
				if window.userList.internalTreeView.IsVisible() {
					window.app.SetFocus(window.userList.internalTreeView)
				} else {
					window.app.SetFocus(window.chatView.internalTextView)
				}
				return nil
			}

			if event.Key() == tcell.KeyRight {
				window.app.SetFocus(window.chatView.internalTextView)
				return nil
			}
		}

		return event
	}

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

	//Channel Container arrow key navigation. Please end my life.
	oldChannelListHandler := channelTree.internalTreeView.GetInputCapture()
	newChannelListHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
				window.app.SetFocus(window.guildList)
				return nil
			}

			if event.Key() == tcell.KeyLeft {
				if window.userList.internalTreeView.IsVisible() {
					window.app.SetFocus(window.userList.internalTreeView)
				} else {
					window.app.SetFocus(window.messageInput.GetPrimitive())
				}
				return nil
			}

			if event.Key() == tcell.KeyRight {
				window.app.SetFocus(window.messageInput.GetPrimitive())
				return nil
			}
		}

		return event
	}

	if oldChannelListHandler == nil {
		channelTree.internalTreeView.SetInputCapture(newChannelListHandler)
	} else {
		channelTree.internalTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newChannelListHandler(event)
			if handledEvent != nil {
				return oldChannelListHandler(event)
			}

			return event
		})
	}

	//Chatview arrow key navigation. Please end my life.
	oldChatViewHandler := window.chatView.internalTextView.GetInputCapture()
	newChatViewHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyUp {
				window.app.SetFocus(window.messageInput.GetPrimitive())
				return nil
			}

			if event.Key() == tcell.KeyLeft {
				if window.leftArea.GetCurrentPage() == guildPageName {
					window.app.SetFocus(window.guildList)
					return nil
				} else if window.leftArea.GetCurrentPage() == guildPageName {
					window.app.SetFocus(window.privateList.internalTreeView)
					return nil
				}
			}

			if event.Key() == tcell.KeyRight {
				if window.userList.internalTreeView.IsVisible() {
					window.app.SetFocus(window.userList.internalTreeView)
				} else {
					if window.leftArea.GetCurrentPage() == guildPageName {
						window.app.SetFocus(window.guildList)
						return nil
					} else if window.leftArea.GetCurrentPage() == guildPageName {
						window.app.SetFocus(window.privateList.internalTreeView)
						return nil
					}
				}
				return nil
			}
		}

		return event
	}

	if oldChatViewHandler == nil {
		window.chatView.internalTextView.SetInputCapture(newChatViewHandler)
	} else {
		window.chatView.internalTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newChatViewHandler(event)
			if handledEvent != nil {
				return oldChatViewHandler(event)
			}

			return event
		})
	}

	//User Container arrow key navigation. Please end my life.
	oldUserListHandler := window.userList.internalTreeView.GetInputCapture()
	newUserListHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyRight {
				if window.leftArea.GetCurrentPage() == guildPageName {
					window.app.SetFocus(window.guildList)
					return nil
				} else if window.leftArea.GetCurrentPage() == guildPageName {
					window.app.SetFocus(window.privateList.internalTreeView)
					return nil
				}
				return nil
			}
			if event.Key() == tcell.KeyLeft {
				window.app.SetFocus(window.chatView.GetPrimitive())
				return nil
			}
		}

		return event
	}

	if oldUserListHandler == nil {
		window.userList.internalTreeView.SetInputCapture(newUserListHandler)
	} else {
		window.userList.internalTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newUserListHandler(event)
			if handledEvent != nil {
				return oldUserListHandler(event)
			}

			return event
		})
	}

	//Private Container arrow key navigation. Please end my life.
	oldPrivateListHandler := window.privateList.internalTreeView.GetInputCapture()
	newPrivateListHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers() == tcell.ModAlt {
			if event.Key() == tcell.KeyLeft {
				if window.userList.internalTreeView.IsVisible() {
					window.app.SetFocus(window.userList.internalTreeView)
				} else {
					window.app.SetFocus(window.chatView.internalTextView)
				}
				return nil
			}

			if event.Key() == tcell.KeyRight {
				window.app.SetFocus(window.chatView.internalTextView)
				return nil
			}
		}

		return event
	}

	if oldPrivateListHandler == nil {
		window.privateList.internalTreeView.SetInputCapture(newPrivateListHandler)
	} else {
		window.privateList.internalTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			handledEvent := newPrivateListHandler(event)
			if handledEvent != nil {
				return oldPrivateListHandler(event)
			}

			return event
		})
	}

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

	app.SetRoot(window.rootContainer, true)
	window.currentContainer = window.rootContainer
	app.SetInputCapture(window.handleGlobalShortcuts)

	conf := config.GetConfig()

	if conf.UseFixedLayout {
		window.middleContainer.AddItem(window.leftArea, conf.FixedSizeLeft, 0, true)
		window.middleContainer.AddItem(window.chatArea, 0, 1, false)
		window.middleContainer.AddItem(window.userList.internalTreeView, conf.FixedSizeRight, 0, false)
	} else {
		window.middleContainer.AddItem(window.leftArea, 0, 7, true)
		window.middleContainer.AddItem(window.chatArea, 0, 20, false)
		window.middleContainer.AddItem(window.userList.internalTreeView, 0, 6, false)
	}

	mentionWindow.SetVisible(false)
	mentionWindow.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Key(); key {
		case tcell.KeyRune, tcell.KeyDelete, tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft, tcell.KeyRight, tcell.KeyCtrlA, tcell.KeyCtrlV:
			window.messageInput.internalTextView.GetInputCapture()(event)
			return nil
		}
		return event
	})
	mentionWindow.SetSelectedFunc(func(node *tview.TreeNode) {
		beginIdx, endIdx := window.messageInput.GetCurrentMentionIndices()
		if beginIdx != endIdx {
			data, ok := node.GetReference().(string)
			oldText := window.messageInput.GetText()
			if ok {
				newText := oldText[:beginIdx] + strings.TrimSpace(data) + oldText[endIdx+1:] + " "
				window.messageInput.SetText(newText)
			} else {
				role, ok := node.GetReference().(*discordgo.Role)
				if ok {
					newText := oldText[:beginIdx-1] + "<@&" + strings.TrimSpace(role.ID) + ">" + oldText[endIdx+1:] + " "
					window.messageInput.SetText(newText)
				}
			}
		}
		window.messageInput.mentionHideHandler()
	})

	window.chatArea.AddItem(window.messageContainer, 0, 1, false)
	window.chatArea.AddItem(mentionWindow, 2, 2, true)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), window.messageInput.GetRequestedHeight(), 0, false)

	window.commandView.commandOutput.SetVisible(false)
	window.commandView.commandInput.internalTextView.SetVisible(false)

	window.chatArea.AddItem(window.commandView.commandOutput, 0, 1, false)
	window.chatArea.AddItem(window.commandView.commandInput.internalTextView, 3, 0, false)

	window.SwitchToGuildsPage()

	app.SetFocus(guildList)

	window.registerMouseFocusListeners()

	return &window, nil
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
	height := tviewutil.CalculateNeccessaryHeight(width, window.dialogTextView.GetText(true))
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

	window.guildList.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.guildList)
			return true
		}

		return false
	})
	window.channelTree.internalTreeView.SetMouseHandler(func(event *tcell.EventMouse) bool {
		if event.Buttons() == tcell.Button1 {
			window.app.SetFocus(window.channelTree.internalTreeView)

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
			window.app.SetFocus(window.privateList.internalTreeView)

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
		//Ignore just-embed edits
		if m.Content != "" {
			edit <- m.Message
		}
	})
}

// startMessageHandlerRoutines registers the handlers for certain message events.
// It updates the cache and the UI if necessary.
func (window *Window) startMessageHandlerRoutines(input, edit, delete chan *discordgo.Message, bulkDelete chan *discordgo.MessageDeleteBulk) {
	go func() {
		for message := range input {
			tempMessage := message
			window.session.State.MessageAdd(tempMessage)

			if window.selectedChannel != nil && tempMessage.ChannelID == window.selectedChannel.ID {
				window.app.QueueUpdateDraw(func() {
					window.chatView.AddMessage(tempMessage)
				})
			}

			if tempMessage.Author.ID == window.session.State.User.ID {
				continue
			}

			channel, stateError := window.session.State.Channel(tempMessage.ChannelID)
			if stateError != nil {
				continue
			}
			if (window.selectedChannel == nil || tempMessage.ChannelID != window.selectedChannel.ID) ||
				!window.userActive {
				mentionsYou := false
				for _, user := range tempMessage.Mentions {
					if user.ID == window.session.State.User.ID {
						mentionsYou = true
						break
					}
				}

				if config.GetConfig().DesktopNotifications {
					if !mentionsYou {
						//TODO Check if channel is muted.
						if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
							mentionsYou = true
						}
					}

					if mentionsYou {
						var notificationLocation string

						if channel.Type == discordgo.ChannelTypeDM {
							notificationLocation = tempMessage.Author.Username
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

							notificationLocation = tempMessage.Author.Username + " - " + notificationLocation
						} else if channel.Type == discordgo.ChannelTypeGuildText {
							if tempMessage.GuildID != "" {
								guild, cacheError := window.session.State.Guild(tempMessage.GuildID)
								if guild != nil && cacheError == nil {
									notificationLocation = fmt.Sprintf("%s - %s - %s", guild.Name, channel.Name, tempMessage.Author.Username)
								} else {
									notificationLocation = fmt.Sprintf("%s - %s", tempMessage.Author.Username, channel.Name)
								}
							} else {
								notificationLocation = fmt.Sprintf("%s - %s", tempMessage.Author.Username, channel.Name)
							}
						}

						beeep.Notify("Cordless - "+notificationLocation, tempMessage.ContentWithMentionsReplaced(), "assets/information.png")
					}
				}

				//We needn't adjust the text of the currently selected channel.
				if window.selectedChannel == nil || tempMessage.ChannelID == window.selectedChannel.ID {
					continue
				}

				if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
					window.app.QueueUpdateDraw(func() {
						window.privateList.MarkChannelAsUnread(channel)
					})
				} else if channel.Type == discordgo.ChannelTypeGuildText {
					window.app.QueueUpdateDraw(func() {
						if mentionsYou {
							window.channelTree.MarkChannelAsMentioned(channel.ID)
						} else {
							window.channelTree.MarkChannelAsUnread(channel.ID)
						}
					})
				}
			}
		}
	}()

	go func() {
		for messageDeleted := range delete {
			tempMessageDeleted := messageDeleted
			window.session.State.MessageRemove(tempMessageDeleted)
			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessageDeleted.ChannelID {
				window.app.QueueUpdateDraw(func() {
					window.chatView.DeleteMessage(tempMessageDeleted)
				})
			}
		}
	}()

	go func() {
		for messagesDeleted := range bulkDelete {
			tempMessagesDeleted := messagesDeleted
			for _, messageID := range messagesDeleted.Messages {
				message, stateError := window.session.State.Message(tempMessagesDeleted.ChannelID, messageID)
				if stateError == nil {
					window.session.State.MessageRemove(message)
				}
			}

			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessagesDeleted.ChannelID {
				window.app.QueueUpdateDraw(func() {
					window.chatView.DeleteMessages(tempMessagesDeleted.Messages)
				})
			}
		}
	}()

	go func() {
		for messageEdited := range edit {
			tempMessageEdited := messageEdited
			window.session.State.MessageAdd(tempMessageEdited)
			if window.selectedChannel != nil && window.selectedChannel.ID == tempMessageEdited.ChannelID {
				for _, message := range window.chatView.data {
					if message.ID == tempMessageEdited.ID {
						message.Content = tempMessageEdited.Content
						window.app.QueueUpdateDraw(func() {
							window.chatView.UpdateMessage(message)
						})
						break
					}
				}
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
			window.app.QueueUpdateDraw(func() {
				for _, relationship := range window.session.State.Relationships {
					if relationship.ID == event.ID {
						window.privateList.RemoveFriend(relationship.User.ID)
						break
					}
				}
			})
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
			window.app.QueueUpdateDraw(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelUpdate) {
		if window.isChannelEventRelevant(event.Channel) {
			window.app.QueueUpdateDraw(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelDelete) {
		if window.isChannelEventRelevant(event.Channel) {
			window.app.QueueUpdateDraw(func() {
				window.channelTree.RemoveChannel(event.Channel)
			})
		}
	})
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
	if event.Key() == tcell.KeyCtrlC {
		window.doRestart <- false
		return event
	}

	// Maybe compare directly to table?
	if window.currentContainer != window.rootContainer {
		return event
	}

	window.userActive = true
	window.userActiveTimer.Reset(userInactiveTime)

	if window.dialogReplacement.IsVisible() {
		return event
	}

	if event.Modifiers()&tcell.ModAlt == tcell.ModAlt && event.Rune() == 'S' {
		var table *shortcuts.ShortcutTable
		var exitButton *tview.Button
		var resetButton *tview.Button

		table = shortcuts.NewShortcutTable()
		table.SetShortcuts(shortcuts.Shortcuts)

		doClose := func() {
			window.app.SetRoot(window.rootContainer, true)
			window.currentContainer = window.rootContainer
			window.app.ForceDraw()
		}
		table.SetOnClose(doClose)

		exitButton = tview.NewButton("Go back")
		exitButton.SetSelectedFunc(doClose)
		exitButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				window.app.SetFocus(table.GetPrimitive())
			} else if event.Key() == tcell.KeyBacktab {
				window.app.SetFocus(resetButton)
			}

			return event
		})

		resetButton = tview.NewButton("Restore defaults")
		resetButton.SetSelectedFunc(func() {
			for _, shortcut := range shortcuts.Shortcuts {
				shortcut.Reset()
			}
			shortcuts.Persist()

			table.SetShortcuts(shortcuts.Shortcuts)
			window.app.ForceDraw()
		})
		resetButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				window.app.SetFocus(exitButton)
			} else if event.Key() == tcell.KeyBacktab {
				window.app.SetFocus(table.GetPrimitive())
			}

			return event
		})

		table.SetFocusNext(func() { window.app.SetFocus(resetButton) })
		table.SetFocusPrevious(func() { window.app.SetFocus(exitButton) })

		buttonBar := tview.NewFlex()
		buttonBar.SetDirection(tview.FlexColumn)

		buttonBar.AddItem(resetButton, 0, 1, false)
		buttonBar.AddItem(tview.NewBox(), 1, 0, false)
		buttonBar.AddItem(exitButton, 0, 1, false)

		shortcutsView := tview.NewFlex()
		shortcutsView.SetDirection(tview.FlexRow)

		shortcutsView.AddItem(table.GetPrimitive(), 0, 1, false)
		shortcutsView.AddItem(buttonBar, 1, 0, false)

		window.app.SetRoot(shortcutsView, true)
		window.app.SetFocus(table.GetPrimitive())
		window.currentContainer = shortcutsView
	} else if shortcuts.EventsEqual(event, shortcuts.ToggleCommandView.Event) {
		window.SetCommandModeEnabled(!window.commandMode)

		if window.commandMode {
			window.app.SetFocus(window.commandView.commandInput.internalTextView)
		} else {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}
	} else if shortcuts.EventsEqual(event, shortcuts.FocusCommandOutput.Event) {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandOutput)
	} else if shortcuts.EventsEqual(event, shortcuts.FocusCommandInput.Event) {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandInput.internalTextView)
	} else if shortcuts.EventsEqual(event, shortcuts.ToggleUserContainer.Event) {
		conf := config.GetConfig()
		conf.ShowUserContainer = !conf.ShowUserContainer

		if !conf.ShowUserContainer {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}

		config.PersistConfig()
		window.RefreshLayout()
	} else if shortcuts.EventsEqual(event, shortcuts.FocusChannelContainer.Event) {
		window.SwitchToGuildsPage()
		window.app.SetFocus(window.channelTree.internalTreeView)
	} else if shortcuts.EventsEqual(event, shortcuts.FocusPrivateChatPage.Event) {
		window.SwitchToFriendsPage()
		window.app.SetFocus(window.privateList.GetComponent())
	} else if shortcuts.EventsEqual(event, shortcuts.FocusGuildContainer.Event) {
		window.SwitchToGuildsPage()
		window.app.SetFocus(window.guildList)
	} else if shortcuts.EventsEqual(event, shortcuts.FocusMessageContainer.Event) {
		window.app.SetFocus(window.chatView.internalTextView)
	} else if shortcuts.EventsEqual(event, shortcuts.FocusUserContainer.Event) {
		if window.leftArea.GetCurrentPage() == guildPageName && window.userList.internalTreeView.IsVisible() {
			window.app.SetFocus(window.userList.internalTreeView)
		}
	} else if shortcuts.EventsEqual(event, shortcuts.FocusMessageInput.Event) {
		window.app.SetFocus(window.messageInput.GetPrimitive())
	} else {
		return event
	}

	return nil
}

//ExecuteCommand tries to execute the given input as a command. The first word
//will be passed as the commands name and the rest will be parameters. If a
//command can't be found, that info will be printed onto the command output.
func (window *Window) ExecuteCommand(command string) {
	parts := commands.ParseCommand(command)
	commandLogic, exists := window.commands[parts[0]]
	if exists {
		commandLogic.Execute(window.commandView, parts[1:])
	} else {
		fmt.Fprintf(window.commandView, "[red]The command '%s' doesn't exist[white]\n", parts[0])
	}
}

func (window *Window) startEditingMessage(message *discordgo.Message) {
	if message.Author.ID == window.session.State.User.ID {
		window.messageInput.SetText(message.Content)
		window.messageInput.SetBackgroundColor(tcell.ColorDarkGoldenrod)
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
	window.messageInput.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
}

//ShowErrorDialog shows a simple error dialog that has only an Okay button,
// a generic title and the given text.
func (window *Window) ShowErrorDialog(text string) {
	window.ShowDialog(tcell.ColorRed, "An error occured - "+text, func(_ string) {}, "Okay")
}

func (window *Window) editMessage(channelID, messageID, messageEdited string) {
	go func() {
		_, discordError := window.session.ChannelMessageEdit(channelID, messageID, messageEdited)
		if discordError != nil {
			window.app.QueueUpdateDraw(func() {
				window.ShowErrorDialog("Error editing message.")
			})
		}
	}()

	window.exitMessageEditMode()
}

//SwitchToGuildsPage the left side of the layout over to the view where you can
//see the servers and their channels. In additional to that, it also shows the
//user list in case the user didn't explicitly hide it.
func (window *Window) SwitchToGuildsPage() {
	if window.leftArea.GetCurrentPage() != guildPageName {
		window.leftArea.SwitchToPage(guildPageName)
		window.RefreshLayout()
	}
}

//SwitchToFriendsPage switches the left side of the layout over to the view
//where you can see your private chats and groups. In addition to that it
//hides the user list.
func (window *Window) SwitchToFriendsPage() {
	if window.leftArea.GetCurrentPage() != privatePageName {
		window.leftArea.SwitchToPage(privatePageName)
		window.RefreshLayout()
	}
}

//RefreshLayout removes and adds the main parts of the layout
//so that the ones that are disabled by settings do not show up.
func (window *Window) RefreshLayout() {
	conf := config.GetConfig()

	window.userList.internalTreeView.SetVisible(conf.ShowUserContainer && (window.selectedGuild != nil ||
		(window.selectedChannel != nil && window.selectedChannel.Type == discordgo.ChannelTypeGroupDM)))

	if conf.UseFixedLayout {
		window.middleContainer.ResizeItem(window.leftArea, conf.FixedSizeLeft, 7)
		window.middleContainer.ResizeItem(window.chatArea, 0, 1)
		window.middleContainer.ResizeItem(window.userList.internalTreeView, conf.FixedSizeRight, 6)
	} else {
		window.middleContainer.ResizeItem(window.leftArea, 0, 7)
		window.middleContainer.ResizeItem(window.chatArea, 0, 20)
		window.middleContainer.ResizeItem(window.userList.internalTreeView, 0, 6)
	}

	window.app.ForceDraw()
}

//LoadChannel eagerly loads the channels messages.
func (window *Window) LoadChannel(channel *discordgo.Channel) error {
	var messages []*discordgo.Message

	// Data not present
	if channel.LastMessageID != "" && len(channel.Messages) == 0 {
		//Check Cache first
		cache, cacheError := window.session.State.Channel(channel.ID)
		if cacheError == nil || cache != nil && len(cache.Messages) == 0 {
			var discordError error
			messages, discordError = window.session.ChannelMessages(channel.ID, 100, "", "", "")
			if discordError == nil {
				if channel.GuildID != "" {
					for _, message := range messages {
						message.GuildID = channel.GuildID
					}
				}
				cache.Messages = append(cache.Messages, messages...)
			}
		} else {
			messages = make([]*discordgo.Message, 0)
		}
	} else {
		messages = channel.Messages
	}

	discordutil.SortMessagesByTimestamp(messages)

	window.chatView.SetMessages(messages)
	window.chatView.ClearSelection()
	window.chatView.internalTextView.ScrollToEnd()

	window.UpdateChatHeader(channel)

	window.selectedChannel = channel
	if channel.GuildID == "" {
		if window.selectedChannelNode != nil {
			window.selectedChannelNode.SetColor(tview.Styles.PrimaryTextColor)
			window.selectedChannelNode = nil
		}

		if window.selectedGuildNode != nil {
			window.selectedGuildNode.SetColor(tview.Styles.PrimaryTextColor)
			window.selectedGuildNode = nil
		}
	}

	if channel.GuildID == "" {
		window.selectedGuild = nil
		window.selectedGuildNode = nil
	}

	if channel.Type == discordgo.ChannelTypeDM || channel.Type == discordgo.ChannelTypeGroupDM {
		window.privateList.MarkChannelAsLoaded(channel)
	}

	window.exitMessageEditModeAndKeepText()

	if config.GetConfig().FocusMessageInputAfterChannelSelection {
		window.app.SetFocus(window.messageInput.internalTextView)
	}

	return nil
}

// RefreshChannelTitle calls UpdateChannelTitle using the current channel.
func (window *Window) RefreshChannelTitle() {
	window.UpdateChatHeader(window.selectedChannel)
}

// UpdateChatHeader updates the bordertitle of the chatviews container.o
// The title consist of the channel name and its topic for guild channels.
// For private channels it's either the recipient in a dm, or all recipients
// in a group dm channel. If the channel has a nickname, that is chosen.
func (window *Window) UpdateChatHeader(channel *discordgo.Channel) {
	if !config.GetConfig().ShowChatHeader || channel == nil {
		window.chatView.SetTitle("")
	} else {
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
}

//RegisterCommand register a command. That makes the command available for
//being called from the message input field, in case the user-defined prefix
//is in front of the input.
func (window *Window) RegisterCommand(command commands.Command) {
	window.commands[command.Name()] = command
}

// GetRegisteredCommands returns the map of all registered commands
func (window *Window) GetRegisteredCommands() map[string]commands.Command {
	return window.commands
}

//Run Shows the window optionally returning an error.
func (window *Window) Run() error {
	return window.app.Run()
}

// Shutdown disconnects from the discord API and stops the tview application.
func (window *Window) Shutdown() {
	window.chatView.shortener.Close()
	window.session.Close()
	window.app.Stop()
}
