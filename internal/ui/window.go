package ui

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/Bios-Marcel/cordless/internal/scripting"
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
	currentPage     string
	privateList     *tview.TreeView
	privateRootNode *tview.TreeNode

	channelRootNode *tview.TreeNode
	channelTitle    *tview.TextView

	chatArea                    *tview.Flex
	chatView                    *ChatView
	messageContainer            tview.Primitive
	messageInput                *Editor
	requestedMessageInputHeight int

	editingMessageID *string

	userList     *tview.TreeView
	userRootNode *tview.TreeNode

	overrideShowUsers bool

	killCurrentGuildUpdateThread *chan bool
	session                      *discordgo.Session

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
func NewWindow(discord *discordgo.Session) (*Window, error) {
	app := tview.NewApplication()

	window := Window{
		session:                     discord,
		app:                         app,
		commands:                    make(map[string]func(io.Writer, *Window, []string), 1),
		requestedMessageInputHeight: 3,
		scripting:                   scripting.New(),
	}

	if err := window.scripting.LoadScripts(config.GetScriptDirectory()); err != nil {
		return nil, err
	}

	guilds, discordError := discord.UserGuilds(100, "", "")
	if discordError != nil {
		return nil, discordError
	}

	window.leftArea = tview.NewPages()

	guildPage := tview.NewFlex()
	guildPage.SetDirection(tview.FlexRow)

	channelTree := tview.NewTreeView()
	channelTree.SetCycleSelection(true)
	channelRootNode := tview.NewTreeNode("")
	window.channelRootNode = channelRootNode
	channelTree.SetRoot(channelRootNode)
	channelTree.SetBorder(true)
	channelTree.SetTopLevel(1)

	guildList := tview.NewTreeView()
	guildList.SetCycleSelection(true)
	guildRootNode := tview.NewTreeNode("")
	guildList.SetRoot(guildRootNode)
	guildList.SetBorder(true)
	guildList.SetTopLevel(1)

	var selectedGuildNode *tview.TreeNode

	for _, tempGuild := range guilds {
		guild := tempGuild
		guildNode := tview.NewTreeNode(guild.Name)
		guildRootNode.AddChild(guildNode)
		guildNode.SetSelectable(true)
		guildNode.SetSelectedFunc(func() {
			if window.killCurrentGuildUpdateThread != nil {
				*window.killCurrentGuildUpdateThread <- true
			}

			if selectedGuildNode != nil {
				selectedGuildNode.SetColor(tcell.ColorWhite)
			}

			selectedGuildNode = guildNode
			selectedGuildNode.SetColor(tcell.ColorTeal)

			window.selectedGuild = guild
			channelRootNode.ClearChildren()

			channels, discordError := discord.GuildChannels(guild.ID)

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

			updateUser := time.NewTicker(userListUpdateInterval)
			go func() {
				killChan := make(chan bool)
				window.killCurrentGuildUpdateThread = &killChan
				if config.GetConfig().ShowUserContainer {
					window.UpdateUsersForGuild(guild)
				}
				for {
					select {
					case <-*window.killCurrentGuildUpdateThread:
						window.killCurrentGuildUpdateThread = nil
						return
					case <-updateUser.C:
						if config.GetConfig().ShowUserContainer {
							window.UpdateUsersForGuild(guild)
						}
					}
				}
			}()
		})
	}

	if len(guildRootNode.GetChildren()) > 0 {
		guildList.SetCurrentNode(guildRootNode)
	}

	guildPage.AddItem(guildList, 0, 1, true)
	guildPage.AddItem(channelTree, 0, 2, true)

	window.leftArea.AddPage(guildPageName, guildPage, true, false)

	window.privateList = tview.NewTreeView().
		SetCycleSelection(true).
		SetTopLevel(1)
	window.privateList.SetBorder(true)

	window.privateRootNode = tview.NewTreeNode("")
	window.privateList.SetRoot(window.privateRootNode)
	window.privateRootNode.SetSelectable(false)

	friendsNode := tview.NewTreeNode("Friends")
	groupChatsNode := tview.NewTreeNode("Groups")
	peopleChatsNode := tview.NewTreeNode("O2pen duo chats")

	window.privateRootNode.AddChild(friendsNode)
	window.privateRootNode.AddChild(groupChatsNode)
	window.privateRootNode.AddChild(peopleChatsNode)

	window.leftArea.AddPage(privatePageName, window.privateList, true, false)

	go func() {
		for _, channel := range window.session.State.PrivateChannels {
			if channel.Type == discordgo.ChannelTypeDM && len(channel.Recipients) > 0 {
				recipient := channel.Recipients[0].Username
				channelCopy := channel
				window.app.QueueUpdate(func() {
					newNode := tview.NewTreeNode(recipient)
					peopleChatsNode.AddChild(newNode)
					newNode.SetSelectedFunc(func() {
						window.LoadChannel(channelCopy)
						window.channelTitle.SetText(recipient)
					})
				})
			} else if channel.Type == discordgo.ChannelTypeGroupDM && len(channel.Recipients) > 0 {
				itemName := ""

				if channel.Name != "" {
					itemName = channel.Name
				} else {
					for index, recipient := range channel.Recipients {
						if index == 0 {
							itemName = recipient.Username
						} else {
							itemName = fmt.Sprintf("%s, %s", itemName, recipient.Username)
						}
					}
				}

				channelCopy := channel
				window.app.QueueUpdate(func() {
					newNode := tview.NewTreeNode(itemName)
					groupChatsNode.AddChild(newNode)
					newNode.SetSelectedFunc(func() {
						window.LoadChannel(channelCopy)
						window.channelTitle.SetText(itemName)
					})
				})
			}
		}

		window.app.QueueUpdate(func() {
			for _, friend := range window.session.State.Relationships {
				if friend.Type != discordgoplus.RelationTypeFriend {
					continue
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
	window.messageContainer = window.chatView.GetPrimitive()

	window.messageInput = NewEditor()
	window.messageInput.SetOnHeightChangeRequest(func(height int) {
		window.requestedMessageInputHeight = height
		window.RefreshLayout()
	})

	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		messageToSend := window.messageInput.GetText()

		if event.Key() == tcell.KeyUp && messageToSend == "" {
			for i := len(window.shownMessages) - 1; i > 0; i-- {
				message := window.shownMessages[i]
				if message.Author.ID == window.session.State.User.ID {
					window.messageInput.SetText(message.ContentWithMentionsReplaced())
					window.messageInput.SetBackgroundColor(tcell.ColorDarkGoldenrod)
					window.editingMessageID = &message.ID
					break
				}
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
							guild, discordError := window.session.State.Guild(window.selectedGuild.ID)
							if discordError == nil {
								for _, member := range guild.Members {
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
						dialog := tview.NewModal()
						dialog.SetText("Do you really want to delete the message?")
						dialog.AddButtons([]string{"Abort", "Delete"})
						dialog.SetDoneFunc(func(index int, label string) {
							if index == 1 {
								msgIDCopy := *window.editingMessageID
								go window.session.ChannelMessageDelete(window.selectedChannel.ID, msgIDCopy)
							}

							window.exitMessageEditMode()
							window.app.SetRoot(window.rootContainer, true)
							window.app.SetFocus(window.messageInput.GetPrimitive())
						})
						window.app.SetRoot(dialog, false)
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
								notificationLocation := channel.Name
								if notificationLocation == "" {
									notificationLocation = channel.Recipients[0].Username
								}
								beeep.Notify("Cordless - "+notificationLocation, message.ContentWithMentionsReplaced(), "assets/information.png")
							}
						}
					}

					window.app.QueueUpdateDraw(func() {

						window.channelRootNode.Walk(func(node, parent *tview.TreeNode) bool {
							data, ok := node.GetReference().(string)
							if ok && data == message.ChannelID {
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
	}()

	go func() {
		for {
			select {
			case messageDeleted := <-messageDeleteChan:
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
	window.channelTitle.SetBorder(true)

	window.chatArea.AddItem(window.channelTitle, 3, 1, true)
	window.chatArea.AddItem(window.messageContainer, 0, 1, true)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), 3, 0, true)

	window.commandView = NewCommandView(window.ExecuteCommand)

	window.userRootNode = tview.NewTreeNode("")
	window.userList = tview.NewTreeView().
		SetRoot(window.userRootNode).
		SetTopLevel(1).
		SetCycleSelection(true)
	window.userList.SetBorder(true)

	if config.GetConfig().OnTypeInListBehaviour == config.SearchOnTypeInList {
		var guildJumpBuffer string
		var guildJumpTime time.Time
		guildList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(guildList, guildRootNode, &guildJumpTime, &guildJumpBuffer))
		var channelJumpBuffer string
		var channelJumpTime time.Time
		channelTree.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(channelTree, channelRootNode, &channelJumpTime, &channelJumpBuffer))
		var userJumpBuffer string
		var userJumpTime time.Time
		window.userList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(window.userList, window.userRootNode, &userJumpTime, &userJumpBuffer))
		var privateJumpBuffer string
		var privateJumpTime time.Time
		window.privateList.SetInputCapture(treeview.CreateSearchOnTypeInuptHandler(window.privateList, window.privateRootNode, &privateJumpTime, &privateJumpBuffer))
	}

	window.rootContainer = tview.NewFlex().
		SetDirection(tview.FlexColumn)
	window.rootContainer.SetTitleAlign(tview.AlignCenter)

	app.SetRoot(window.rootContainer, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '.' &&
			(event.Modifiers()&tcell.ModAlt) == tcell.ModAlt {
			if window.commandMode {
				window.commandMode = false
			} else {
				window.commandMode = true
			}

			window.RefreshLayout()

			if window.commandMode {
				app.SetFocus(window.commandView.commandInput)
			}

			return nil
		}

		if window.commandMode && event.Key() == tcell.KeyCtrlO {
			app.SetFocus(window.commandView.commandOutput)
		}

		if window.commandMode && event.Key() == tcell.KeyCtrlI {
			app.SetFocus(window.commandView.commandInput)
		}

		if event.Rune() == 'U' &&
			(event.Modifiers()&tcell.ModAlt) == tcell.ModAlt {
			conf := config.GetConfig()
			conf.ShowUserContainer = !conf.ShowUserContainer
			config.PersistConfig()
			window.RefreshLayout()
			return nil
		}

		if event.Modifiers()&tcell.ModAlt == tcell.ModAlt {
			if event.Rune() == 'f' {
				window.SwitchToFriendsPage()
				app.SetFocus(window.privateList)
				return nil
			}

			if event.Rune() == 'c' {
				window.SwitchToGuildsPage()
				app.SetFocus(channelTree)
				return nil
			}

			if event.Rune() == 's' {
				window.SwitchToGuildsPage()
				app.SetFocus(guildList)
				return nil
			}

			if event.Rune() == 't' {
				app.SetFocus(window.messageContainer)
				return nil
			}

			if event.Rune() == 'u' {
				if window.currentPage == guildPageName {
					app.SetFocus(window.userList)
				}
				return nil
			}

			if event.Rune() == 'm' {
				app.SetFocus(window.messageInput.GetPrimitive())
				return nil
			}
		}

		return event
	})

	window.SwitchToGuildsPage()

	app.SetFocus(guildList)

	return &window, nil
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
		updatedMessage, discordError := window.session.ChannelMessageEdit(channelID, messageID, messageEdited)
		if discordError == nil {
			for index, msg := range window.shownMessages {
				if msg.ID == updatedMessage.ID {
					window.shownMessages[index] = updatedMessage
					break
				}
			}
		}
		window.app.QueueUpdateDraw(func() {
			window.SetMessages(window.shownMessages)
		})
	}()

	window.exitMessageEditMode()
}

//SwitchToGuildsPage the left side of the layout over to the view where you can
//see the servers and their channels. In additional to that, it also shows the
//user list in case the user didn't explicitly hide it.
func (window *Window) SwitchToGuildsPage() {
	if window.currentPage != guildPageName {
		window.currentPage = guildPageName
		window.leftArea.SwitchToPage(guildPageName)
		window.overrideShowUsers = true
		window.RefreshLayout()
	}
}

//SwitchToFriendsPage switches the left side of the layout over to the view
//where you can see your private chats and groups. In addition to that it
//hides the user list.
func (window *Window) SwitchToFriendsPage() {
	if window.currentPage != privatePageName {
		window.currentPage = privatePageName
		window.leftArea.SwitchToPage(privatePageName)
		window.overrideShowUsers = false
		window.RefreshLayout()
	}
}

//RefreshLayout removes and adds the main parts of the layout
//so that the ones that are disabled by settings do not show up.
func (window *Window) RefreshLayout() {
	window.rootContainer.RemoveItem(window.leftArea)
	window.rootContainer.RemoveItem(window.chatArea)
	window.rootContainer.RemoveItem(window.userList)

	conf := config.GetConfig()

	if conf.UseFixedLayout {
		window.rootContainer.AddItem(window.leftArea, conf.FixedSizeLeft, 7, true)
		window.rootContainer.AddItem(window.chatArea, 0, 1, false)

		if conf.ShowUserContainer && window.overrideShowUsers {
			window.rootContainer.AddItem(window.userList, conf.FixedSizeRight, 6, false)
		}
	} else {
		window.rootContainer.AddItem(window.leftArea, 0, 7, true)

		if conf.ShowUserContainer && window.overrideShowUsers {
			window.rootContainer.AddItem(window.chatArea, 0, 20, false)
			window.rootContainer.AddItem(window.userList, 0, 6, false)
		} else {
			window.rootContainer.AddItem(window.chatArea, 0, 26, false)
		}
	}

	window.chatArea.RemoveItem(window.channelTitle)
	window.chatArea.RemoveItem(window.messageContainer)
	window.chatArea.RemoveItem(window.messageInput.GetPrimitive())
	window.chatArea.RemoveItem(window.commandView.commandInput)
	window.chatArea.RemoveItem(window.commandView.commandOutput)

	if conf.ShowChatHeader {
		window.chatArea.AddItem(window.channelTitle, 3, 0, false)
	}

	window.chatArea.AddItem(window.messageContainer, 0, 1, false)
	window.chatArea.AddItem(window.messageInput.GetPrimitive(), window.requestedMessageInputHeight, 0, false)

	if window.commandMode {
		window.chatArea.AddItem(window.commandView.commandOutput, 0, 1, false)
		window.chatArea.AddItem(window.commandView.commandInput, 3, 0, false)
	}

	if conf.ShowFrame {
		window.rootContainer.SetTitle("Cordless")
		window.rootContainer.SetBorder(true)
	} else {
		window.rootContainer.SetTitle("")
		window.rootContainer.SetBorder(false)
	}

	window.app.ForceDraw()
}

//LoadChannel eagerly loads the channels messages.
func (window *Window) LoadChannel(channel *discordgo.Channel) error {
	messages, discordError := window.session.ChannelMessages(channel.ID, 100, "", "", "")
	if discordError != nil {
		return discordError
	}

	if messages != nil && len(messages) > 0 {
		//HACK: Reversing them, as they are sorted anyway.
		msgAmount := len(messages)
		for i := 0; i < msgAmount/2; i++ {
			j := msgAmount - i - 1
			messages[i], messages[j] = messages[j], messages[i]
		}

		window.SetMessages(messages)
	} else {
		window.SetMessages(make([]*discordgo.Message, 0))
	}

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

func (window *Window) UpdateUsersForGuild(guild *discordgo.UserGuild) {
	guildRefreshed, discordError := window.session.Guild(guild.ID)
	//TODO Handle error
	if discordError != nil {
		return
	}

	discordError = window.session.State.GuildAdd(guildRefreshed)
	//TODO Handle error
	if discordError != nil {
		return
	}

	guildState, discordError := window.session.State.Guild(guildRefreshed.ID)
	//TODO Handle error
	if discordError != nil {
		return
	}

	users := guildState.Members
	/*users := make([]*discordgo.Member, 0)

	for _, user := range usersUnfiltered {
		if true {
			users = append(users, user)
			continue USER_MATCHED
		}
	}*/

	roles := guildState.Roles

	sort.Slice(roles, func(a, b int) bool {
		return roles[a].Position > roles[b].Position
	})

	window.app.QueueUpdateDraw(func() {
		window.userRootNode.ClearChildren()

		roleNodes := make(map[string]*tview.TreeNode)

		for _, role := range roles {
			if role.Hoist {
				roleNode := tview.NewTreeNode(role.Name)
				roleNode.SetSelectable(false)
				roleNodes[role.ID] = roleNode
				window.userRootNode.AddChild(roleNode)
			}
		}

		nonHoistNode := tview.NewTreeNode("No Hoist Role")
		nonHoistNode.SetSelectable(false)
		window.userRootNode.AddChild(nonHoistNode)

	USER:
		for _, user := range users {

			var nameToUse string

			if user.Nick != "" {
				nameToUse = user.Nick
			} else {
				nameToUse = user.User.Username
			}

			userNode := tview.NewTreeNode(nameToUse)

			sort.Slice(user.Roles, func(a, b int) bool {
				firstIdentifier := user.Roles[a]
				secondIdentifier := user.Roles[b]

				var firstRole *discordgo.Role
				var secondRole *discordgo.Role
				for _, role := range roles {
					if role.ID == firstIdentifier {
						firstRole = role
					} else if role.ID == secondIdentifier {
						secondRole = role
					}
				}

				return firstRole.Position > secondRole.Position
			})

			for _, userRole := range user.Roles {
				roleNode, exists := roleNodes[userRole]
				if exists {
					roleNode.AddChild(userNode)
					continue USER
				}
			}

			nonHoistNode.AddChild(userNode)
		}

		if window.userList.GetCurrentNode() == nil {
			userNodes := window.userRootNode.GetChildren()
			if userNodes != nil && len(userNodes) > 0 {
				window.userList.SetCurrentNode(window.userRootNode.GetChildren()[0])
			}
		}
	})
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
