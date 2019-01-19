package ui

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/bwmarrin/discordgo"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	userListUpdateInterval = 5 * time.Second

	guildPageName   = "Guilds"
	friendsPageName = "Friends"
)

type Window struct {
	app           *tview.Application
	rootContainer *tview.Flex

	leftArea        *tview.Pages
	currentPage     string
	friendsList     *tview.TreeView
	friendsRootNode *tview.TreeNode

	channelRootNode *tview.TreeNode
	channelTitle    *tview.TextView

	chatArea         *tview.Flex
	chatView         *ChatView
	messageContainer tview.Primitive
	messageInput     *tview.InputField

	editingMessageID *string

	userContainer *tview.TreeView
	userRootNode  *tview.TreeNode

	overrideShowUsers bool

	killCurrentGuildUpdateThread *chan bool
	session                      *discordgo.Session

	shownMessages       []*discordgo.Message
	selectedGuild       *discordgo.UserGuild
	selectedChannelNode *tview.TreeNode

	selectedChannel *discordgo.Channel
}

func NewWindow(discord *discordgo.Session) (*Window, error) {
	app := tview.NewApplication()

	window := Window{
		session: discord,
		app:     app,
	}

	guilds, discordError := discord.UserGuilds(100, "", "")
	if discordError != nil {
		return nil, discordError
	}

	window.leftArea = tview.NewPages()

	guildPage := tview.NewFlex()
	guildPage.SetDirection(tview.FlexRow)

	channelTree := tview.NewTreeView()
	channelRootNode := tview.NewTreeNode("")
	window.channelRootNode = channelRootNode
	channelTree.SetRoot(channelRootNode)
	channelTree.SetBorder(true)
	channelTree.SetTopLevel(1)

	guildList := tview.NewTreeView()
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

			//TODO Handle error
			channels, _ := discord.GuildChannels(guild.ID)

			sort.Slice(channels, func(a, b int) bool {
				return channels[a].Position < channels[b].Position
			})

			channelCategories := make(map[string]*tview.TreeNode)
			for _, channel := range channels {
				if channel.ParentID == "" {
					newNode := tview.NewTreeNode(channel.Name)
					channelRootNode.AddChild(newNode)

					if channel.Type == discordgo.ChannelTypeGuildCategory {
						newNode.SetSelectable(false)
						channelCategories[channel.ID] = newNode
					}
				}
			}

			for _, channel := range channels {
				if channel.Type == discordgo.ChannelTypeGuildText && channel.ParentID != "" {
					nodeName := channel.Name
					if channel.NSFW {
						nodeName = nodeName + " NSFW"
					}
					newNode := tview.NewTreeNode(nodeName)

					newNode.SetSelectable(true)
					//This copy is necessary in order to use the correct channel instead
					//of always the same one.
					channelToConnectTo := channel
					newNode.SetSelectedFunc(func() {
						if window.selectedChannelNode != nil {
							//For some reason using tcell.ColorDefault causes hovering to render incorrect.
							window.selectedChannelNode.SetColor(tcell.ColorWhite)
						}

						window.selectedChannelNode = newNode

						newNode.SetColor(tcell.ColorTeal)
						discordError := window.LoadChannel(channelToConnectTo)
						if discordError != nil {
							log.Fatalf("Error loading messages (%s).", discordError.Error())
						}
					})

					channelCategories[channelToConnectTo.ParentID].AddChild(newNode)
				}
			}

			//No selection will prevent selection from working at all.
			if len(window.channelRootNode.GetChildren()) > 0 {
				channelTree.SetCurrentNode(window.channelRootNode)
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

	window.friendsList = tview.NewTreeView()
	window.friendsList.SetBorder(true)
	window.friendsList.SetTopLevel(1)

	window.friendsRootNode = tview.NewTreeNode("")
	window.friendsList.SetRoot(window.friendsRootNode)
	window.friendsRootNode.SetSelectable(false)

	friendsNode := tview.NewTreeNode("Friends")
	groupChatsNode := tview.NewTreeNode("Groups")
	peopleChatsNode := tview.NewTreeNode("Open duo chats")

	window.friendsRootNode.AddChild(friendsNode)
	window.friendsRootNode.AddChild(groupChatsNode)
	window.friendsRootNode.AddChild(peopleChatsNode)

	window.leftArea.AddPage(friendsPageName, window.friendsList, true, false)

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

			if len(window.friendsRootNode.GetChildren()) > 0 {
				window.friendsList.SetCurrentNode(window.friendsRootNode)
			}
		})
	}()

	window.chatArea = tview.NewFlex()
	window.chatArea.SetDirection(tview.FlexRow)

	window.chatView = NewChatView()
	window.messageContainer = window.chatView.GetPrimitive()

	window.messageInput = tview.NewInputField()
	window.messageInput.SetBorder(true)

	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp && window.messageInput.GetText() == "" {
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
				messageToSend := window.messageInput.GetText()
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

							for _, member := range guild.Members {
								if member.Nick != "" {
									messageToSend = strings.Replace(messageToSend, "@"+member.Nick, "<@"+member.User.ID+">", -1)
								}
								messageToSend = strings.Replace(messageToSend, "@"+member.User.Username, "<@"+member.User.ID+">", -1)
							}
						}
					}

					if window.editingMessageID != nil {
						go window.editMessage(window.selectedChannel.ID, *window.editingMessageID, messageToSend)
						window.exitMessageEditMode()
					} else {
						go discord.ChannelMessageSend(window.selectedChannel.ID, messageToSend)
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
							window.app.SetFocus(window.messageInput)
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
			if m.ChannelID == window.selectedChannel.ID {
				messageInputChan <- m.Message
			}
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
			if m.ChannelID == window.selectedChannel.ID {
				messageEditChan <- m.Message
			}
		}
	})

	go func() {
		for {
			select {
			case message := <-messageInputChan:
				window.app.QueueUpdateDraw(func() {
					window.SetMessages(append(window.shownMessages, message))
				})
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
	window.chatArea.AddItem(window.messageInput, 3, 0, true)

	window.userContainer = tview.NewTreeView()
	window.userRootNode = tview.NewTreeNode("")
	window.userContainer.SetTopLevel(1)
	window.userContainer.SetRoot(window.userRootNode)
	window.userContainer.SetBorder(true)

	window.rootContainer = tview.NewFlex()
	window.rootContainer.SetDirection(tview.FlexColumn)
	window.rootContainer.SetTitleAlign(tview.AlignCenter)

	app.SetRoot(window.rootContainer, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
				app.SetFocus(window.friendsList)
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
				app.SetFocus(window.userContainer)
				return nil
			}

			if event.Rune() == 'm' {
				app.SetFocus(window.messageInput)
				return nil
			}
		}

		return event
	})

	window.SwitchToGuildsPage()

	app.SetFocus(guildList)

	return &window, nil
}

func (window *Window) exitMessageEditMode() {
	window.editingMessageID = nil
	window.messageInput.SetBackgroundColor(tcell.ColorDefault)
	window.messageInput.SetText("")
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

func (window *Window) SwitchToGuildsPage() {
	if window.currentPage != guildPageName {
		window.currentPage = guildPageName
		window.leftArea.SwitchToPage(guildPageName)
		window.overrideShowUsers = true
		window.RefreshLayout()
	}
}

func (window *Window) SwitchToFriendsPage() {
	if window.currentPage != friendsPageName {
		window.currentPage = friendsPageName
		window.leftArea.SwitchToPage(friendsPageName)
		window.overrideShowUsers = false
		window.RefreshLayout()
	}
}

//RefreshLayout removes and adds the main parts of the layout
//so that the ones that are disabled by settings do not show up.
func (window *Window) RefreshLayout() {
	window.rootContainer.RemoveItem(window.leftArea)
	window.rootContainer.RemoveItem(window.chatArea)
	window.rootContainer.RemoveItem(window.userContainer)

	window.rootContainer.AddItem(window.leftArea, 0, 7, true)

	conf := config.GetConfig()
	if conf.ShowUserContainer && window.overrideShowUsers {
		window.rootContainer.AddItem(window.chatArea, 0, 20, false)
		window.rootContainer.AddItem(window.userContainer, 0, 6, false)
	} else {
		window.rootContainer.AddItem(window.chatArea, 0, 26, false)
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

		window.AddMessages(messages)
	}

	if channel.Topic != "" {
		window.channelTitle.SetText(channel.Name + " - " + channel.Topic)
	} else {
		window.channelTitle.SetText(channel.Name)
	}

	window.selectedChannel = channel

	return nil
}

func (window *Window) AddMessages(messages []*discordgo.Message) {
	window.SetMessages(append(window.shownMessages, messages...))
}

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

		if window.userContainer.GetCurrentNode() == nil {
			userNodes := window.userRootNode.GetChildren()
			if userNodes != nil && len(userNodes) > 0 {
				window.userContainer.SetCurrentNode(window.userRootNode.GetChildren()[0])
			}
		}
	})
}

//Run Shows the window optionally returning an error.
func (window *Window) Run() error {
	return window.app.Run()
}
