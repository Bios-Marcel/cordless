package ui

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/Bios-Marcel/cordless/internal/maths"
	"github.com/Bios-Marcel/cordless/internal/scripting"
	"github.com/Bios-Marcel/cordless/internal/times"
	"github.com/Bios-Marcel/cordless/internal/ui/tview/treeview"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
	"github.com/gen2brain/beeep"
	"github.com/kyokomi/emoji"
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

	leftArea        *tview.Pages
	guildList       *tview.TreeView
	channelTree     *tview.TreeView
	privateList     *tview.TreeView
	privateRootNode *tview.TreeNode

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

	shownMessages       []*discordgo.Message
	selectedGuild       *discordgo.UserGuild
	selectedChannelNode *tview.TreeNode
	selectedChannel     *discordgo.Channel

	scripting scripting.Engine

	commandMode bool
	commandView *CommandView
	commands    map[string]func(io.Writer, *Window, []string)
}

//NewWindow constructs the whole application window and also registers all
//necessary handlers and functions. If this function returns an error, we can't
//start the application.
func NewWindow(app *tview.Application, discord *discordgo.Session) (*Window, error) {
	window := Window{
		session:   discord,
		app:       app,
		commands:  make(map[string]func(io.Writer, *Window, []string), 1),
		scripting: scripting.New(),
	}

	if err := window.scripting.LoadScripts(config.GetScriptDirectory()); err != nil {
		return nil, err
	}

	//FIXME Bug: If you are in more than 100 guilds, you won't see a lot of them!
	guilds, discordError := discord.UserGuilds(100, "", "")
	if discordError != nil {
		return nil, discordError
	}

	mentionWindow := tview.NewTreeView()
	mentionWindow.SetCycleSelection(true)

	window.leftArea = tview.NewPages()

	guildPage := tview.NewFlex()
	guildPage.SetDirection(tview.FlexRow)

	channelTree := tview.NewTreeView().
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true)
	window.channelTree = channelTree

	channelRootNode := tview.NewTreeNode("")
	window.channelRootNode = channelRootNode
	channelTree.SetRoot(channelRootNode)
	channelTree.SetBorder(true)
	channelTree.SetTopLevel(1)

	guildList := tview.NewTreeView().
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true)
	window.guildList = guildList

	guildRootNode := tview.NewTreeNode("")
	guildList.SetRoot(guildRootNode)
	guildList.SetBorder(true)
	guildList.SetTopLevel(1)

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

	var selectedGuildNode *tview.TreeNode

	for _, tempGuild := range guilds {
		guild := tempGuild
		guildNode := tview.NewTreeNode(guild.Name)
		guildRootNode.AddChild(guildNode)
		guildNode.SetSelectable(true)
		guildNode.SetSelectedFunc(func() {
			if selectedGuildNode != nil {
				selectedGuildNode.SetColor(tcell.ColorWhite)
			}

			selectedGuildNode = guildNode
			selectedGuildNode.SetColor(tcell.ColorTeal)

			window.selectedGuild = guild
			channelRootNode.ClearChildren()

			channels, discordError := discord.GuildChannels(guild.ID)

			discord.RequestGuildMembers(guild.ID, "", 0)

			if discordError != nil {
				window.ShowErrorDialog(fmt.Sprintf("An error occurred while trying to receive the channels: %s", discordError.Error()))
				//TODO Is returning here a good idea?
				return
			}

			sort.Slice(channels, func(a, b int) bool {
				return channels[a].Position < channels[b].Position
			})

			registerChannelForChatting := func(node *tview.TreeNode, channelToConnectTo *discordgo.Channel) {
				node.SetSelectable(true)
				node.SetSelectedFunc(func() {
					discordError := window.LoadChannel(channelToConnectTo)
					if discordError != nil {
						errorMessage := fmt.Sprintf("An error occurred while trying to load the channel '%s': %s", channelToConnectTo.Name, discordError.Error())
						window.ShowErrorDialog(errorMessage)
						return
					}

					if window.selectedChannelNode != nil {
						//For some reason using tcell.ColorDefault causes hovering to render incorrect.
						window.selectedChannelNode.SetColor(tcell.ColorWhite)
					}

					window.selectedChannelNode = node
					node.SetText(channelToConnectTo.Name)
					node.SetColor(tcell.ColorTeal)
				})
			}

			createNodeForChannel := func(channel *discordgo.Channel) *tview.TreeNode {
				nodeName := channel.Name
				if channel.NSFW {
					nodeName = nodeName + " NSFW"
				}

				return tview.NewTreeNode(nodeName)
			}

			channelCategories := make(map[string]*tview.TreeNode)
			for _, channel := range channels {
				if channel.ParentID == "" {
					newNode := createNodeForChannel(channel)
					channelRootNode.AddChild(newNode)

					if channel.Type == discordgo.ChannelTypeGuildCategory {
						//Categories
						newNode.SetSelectable(false)
						channelCategories[channel.ID] = newNode
					} else {
						//Toplevel channels
						newNode.SetReference(channel.ID)
						registerChannelForChatting(newNode, channel)
					}
				}
			}

			//Channels that are in categories
			for _, channel := range channels {
				if channel.Type == discordgo.ChannelTypeGuildText && channel.ParentID != "" {
					newNode := createNodeForChannel(channel)
					newNode.SetReference(channel.ID)
					registerChannelForChatting(newNode, channel)
					channelCategories[channel.ParentID].AddChild(newNode)
				}
			}

			//No selection will prevent selection from working at all.
			if len(window.channelRootNode.GetChildren()) > 0 {
				channelTree.SetCurrentNode(window.channelRootNode)
			}

			if config.GetConfig().FocusChannelAfterGuildSelection {
				window.app.SetFocus(channelTree)
			}

			loadError := window.userList.LoadGuild(guild.ID)
			if loadError != nil {
				window.ShowErrorDialog(loadError.Error())
			}
		})
	}

	if len(guildRootNode.GetChildren()) > 0 {
		guildList.SetCurrentNode(guildRootNode)
	}

	guildPage.AddItem(guildList, 0, 1, true)
	guildPage.AddItem(channelTree, 0, 2, true)

	window.leftArea.AddPage(guildPageName, guildPage, true, false)

	window.privateList = tview.NewTreeView().
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true).
		SetTopLevel(1)
	window.privateList.SetBorder(true)

	window.privateRootNode = tview.NewTreeNode("")
	window.privateList.SetRoot(window.privateRootNode)
	window.privateRootNode.SetSelectable(false)

	privateChatsNode := tview.NewTreeNode("Chats").
		SetSelectable(false)
	friendsNode := tview.NewTreeNode("Friends").
		SetSelectable(false)

	window.privateRootNode.AddChild(privateChatsNode)
	window.privateRootNode.AddChild(friendsNode)

	window.leftArea.AddPage(privatePageName, window.privateList, true, false)

	go func() {
		privateChannels := make([]*discordgo.Channel, len(window.session.State.PrivateChannels))
		copy(privateChannels, window.session.State.PrivateChannels)
		sort.Slice(privateChannels, func(a, b int) bool {
			channelA := privateChannels[a]
			channelB := privateChannels[b]

			messageA, parseError := strconv.ParseInt(channelA.LastMessageID, 10, 64)
			if parseError != nil {
				return false
			}

			messageB, parseError := strconv.ParseInt(channelB.LastMessageID, 10, 64)
			if parseError != nil {
				return true
			}

			return messageA > messageB
		})

		window.app.QueueUpdate(func() {
			for _, channel := range privateChannels {
				channelName := discordgoplus.GetPrivateChannelName(channel)
				channelCopy := channel
				newNode := tview.NewTreeNode(channelName)

				privateChatsNode.AddChild(newNode)
				newNode.SetSelectedFunc(func() {
					window.LoadChannel(channelCopy)
					window.channelTitle.SetText(channelName)
					if channelCopy.Type == discordgo.ChannelTypeDM {
						window.overrideShowUsers = false
						window.RefreshLayout()
					} else if channelCopy.Type == discordgo.ChannelTypeGroupDM {
						window.overrideShowUsers = true
						loadError := window.userList.LoadGroup(channelCopy.ID)
						if loadError != nil {
							fmt.Fprintln(window.commandView.commandOutput, "Error loading users for channel.")
						}

						window.RefreshLayout()
					}
				})
			}

		FRIEND_LOOP:
			for _, friend := range window.session.State.Relationships {
				if friend.Type != discordgoplus.RelationTypeFriend {
					continue
				}

				for _, channel := range privateChannels {
					if channel.Type != discordgo.ChannelTypeDM {
						continue
					}

					if channel.Recipients[0].ID == friend.ID ||
						(len(channel.Recipients) > 1 && channel.Recipients[1].ID == friend.ID) {
						continue FRIEND_LOOP
					}
				}

				newNode := tview.NewTreeNode(friend.User.Username)
				friendsNode.AddChild(newNode)

				friendCopy := friend.User
				newNode.SetSelectedFunc(func() {
					userChannels, _ := window.session.UserChannels()
					for _, userChannel := range userChannels {
						if userChannel.Type == discordgo.ChannelTypeDM &&
							(userChannel.Recipients[0].ID == friendCopy.ID) {
							window.LoadChannel(userChannel)
							window.channelTitle.SetText(newNode.GetText())
							return
						}
					}

					newChannel, discordError := window.session.UserChannelCreate(friendCopy.ID)
					if discordError == nil {
						window.LoadChannel(newChannel)
						window.channelTitle.SetText(newChannel.Recipients[0].Username)
					}
				})
			}
			if len(window.privateRootNode.GetChildren()) > 0 {
				window.privateList.SetCurrentNode(window.privateRootNode)
			}
		})
	}()

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

	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		messageToSend := window.messageInput.GetText()

		if event.Key() == tcell.KeyUp && messageToSend == "" {
			for i := len(window.shownMessages) - 1; i > 0; i-- {
				message := window.shownMessages[i]
				window.startEditingMessage(message)
			}

			return nil
		}

		if event.Key() == tcell.KeyEsc {
			window.exitMessageEditMode()
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

					//Replace formatter characters and replace emoji codes.
					messageToSend = emoji.Sprintf(strings.Replace(messageToSend, "%", "%%", -1))

					if strings.Contains(messageToSend, "@") {
						messageToSend = mentionRegex.
							ReplaceAllStringFunc(messageToSend, func(part string) string {
								return strings.ToLower(part)
							})

						if window.selectedGuild != nil {
							members, discordError := window.session.State.Members(window.selectedGuild.ID)
							if discordError == nil {
								for _, member := range members {
									if member.Nick != "" {
										messageToSend = strings.Replace(messageToSend, "@"+strings.ToLower(member.Nick), "<@"+member.User.ID+">", -1)
									}

									messageToSend = strings.Replace(messageToSend, "@"+strings.ToLower(member.User.Username), "<@"+member.User.ID+">", -1)
								}
							}
						} else if window.selectedChannel != nil {
							for _, user := range window.selectedChannel.Recipients {
								messageToSend = strings.Replace(messageToSend, "@"+strings.ToLower(user.Username), "<@"+user.ID+">", -1)
							}
						}
					}

					if window.editingMessageID != nil {
						go window.editMessage(window.selectedChannel.ID, *window.editingMessageID, messageToSend)
						window.exitMessageEditMode()
					} else {
						go func() {
							_, sendError := discord.ChannelMessageSend(window.selectedChannel.ID, window.scripting.OnMessageSend(messageToSend))
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

	messageInputChan := make(chan *discordgo.Message, 50)
	messageDeleteChan := make(chan *discordgo.Message, 50)
	messageEditChan := make(chan *discordgo.Message, 50)

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if window.selectedChannel != nil {
			messageInputChan <- m.Message
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		if window.selectedChannel != nil {
			if m.ChannelID == window.selectedChannel.ID {
				messageDeleteChan <- m.Message
			}
		}
	})

	window.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if window.selectedChannel != nil {
			if m.ChannelID == window.selectedChannel.ID &&
				//Ignore just-embed edits
				m.Content != "" {
				messageEditChan <- m.Message
			}
		}
	})

	go func() {
		for {
			select {
			case message := <-messageInputChan:
				//UPDATE CACHE
				window.session.State.MessageAdd(message)

				if message.ChannelID == window.selectedChannel.ID {
					window.app.QueueUpdateDraw(func() {
						window.AddMessages([]*discordgo.Message{message})
					})
				} else {
					mentionsYou := false
					if message.Author.ID != window.session.State.User.ID {
						for _, user := range message.Mentions {
							if user.ID == window.session.State.User.ID {
								mentionsYou = true
								break
							}
						}

						channel, stateError := window.session.State.Channel(message.ChannelID)
						if stateError == nil {
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

									notificationLocation = message.Author.Username + "-" + notificationLocation
								} else if channel.Type == discordgo.ChannelTypeGuildText {
									notificationLocation = message.Author.Username + "-" + channel.Name
								}

								beeep.Notify("Cordless - "+notificationLocation, message.ContentWithMentionsReplaced(), "assets/information.png")
							}
						}

						window.app.QueueUpdateDraw(func() {
							window.channelRootNode.Walk(func(node, parent *tview.TreeNode) bool {
								data, ok := node.GetReference().(string)
								if ok && data == message.ChannelID && window.selectedChannel.ID != data {
									if mentionsYou {
										channel, stateError := window.session.State.Channel(message.ChannelID)
										if stateError == nil {
											node.SetText("(@You) " + channel.Name)
										}
									}

									node.SetColor(tcell.ColorRed)
									return false
								}
								return true
							})
						})
					}
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case messageDeleted := <-messageDeleteChan:
				//UPDATE CACHE
				window.session.State.MessageRemove(messageDeleted)
				for index, message := range window.shownMessages {
					if message.ID == messageDeleted.ID {
						window.app.QueueUpdateDraw(func() {
							window.SetMessages(append(window.shownMessages[:index], window.shownMessages[index+1:]...))
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
			case messageEdited := <-messageEditChan:
				//UPDATE CACHE
				window.session.State.MessageAdd(messageEdited)
				for _, message := range window.shownMessages {
					if message.ID == messageEdited.ID {
						message.Content = messageEdited.Content
						window.app.QueueUpdateDraw(func() {
							window.SetMessages(window.shownMessages)
						})
						break
					}
				}
			}
		}
	}()

	window.channelTitle = tview.NewTextView()
	window.channelTitle.SetBorderSides(true, true, false, true)
	window.channelTitle.SetBorder(true)

	window.commandView = NewCommandView(window.ExecuteCommand)

	window.userList = NewUserTree(window.session.State)

	if config.GetConfig().OnTypeInListBehaviour == config.SearchOnTypeInList {
		var guildJumpBuffer string
		var guildJumpTime time.Time
		guildList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(
			guildList, guildRootNode, &guildJumpTime, &guildJumpBuffer))
		var channelJumpBuffer string
		var channelJumpTime time.Time
		channelTree.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(
			channelTree, channelRootNode, &channelJumpTime, &channelJumpBuffer))
		var userJumpBuffer string
		var userJumpTime time.Time
		window.userList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(
			window.userList.internalTreeView, window.userList.rootNode, &userJumpTime, &userJumpBuffer))
		var privateJumpBuffer string
		var privateJumpTime time.Time
		window.privateList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(
			window.privateList, window.privateRootNode, &privateJumpTime, &privateJumpBuffer))
	} else if config.GetConfig().OnTypeInListBehaviour == config.FocusMessageInputOnTypeInList {
		guildList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			guildList.Box, window.app, window.messageInput.internalTextView))
		channelTree.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			channelTree.Box, window.app, window.messageInput.internalTextView))
		window.userList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			window.userList.internalTreeView.Box, window.app, window.messageInput.internalTextView))
		window.privateList.SetInputCapture(treeview.CreateFocusTextViewOnTypeInputHandler(
			window.privateList.Box, window.app, window.messageInput.internalTextView))
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

	window.chatArea.AddItem(window.channelTitle, 2, 0, false)
	window.chatArea.AddItem(window.messageContainer, 0, 1, false)
	window.chatArea.AddItem(mentionWindow, 2, 2, true)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), window.messageInput.GetRequestedHeight(), 0, false)

	window.commandView.commandOutput.SetVisible(false)
	window.commandView.commandInput.SetVisible(false)

	window.chatArea.AddItem(window.commandView.commandOutput, 0, 1, false)
	window.chatArea.AddItem(window.commandView.commandInput, 3, 0, false)

	if conf.ShowFrame {
		window.rootContainer.SetTitle("Cordless")
		window.rootContainer.SetBorder(true)
	} else {
		window.rootContainer.SetTitle("")
		window.rootContainer.SetBorder(false)
	}

	window.SwitchToGuildsPage()

	app.SetFocus(guildList)

	return &window, nil
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

func (window *Window) handleGlobalShortcuts(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == '.' &&
		(event.Modifiers()&tcell.ModAlt) == tcell.ModAlt {

		window.commandMode = !window.commandMode

		if window.commandMode {
			window.app.SetFocus(window.commandView.commandInput)
		} else {
			window.app.SetFocus(window.messageInput.GetPrimitive())
		}

		window.commandView.SetVisible(window.commandMode)

		return nil
	}

	if window.commandMode && event.Key() == tcell.KeyCtrlO {
		if window.commandView.commandOutput.IsVisible() {
			window.app.SetFocus(window.commandView.commandOutput)
		}
	}

	if window.commandMode && event.Key() == tcell.KeyCtrlI {
		if window.commandView.commandInput.IsVisible() {
			window.app.SetFocus(window.commandView.commandInput)
		}
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
		if event.Rune() == 'f' {
			window.SwitchToFriendsPage()
			window.app.SetFocus(window.privateList)
			return nil
		}

		if event.Rune() == 'c' {
			window.SwitchToGuildsPage()
			window.app.SetFocus(window.channelTree)
			return nil
		}

		if event.Rune() == 's' {
			window.SwitchToGuildsPage()
			window.app.SetFocus(window.guildList)
			return nil
		}

		if event.Rune() == 't' {
			window.app.SetFocus(window.messageContainer)
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
	parts := strings.Split(command, " ")
	commandLogic, exists := window.commands[parts[0]]
	if exists {
		commandLogic(window.commandView.commandOutput, window, parts[1:])
	} else {
		fmt.Fprintf(window.commandView.commandOutput, "The command '%s' doesn't exist\n", parts[0])
	}
}

func (window *Window) startEditingMessage(message *discordgo.Message) {
	if message.Author.ID == window.session.State.User.ID {
		window.messageInput.SetText(message.ContentWithMentionsReplaced())
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

	sort.Slice(messages, func(a, b int) bool {
		timeA, parseError := messages[a].Timestamp.Parse()
		if parseError != nil {
			return false
		}

		timeB, parseError := messages[b].Timestamp.Parse()
		if parseError != nil {
			return true
		}

		return timeA.Before(timeB)
	})

	window.SetMessages(messages)
	window.chatView.ClearSelection()
	window.chatView.internalTextView.ScrollToEnd()

	if channel.Topic != "" {
		window.channelTitle.SetText(channel.Name + " - " + channel.Topic)
	} else {
		window.channelTitle.SetText(channel.Name)
	}

	window.selectedChannel = channel
	window.exitMessageEditModeAndKeepText()

	if config.GetConfig().FocusMessageInputAfterChannelSelection {
		window.app.SetFocus(window.messageInput.internalTextView)
	}

	return nil
}

//AddMessages adds the passed array of messages to the chat.
func (window *Window) AddMessages(messages []*discordgo.Message) {
	window.shownMessages = append(window.shownMessages, messages...)
	window.chatView.AddMessages(messages)
}

//SetMessages clears the current chat and adds the passed messages.s
func (window *Window) SetMessages(messages []*discordgo.Message) {
	window.shownMessages = messages
	window.chatView.SetMessages(window.shownMessages)
}

//RegisterCommand register a command. That makes the command available for
//being called from the message input field, in case the user-defined prefix
//is in front of the input.
func (window *Window) RegisterCommand(name string, logic func(writer io.Writer, window *Window, parameters []string)) {
	window.commands[name] = logic
}

//Run Shows the window optionally returning an error.
func (window *Window) Run() error {
	return window.app.Run()
}
