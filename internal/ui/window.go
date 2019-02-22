package ui

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Bios-Marcel/goclipimg"

	"github.com/atotto/clipboard"

	"github.com/Bios-Marcel/cordless/internal/commands"
	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/Bios-Marcel/cordless/internal/maths"
	"github.com/Bios-Marcel/cordless/internal/scripting"
	"github.com/Bios-Marcel/cordless/internal/scripting/js"
	"github.com/Bios-Marcel/cordless/internal/times"
	"github.com/Bios-Marcel/cordless/internal/ui/tview/treeview"
	"github.com/Bios-Marcel/discordemojimap"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
	"github.com/gen2brain/beeep"
)

const (
	userListUpdateInterval = 5 * time.Second

	guildPageName   = "Guilds"
	privatePageName = "Private"
)

var (
	mentionRegex = regexp.MustCompile("@.*?(?:$|\\s)")
)

// Window is basically the whole application, as it contains all the
// components and the necccessary global state.
type Window struct {
	app           *tview.Application
	rootContainer *tview.Flex

	leftArea    *tview.Pages
	guildList   *tview.TreeView
	channelTree *ChannelTree
	privateList *PrivateChatList

	channelRootNode *tview.TreeNode
	channelTitle    *tview.TextView

	chatArea         *tview.Flex
	chatView         *ChatView
	messageContainer tview.Primitive
	messageInput     *Editor

	editingMessageID *string

	userList *UserTree

	overrideShowUsers bool

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
func NewWindow(doRestart chan bool, app *tview.Application, discord *discordgo.Session) (*Window, error) {
	window := Window{
		doRestart:       doRestart,
		session:         discord,
		app:             app,
		commands:        make(map[string]commands.Command, 1),
		jsEngine:        js.New(),
		userActiveTimer: time.NewTimer(10 * time.Second),
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

	guilds, discordError := discordgoplus.LoadGuilds(window.session)
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

	discordgoplus.SortGuilds(window.session.State.Settings, guilds)

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
			window.selectedGuildNode.SetColor(tcell.ColorTeal)

			window.selectedGuild = guild
			discord.RequestGuildMembers(guild.ID, "", 0)

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
		})
	}

	if len(guildRootNode.GetChildren()) > 0 {
		guildList.SetCurrentNode(guildRootNode)
	}

	guildPage.AddItem(guildList, 0, 1, true)
	guildPage.AddItem(channelTree.internalTreeView, 0, 2, true)

	window.leftArea.AddPage(guildPageName, guildPage, true, false)

	window.privateList = NewPrivateChatList(window.session.State)
	//TODO Currently there can't be an error ... might as well remove the error
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
		window.channelTitle.SetText(discordgoplus.GetPrivateChannelName(channel))
		if channel.Type == discordgo.ChannelTypeDM {
			window.overrideShowUsers = false
		} else if channel.Type == discordgo.ChannelTypeGroupDM {
			window.overrideShowUsers = true
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
				window.channelTitle.SetText(userChannel.Recipients[0].Username)
				return
			}
		}

		newChannel, discordError := window.session.UserChannelCreate(userID)
		if discordError == nil {
			window.LoadChannel(newChannel)
			window.channelTitle.SetText(newChannel.Recipients[0].Username)
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
			for i := len(window.chatView.data) - 1; i > 0; i-- {
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
					if window.selectedGuild != nil {
						guild, discordError := window.session.State.Guild(window.selectedGuild.ID)
						if discordError == nil {

							//Those could be optimized by searching the string for patterns.
							for _, channel := range guild.Channels {
								if channel.Type == discordgo.ChannelTypeGuildText {
									messageToSend = strings.Replace(messageToSend, "#"+channel.Name, "<#"+channel.ID+">", -1)
								}
							}

						}
					}

					messageToSend = codeBlockRegex.ReplaceAllStringFunc(messageToSend, func(input string) string {
						return strings.Replace(input, ":", "\\:", -1)
					})
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

							_, sendError := discord.ChannelMessageSend(window.selectedChannel.ID, messageText)
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

	window.addMessageEventHandler(messageInputChan, messageEditChan, messageDeleteChan)
	window.startMessageHandlerRoutines(messageInputChan, messageEditChan, messageDeleteChan)

	window.channelTitle = tview.NewTextView()
	window.channelTitle.SetBorderSides(true, true, false, true)
	window.channelTitle.SetBorder(true)

	window.userList = NewUserTree(window.session.State)

	if config.GetConfig().OnTypeInListBehaviour == config.SearchOnTypeInList {
		guildList.SetSearchOnTypeEnabled(true)
		channelTree.internalTreeView.SetSearchOnTypeEnabled(true)
		window.userList.internalTreeView.SetSearchOnTypeEnabled(true)
		window.privateList.internalTreeView.SetSearchOnTypeEnabled(true)
	} else if config.GetConfig().OnTypeInListBehaviour == config.FocusMessageInputOnTypeInList {
		guildList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			guildList.Box, window.app, window.messageInput.internalTextView))
		channelTree.internalTreeView.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			channelTree.internalTreeView.Box, window.app, window.messageInput.internalTextView))
		window.userList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			window.userList.internalTreeView.Box, window.app, window.messageInput.internalTextView))
		window.privateList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			window.privateList.GetComponent().Box, window.app, window.messageInput.internalTextView))
		window.chatView.internalTextView.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			window.chatView.internalTextView.Box, window.app, window.messageInput.internalTextView))
	}

	window.rootContainer = tview.NewFlex().
		SetDirection(tview.FlexColumn)
	window.rootContainer.SetTitleAlign(tview.AlignCenter)

	app.SetRoot(window.rootContainer, true)
	app.SetInputCapture(window.handleGlobalShortcuts)

	conf := config.GetConfig()

	if conf.UseFixedLayout {
		window.rootContainer.AddItem(window.leftArea, conf.FixedSizeLeft, 7, true)
		window.rootContainer.AddItem(window.chatArea, 0, 1, false)
		window.rootContainer.AddItem(window.userList.internalTreeView, conf.FixedSizeRight, 6, false)
	} else {
		window.rootContainer.AddItem(window.leftArea, 0, 7, true)
		window.rootContainer.AddItem(window.chatArea, 0, 20, false)
		window.rootContainer.AddItem(window.userList.internalTreeView, 0, 6, false)
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

	window.chatArea.AddItem(window.channelTitle, 2, 0, false)
	window.chatArea.AddItem(window.messageContainer, 0, 1, false)
	window.chatArea.AddItem(mentionWindow, 2, 2, true)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), window.messageInput.GetRequestedHeight(), 0, false)

	window.commandView.commandOutput.SetVisible(false)
	window.commandView.commandInput.internalTextView.SetVisible(false)

	window.chatArea.AddItem(window.commandView.commandOutput, 0, 1, false)
	window.chatArea.AddItem(window.commandView.commandInput.internalTextView, 3, 0, false)

	if conf.ShowFrame {
		window.rootContainer.SetTitle("Cordless")
		window.rootContainer.SetBorder(true)
	} else {
		window.rootContainer.SetTitle("")
		window.rootContainer.SetBorder(false)
	}

	window.SwitchToGuildsPage()

	app.SetFocus(guildList)

	window.registerMouseFocusListeners()

	return &window, nil
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

func (window *Window) addMessageEventHandler(inputChannel, editChannel, deleteChannel chan *discordgo.Message) {
	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if window.selectedChannel != nil {
			inputChannel <- m.Message
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		if window.selectedChannel != nil {
			if m.ChannelID == window.selectedChannel.ID {
				deleteChannel <- m.Message
			}
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if window.selectedChannel != nil {
			if m.ChannelID == window.selectedChannel.ID &&
				//Ignore just-embed edits
				m.Content != "" {
				editChannel <- m.Message
			}
		}
	})
}

func (window *Window) startMessageHandlerRoutines(inputChannel, editChannel, deleteChannel chan *discordgo.Message) {
	go func() {
		for {
			select {
			case message := <-inputChannel:
				//UPDATE CACHE
				window.session.State.MessageAdd(message)

				if message.ChannelID == window.selectedChannel.ID {
					window.app.QueueUpdateDraw(func() {
						window.chatView.AddMessage(message)
					})
				}

				if message.Author.ID == window.session.State.User.ID {
					continue
				}

				channel, stateError := window.session.State.Channel(message.ChannelID)
				if stateError != nil {
					continue
				}

				if message.ChannelID != window.selectedChannel.ID || !window.userActive {
					mentionsYou := false
					for _, user := range message.Mentions {
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
								notificationLocation = message.Author.Username + " - " + channel.Name
							}

							beeep.Notify("Cordless - "+notificationLocation, message.ContentWithMentionsReplaced(), "assets/information.png")
						}
					}

					//We needn't adjust the text of the currently selected channel.
					if message.ChannelID == window.selectedChannel.ID {
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
		}
	}()

	go func() {
		for {
			select {
			case messageDeleted := <-deleteChannel:
				//UPDATE CACHE
				window.session.State.MessageRemove(messageDeleted)
				for _, message := range window.chatView.data {
					if message.ID == messageDeleted.ID {
						window.app.QueueUpdateDraw(func() {
							window.chatView.DeleteMessage(message)
						})
						break
					}
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case messageEdited := <-editChannel:
				//UPDATE CACHE
				window.session.State.MessageAdd(messageEdited)
				for _, message := range window.chatView.data {
					if message.ID == messageEdited.ID {
						message.Content = messageEdited.Content
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
		if event.Relationship.Type == discordgoplus.RelationTypeFriend {
			window.app.QueueUpdateDraw(func() {
				window.privateList.addFriend(event.User)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.RelationshipRemove) {
		if event.Relationship.Type == discordgoplus.RelationTypeFriend {
			window.app.QueueUpdateDraw(func() {
				window.privateList.addFriend(event.User)
			})
		}
	})
}

func (window *Window) registerGuildChannelHandler() {
	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelCreate) {
		if window.selectedGuild == nil {
			return
		}

		if event.Type == discordgo.ChannelTypeGuildText || event.Type == discordgo.ChannelTypeGuildCategory {
			window.app.QueueUpdateDraw(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelUpdate) {
		if event.Type == discordgo.ChannelTypeGuildText || event.Type == discordgo.ChannelTypeGuildCategory {
			if window.selectedGuild == nil {
				return
			}

			window.app.QueueUpdateDraw(func() {
				window.channelTree.AddOrUpdateChannel(event.Channel)
			})
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, event *discordgo.ChannelDelete) {
		if event.Type == discordgo.ChannelTypeGuildText || event.Type == discordgo.ChannelTypeGuildCategory {
			if window.selectedGuild == nil {
				return
			}

			window.app.QueueUpdateDraw(func() {
				window.channelTree.RemoveChannel(event.Channel)
			})
		}
	})
}

func (window *Window) askForMessageDeletion(messageID string, usedWithSelection bool) {
	previousFocus := window.app.GetFocus()
	dialog := tview.NewModal()
	dialog.SetText("Do you really want to delete the message?")
	dialog.AddButtons([]string{"Abort", "Delete"})

	dialog.SetDoneFunc(func(index int, label string) {
		if index == 1 {
			go window.session.ChannelMessageDelete(window.selectedChannel.ID, messageID)
		}

		window.exitMessageEditMode()
		window.app.SetRoot(window.rootContainer, true)
		if usedWithSelection {
			window.app.SetFocus(previousFocus)
			window.chatView.SignalSelectionDeleted()
		} else {
			window.app.SetFocus(window.messageInput.GetPrimitive())

		}
	})

	window.app.SetRoot(dialog, false)
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

	window.userActive = true
	window.userActiveTimer.Reset(10 * time.Second)

	if event.Rune() == '.' &&
		(event.Modifiers()&tcell.ModAlt) == tcell.ModAlt {

		window.SetCommandModeEnabled(!window.commandMode)

		if window.commandMode {
			window.app.SetFocus(window.commandView.commandInput.internalTextView)
		} else {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}

		return nil
	}

	if window.commandMode && event.Key() == tcell.KeyCtrlO {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandOutput)
	}

	if window.commandMode && event.Key() == tcell.KeyCtrlI {
		if !window.commandMode {
			window.SetCommandModeEnabled(true)
		}

		window.app.SetFocus(window.commandView.commandInput.internalTextView)
	}

	if event.Rune() == 'U' &&
		(event.Modifiers()&tcell.ModAlt) == tcell.ModAlt {
		conf := config.GetConfig()
		conf.ShowUserContainer = !conf.ShowUserContainer

		if !conf.ShowUserContainer {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}

		config.PersistConfig()
		window.RefreshLayout()
		return nil
	}

	if event.Modifiers()&tcell.ModAlt == tcell.ModAlt {
		if event.Rune() == 'p' {
			window.SwitchToFriendsPage()
			window.app.SetFocus(window.privateList.GetComponent())
			return nil
		}

		if event.Rune() == 'c' {
			window.SwitchToGuildsPage()
			window.app.SetFocus(window.channelTree.internalTreeView)
			return nil
		}

		if event.Rune() == 's' {
			window.SwitchToGuildsPage()
			window.app.SetFocus(window.guildList)
			return nil
		}

		if event.Rune() == 't' {
			window.app.SetFocus(window.chatView.internalTextView)
			return nil
		}

		if event.Rune() == 'u' {
			if window.leftArea.GetCurrentPage() == guildPageName && window.userList.internalTreeView.IsVisible() {
				window.app.SetFocus(window.userList.internalTreeView)
			}
			return nil
		}

		if event.Rune() == 'm' {
			window.app.SetFocus(window.messageInput.GetPrimitive())
			return nil
		}
	}

	return event
}

//ExecuteCommand tries to execute the given input as a command. The first word
//will be passed as the commands name and the rest will be parameters. If a
//command can't be found, that info will be printed onto the command output.
func (window *Window) ExecuteCommand(command string) {
	//TODO Improve splitting for more complex stuff
	parts := strings.Split(command, " ")
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
	window.messageInput.SetBackgroundColor(tcell.ColorBlack)
}

//ShowErrorDialog shows a simple error dialog that has only an Okay button,
// a generic title and the given text.
func (window *Window) ShowErrorDialog(text string) {
	previousFocus := window.app.GetFocus()

	dialog := tview.NewModal()
	dialog.SetTitle("An error occurred")
	dialog.SetText(text)
	dialog.AddButtons([]string{"Okay"})

	dialog.SetDoneFunc(func(index int, label string) {
		window.app.SetRoot(window.rootContainer, true)
		window.app.SetFocus(previousFocus)
	})

	window.app.SetRoot(dialog, false)
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
		window.overrideShowUsers = true
		window.RefreshLayout()
	}
}

//SwitchToFriendsPage switches the left side of the layout over to the view
//where you can see your private chats and groups. In addition to that it
//hides the user list.
func (window *Window) SwitchToFriendsPage() {
	if window.leftArea.GetCurrentPage() != privatePageName {
		window.leftArea.SwitchToPage(privatePageName)
		window.overrideShowUsers = false
		window.RefreshLayout()
	}
}

//RefreshLayout removes and adds the main parts of the layout
//so that the ones that are disabled by settings do not show up.
func (window *Window) RefreshLayout() {
	conf := config.GetConfig()

	window.userList.internalTreeView.SetVisible(conf.ShowUserContainer && window.overrideShowUsers)
	window.channelTitle.SetVisible(conf.ShowChatHeader)

	if conf.UseFixedLayout {
		window.rootContainer.ResizeItem(window.leftArea, conf.FixedSizeLeft, 7)
		window.rootContainer.ResizeItem(window.chatArea, 0, 1)
		window.rootContainer.ResizeItem(window.userList.internalTreeView, conf.FixedSizeRight, 6)
	} else {
		window.rootContainer.ResizeItem(window.leftArea, 0, 7)
		window.rootContainer.ResizeItem(window.chatArea, 0, 20)
		window.rootContainer.ResizeItem(window.userList.internalTreeView, 0, 6)
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
		if cacheError != nil || len(cache.Messages) == 0 {
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
			messages = cache.Messages
		}
	} else {
		messages = channel.Messages
	}

	discordgoplus.SortMessagesByTimestamp(messages)

	window.chatView.SetMessages(messages)
	window.chatView.ClearSelection()
	window.chatView.internalTextView.ScrollToEnd()

	if channel.Topic != "" {
		window.channelTitle.SetText(channel.Name + " - " + channel.Topic)
	} else {
		window.channelTitle.SetText(channel.Name)
	}

	window.selectedChannel = channel
	if channel.GuildID == "" {
		if window.selectedChannelNode != nil {
			window.selectedChannelNode.SetColor(tcell.ColorWhite)
			window.selectedChannelNode = nil
		}

		if window.selectedGuildNode != nil {
			window.selectedGuildNode.SetColor(tcell.ColorWhite)
			window.selectedGuildNode = nil
		}
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
