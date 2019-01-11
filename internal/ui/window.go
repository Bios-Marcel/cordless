package ui

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Window struct {
	app              *tview.Application
	messageContainer *tview.Table
	userContainer    *tview.List
	messageInput     *tview.InputField

	session *discordgo.Session

	shownMessages   []*discordgo.Message
	selectedServer  *discordgo.UserGuild
	selectedChannel *discordgo.Channel
}

func NewWindow(discord *discordgo.Session) (*Window, error) {
	window := Window{
		session: discord,
	}

	guilds, discordError := discord.UserGuilds(100, "", "")
	if discordError != nil {
		return nil, discordError
	}

	app := tview.NewApplication()

	left := tview.NewPages()

	serversPageName := "Servers"
	serversPage := tview.NewFlex()
	serversPage.SetDirection(tview.FlexRow)

	channelsPlaceholder := tview.NewList()
	channelsPlaceholder.SetBorder(true)
	channelsPlaceholder.ShowSecondaryText(false)

	channelsPlaceholder.SetSelectedFunc(func(index int, primary, secondary string, shortcut rune) {
		window.ClearMessages()

		channels, _ := discord.GuildChannels(window.selectedServer.ID)
		for _, channel := range channels {
			if channel.Name == primary {
				window.selectedChannel = channel
				break
			}
		}

		if window.selectedChannel != nil {
			discordError := window.LoadChannel(window.selectedChannel)
			if discordError != nil {
				log.Fatalf("Error loading messages for channel (%s).", discordError.Error())
			}
		}
	})

	serversPlaceholder := tview.NewList()
	serversPlaceholder.SetBorder(true)
	serversPlaceholder.ShowSecondaryText(false)
	for _, guild := range guilds {
		serversPlaceholder.AddItem(guild.Name, "", 0, nil)
	}

	serversPlaceholder.SetSelectedFunc(func(index int, primary, secondary string, shortcut rune) {
		for _, guild := range guilds {
			if guild.Name == primary {
				window.selectedServer = guild
				channelsPlaceholder.Clear()
				//TODO Handle error
				channels, _ := discord.GuildChannels(guild.ID)
				for _, channel := range channels {
					//TODO Filter by permissions
					channelsPlaceholder.AddItem(channel.Name, "", 0, nil)
				}

				//TODO Handle error
				window.userContainer.Clear()
				users, _ := discord.GuildMembers(guild.ID, "", 1000)
				for _, user := range users {
					if user.Nick != "" {
						window.userContainer.AddItem(user.Nick, "", 0, nil)
					} else {
						window.userContainer.AddItem(user.User.Username, "", 0, nil)
					}
				}
				break
			}
		}
	})

	serversPage.AddItem(serversPlaceholder, 0, 1, true)
	serversPage.AddItem(channelsPlaceholder, 0, 2, true)

	left.AddPage(serversPageName, serversPage, true, true)

	friendsPageName := "Friends"
	friendsPage := tview.NewFlex()
	friendsPage.SetDirection(tview.FlexRow)
	left.AddPage(friendsPageName, friendsPage, true, false)

	mid := tview.NewFlex()
	mid.SetDirection(tview.FlexRow)

	messageContainer := tview.NewTable()
	window.messageContainer = messageContainer
	messageContainer.SetBorder(true)
	messageContainer.SetSelectable(true, false)

	messageTick := time.NewTicker(250 * time.Millisecond)
	quitMessageListener := make(chan struct{})
	go func() {
		for {
			select {
			case <-messageTick.C:
				if window.selectedChannel != nil {
					messageAmount := len(window.shownMessages)
					var messages []*discordgo.Message
					var discordError error
					if window.shownMessages != nil && messageAmount > 0 {
						messages, discordError = discord.ChannelMessages(window.selectedChannel.ID, 100, "", window.shownMessages[messageAmount-1].ID, "")
					} else {
						messages, discordError = discord.ChannelMessages(window.selectedChannel.ID, 100, "", "", "")
					}

					//TODO Handle properly
					if discordError != nil {
						continue
					}

					if messages == nil || len(messages) == 0 {
						continue
					}

					window.AddMessages(messages)
				}
			case <-quitMessageListener:
				messageTick.Stop()
				return
			}
		}
	}()

	window.messageInput = tview.NewInputField()
	window.messageInput.SetBorder(true)
	window.messageInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			if window.selectedChannel != nil {
				discord.ChannelMessageSend(window.selectedChannel.ID, window.messageInput.GetText())
				window.messageInput.SetText("")
			}

			return nil
		}

		return event
	})

	mid.AddItem(messageContainer, 0, 1, true)
	mid.AddItem(window.messageInput, 3, 0, true)

	window.userContainer = tview.NewList()
	window.userContainer.ShowSecondaryText(false)
	window.userContainer.SetBorder(true)

	root := tview.NewFlex()
	root.SetDirection(tview.FlexColumn)
	root.SetBorderPadding(-1, -1, 0, 0)

	root.AddItem(left, 0, 7, true)
	root.AddItem(mid, 0, 20, false)
	root.AddItem(window.userContainer, 0, 6, false)

	frame := tview.NewFrame(root)
	frame.SetBorder(true)
	frame.SetTitleAlign(tview.AlignCenter)
	frame.SetTitle("Cordless")

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'C' &&
			event.Modifiers() == tcell.ModAlt {
			app.SetFocus(channelsPlaceholder)
			return nil
		}

		if event.Rune() == 'S' &&
			event.Modifiers() == tcell.ModAlt {
			app.SetFocus(serversPlaceholder)
			return nil
		}

		if event.Rune() == 'T' &&
			event.Modifiers() == tcell.ModAlt {
			app.SetFocus(window.messageContainer)
			return nil
		}

		if event.Rune() == 'U' &&
			event.Modifiers() == tcell.ModAlt {
			app.SetFocus(window.userContainer)
			return nil
		}

		if event.Rune() == 'M' &&
			event.Modifiers() == tcell.ModAlt {
			app.SetFocus(window.messageInput)
			return nil
		}

		return event
	})

	app.SetRoot(frame, true)

	window.app = app

	return &window, nil
}

func (window *Window) ClearMessages() {
	window.messageContainer.Clear()
}

func (window *Window) LoadChannel(channel *discordgo.Channel) error {

	messages, discordError := window.session.ChannelMessages(channel.ID, 100, channel.LastMessageID, "", "")
	if discordError != nil {
		return discordError
	}

	sort.Slice(messages, func(x, y int) bool {
		timeOne, parseError := messages[x].Timestamp.Parse()
		if parseError != nil {
			fmt.Println("Error 1")
			return false
		}

		timeTwo, parseError := messages[y].Timestamp.Parse()
		if parseError != nil {
			fmt.Println("Error 2")
			return false
		}

		return timeOne.Before(timeTwo)
	})

	window.AddMessages(messages)
	return nil
}

func (window *Window) AddMessages(messages []*discordgo.Message) {
	window.shownMessages = append(window.shownMessages, messages...)

	window.app.QueueUpdateDraw(func() {
		for index, message := range messages {
			time, parseError := message.Timestamp.Parse()
			if parseError == nil {
				timeCellText := fmt.Sprintf("%02d:%02d:%02d", time.Hour(), time.Minute(), time.Second())
				window.messageContainer.SetCell(index, 0, tview.NewTableCell(timeCellText))
			}

			//TODO use nickname instead.
			window.messageContainer.SetCell(index, 1, tview.NewTableCell(message.Author.Username))
			window.messageContainer.SetCell(index, 2, tview.NewTableCell(message.Content))
		}

		window.messageContainer.Select(window.messageContainer.GetRowCount()-1, 0)
		window.messageContainer.ScrollToEnd()
	})
}

//Run Shows the window optionally returning an error.
func (window *Window) Run() error {
	return window.app.Run()
}
