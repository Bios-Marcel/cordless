package ui

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const (
	updateInterval         = 250 * time.Millisecond
	userListUpdateInterval = 5 * time.Second
)

type Window struct {
	app *tview.Application

	rootContainer *tview.Flex

	leftArea      *tview.Pages
	chatArea      *tview.Flex
	userContainer *tview.TreeView

	messageContainer *tview.Table
	userRootNode     *tview.TreeNode
	messageInput     *tview.InputField
	channelRootNode  *tview.TreeNode
	channelTitle     *tview.TextView

	killCurrentGuildUpdateThread   *chan bool
	killCurrentChannelUpdateThread *chan bool
	session                        *discordgo.Session

	shownMessages   []*discordgo.Message
	selectedGuild   *discordgo.UserGuild
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

	guildPageName := "Guilds"
	guildPage := tview.NewFlex()
	guildPage.SetDirection(tview.FlexRow)

	channelTree := tview.NewTreeView()
	channelRootNode := tview.NewTreeNode("")
	window.channelRootNode = channelRootNode
	channelTree.SetRoot(channelRootNode)
	channelTree.SetBorder(true)
	channelTree.SetTopLevel(1)

	guildList := tview.NewList()
	guildList.SetBorder(true)
	guildList.ShowSecondaryText(false)
	for _, guild := range guilds {
		guildList.AddItem(guild.Name, "", 0, nil)
	}

	guildList.SetSelectedFunc(func(index int, primary, secondary string, shortcut rune) {
		for _, guild := range guilds {
			if guild.Name == primary {
				if window.killCurrentGuildUpdateThread != nil {
					*window.killCurrentGuildUpdateThread <- true
				}

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

						//No selection will prevent selection from working at all.
						if channelTree.GetCurrentNode() == nil {
							channelTree.SetCurrentNode(newNode)
						}

						newNode.SetSelectable(true)
						//This copy is necessary in order to use the correct channel instead
						//of always the same one.
						channelToConnectTo := channel
						newNode.SetSelectedFunc(func() {
							window.selectedChannel = channelToConnectTo

							window.ClearMessages()
							discordError := window.LoadChannel(channelToConnectTo)
							if discordError != nil {
								log.Fatalf("Error loading messages (%s).", discordError.Error())
							}
						})

						channelCategories[channelToConnectTo.ParentID].AddChild(newNode)
					}
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
				break
			}
		}
	})

	guildPage.AddItem(guildList, 0, 1, true)
	guildPage.AddItem(channelTree, 0, 2, true)

	window.leftArea.AddPage(guildPageName, guildPage, true, true)

	/*friendsPageName := "Friends"
	friendsPage := tview.NewFlex()
	friendsPage.SetDirection(tview.FlexRow)
	left.AddPage(friendsPageName, friendsPage, true, false)*/

	window.chatArea = tview.NewFlex()
	window.chatArea.SetDirection(tview.FlexRow)

	messageContainer := tview.NewTable()
	window.messageContainer = messageContainer
	messageContainer.SetBorder(true)
	messageContainer.SetSelectable(true, false)

	window.messageInput = tview.NewInputField()
	window.messageInput.SetBorder(true)
	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if window.selectedChannel != nil {
				messageToSend := window.messageInput.GetText()
				window.messageInput.SetText("")

				guild, discordError := window.session.State.Guild(window.selectedGuild.ID)
				if discordError == nil {

					//Those could be optimized by searching the string for patterns.

					for _, channel := range guild.Channels {
						if channel.Type == discordgo.ChannelTypeGuildText {
							messageToSend = strings.Replace(messageToSend, "#"+channel.Name, "<#"+channel.ID+">", -1)
						}
					}

					for _, member := range guild.Members {
						messageToSend = strings.Replace(messageToSend, "@"+member.User.Username, "<@"+member.User.ID+">", -1)
					}
				}

				go discord.ChannelMessageSend(window.selectedChannel.ID, messageToSend)
			}

			return nil
		}

		return event
	})

	window.channelTitle = tview.NewTextView()
	window.channelTitle.SetBorder(true)

	window.chatArea.AddItem(window.channelTitle, 3, 1, true)
	window.chatArea.AddItem(messageContainer, 0, 1, true)
	window.chatArea.AddItem(window.messageInput, 3, 0, true)

	window.userContainer = tview.NewTreeView()
	window.userRootNode = tview.NewTreeNode("")
	window.userContainer.SetTopLevel(1)
	window.userContainer.SetRoot(window.userRootNode)
	window.userContainer.SetBorder(true)

	window.rootContainer = tview.NewFlex()
	window.rootContainer.SetDirection(tview.FlexColumn)
	window.rootContainer.SetBorderPadding(-1, -1, 0, 0)

	window.RefreshLayout()

	if config.GetConfig().ShowFrame {
		frame := tview.NewFrame(window.rootContainer)
		frame.SetBorder(true)
		frame.SetTitleAlign(tview.AlignCenter)
		frame.SetTitle("Cordless")
		app.SetRoot(frame, true)
	} else {
		app.SetRoot(window.rootContainer, true)
	}

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
			if event.Rune() == 'c' {
				app.SetFocus(channelTree)
				return nil
			}

			if event.Rune() == 's' {
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

	return &window, nil
}

func (window *Window) RefreshLayout() {
	window.rootContainer.RemoveItem(window.leftArea)
	window.rootContainer.RemoveItem(window.chatArea)
	window.rootContainer.RemoveItem(window.userContainer)

	window.rootContainer.AddItem(window.leftArea, 0, 7, true)
	if config.GetConfig().ShowUserContainer {
		window.rootContainer.AddItem(window.chatArea, 0, 20, false)
		window.rootContainer.AddItem(window.userContainer, 0, 6, false)
	} else {
		window.rootContainer.AddItem(window.chatArea, 0, 26, false)
	}

	window.app.ForceDraw()
}

func (window *Window) ClearMessages() {
	window.messageContainer.Clear()
}

func (window *Window) LoadChannel(channel *discordgo.Channel) error {
	if window.killCurrentChannelUpdateThread != nil {
		*window.killCurrentChannelUpdateThread <- true
	}

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

	updateTicker := time.NewTicker(updateInterval)
	go func() {
		killChan := make(chan bool)
		window.killCurrentChannelUpdateThread = &killChan
		for {
			select {
			case <-*window.killCurrentChannelUpdateThread:
				window.killCurrentChannelUpdateThread = nil
				return

			case <-updateTicker.C:
				window.LoadMessagesInChannelAfter(channel)
			}
		}
	}()

	return nil
}

func (window *Window) LoadMessagesInChannelAfter(channel *discordgo.Channel) {
	lastMessageID := window.shownMessages[len(window.shownMessages)-1].ID
	messages, discordError := window.session.ChannelMessages(channel.ID, 100, "", lastMessageID, "")

	//TODO Handle
	if discordError != nil {
		return
	}

	if messages == nil || len(messages) == 0 {
		return
	}

	window.AddMessages(messages)
}

func (window *Window) AddMessages(messages []*discordgo.Message) {
	window.shownMessages = append(window.shownMessages, messages...)

	window.app.QueueUpdateDraw(func() {
		for _, message := range messages {

			rowIndex := window.messageContainer.GetRowCount()

			time, parseError := message.Timestamp.Parse()
			if parseError == nil {
				time := time.Local()
				var timeCellText string
				conf := config.GetConfig()
				if conf.Times == config.HourMinuteAndSeconds {
					timeCellText = fmt.Sprintf("%02d:%02d:%02d", time.Hour(), time.Minute(), time.Second())
					window.messageContainer.SetCell(rowIndex, 0, tview.NewTableCell(timeCellText))
				} else if conf.Times == config.HourAndMinute {
					timeCellText = fmt.Sprintf("%02d:%02d", time.Hour(), time.Minute())
					window.messageContainer.SetCell(rowIndex, 0, tview.NewTableCell(timeCellText))
				}
			}

			//TODO use nickname instead.
			window.messageContainer.SetCell(rowIndex, 1, tview.NewTableCell(message.Author.Username))
			messageText := message.ContentWithMentionsReplaced()
			if message.Attachments != nil && len(message.Attachments) != 0 {
				if messageText != "" {
					messageText = messageText + " "
				}
				messageText = messageText + message.Attachments[0].URL
			}

			window.messageContainer.SetCell(rowIndex, 2, tview.NewTableCell(messageText))
		}

		window.messageContainer.Select(window.messageContainer.GetRowCount()-1, 0)
		window.messageContainer.ScrollToEnd()
	})
}

func (window *Window) UpdateUsersForGuild(guild *discordgo.UserGuild) {
	users, discordError := window.session.GuildMembers(guild.ID, "", 1000)

	//TODO Handle error
	if discordError != nil {
		return
	}

	window.app.QueueUpdateDraw(func() {
		window.userRootNode.ClearChildren()

		roles, _ := window.session.GuildRoles(guild.ID)
		roleNodes := make(map[string]*tview.TreeNode)

		sort.Slice(roles, func(a, b int) bool {
			return roles[a].Position > roles[b].Position
		})

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
				for _, role := range roles {
					if role.ID == firstIdentifier {
						firstRole = role
					}
				}

				var secondRole *discordgo.Role
				for _, role := range roles {
					if role.ID == secondIdentifier {
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
