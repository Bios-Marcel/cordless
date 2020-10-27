package ui

import (
	"fmt"
	"sort"

	tcell "github.com/gdamore/tcell/v2"

	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/config"
)

type privateChannelState int

const (
	loaded privateChannelState = iota
	unread
	read
)

// PrivateChatList holds the nodes and handlers for the private view. That
// is the one responsible for managing private chats and friends.
type PrivateChatList struct {
	internalTreeView *tview.TreeView

	state *discordgo.State

	chatsNode   *tview.TreeNode
	friendsNode *tview.TreeNode

	onChannelSelect      func(channelID string)
	onFriendSelect       func(userID string)
	privateChannelStates map[*tview.TreeNode]privateChannelState
}

// NewPrivateChatList creates a new ready to use private chat list.
func NewPrivateChatList(state *discordgo.State) *PrivateChatList {
	privateList := &PrivateChatList{
		state: state,

		internalTreeView: tview.NewTreeView(),
		chatsNode:        tview.NewTreeNode("Chats"),
		friendsNode:      tview.NewTreeNode("Friends"),

		privateChannelStates: make(map[*tview.TreeNode]privateChannelState),
	}

	privateList.internalTreeView.
		SetVimBindingsEnabled(config.Current.OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetCycleSelection(true).
		SetSelectedFunc(privateList.onNodeSelected).
		SetBorder(true).
		SetIndicateOverflow(true).
		SetTitle("DMs").
		SetTitleAlign(tview.AlignLeft)

	privateList.internalTreeView.GetRoot().AddChild(privateList.chatsNode)

	if !state.User.Bot {
		privateList.internalTreeView.GetRoot().AddChild(privateList.friendsNode)
	}

	return privateList
}

func (privateList *PrivateChatList) setNotificationCount(count int) {
	if count == 0 {
		privateList.internalTreeView.SetTitle("DMs")
	} else {
		privateList.internalTreeView.SetTitle(fmt.Sprintf("DMs[%s](%d)", tviewutil.ColorToHex(config.GetTheme().AttentionColor), count))
	}
}

func (privateList *PrivateChatList) onNodeSelected(node *tview.TreeNode) {
	if node.GetParent() == privateList.chatsNode {
		if privateList.onChannelSelect != nil {
			channelID, ok := node.GetReference().(string)
			if ok {
				privateList.onChannelSelect(channelID)
			}
		}
	} else if node.GetParent() == privateList.friendsNode {
		if privateList.onFriendSelect != nil {
			userID, ok := node.GetReference().(string)
			if ok {
				privateList.onFriendSelect(userID)
			}
		}
	}
}

// AddOrUpdateChannel either adds a channel or updates the node if it is
// already present
func (privateList *PrivateChatList) AddOrUpdateChannel(channel *discordgo.Channel) {
	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channel.ID {
			node.SetText(discordutil.GetPrivateChannelName(channel))
			return
		}
	}

	if channel.Type == discordgo.ChannelTypeDM {
		user := channel.Recipients[0]
		for _, friendNode := range privateList.friendsNode.GetChildren() {
			userID, ok := friendNode.GetReference().(string)
			if ok && userID == user.ID {
				privateList.RemoveFriend(userID)
				break
			}
		}
	}

	privateList.prependChannel(channel)
}

func (privateList *PrivateChatList) prependChannel(channel *discordgo.Channel) {
	newChildren := append([]*tview.TreeNode{createPrivateChannelNode(channel)}, privateList.chatsNode.GetChildren()...)
	privateList.chatsNode.SetChildren(newChildren)
}

func (privateList *PrivateChatList) addChannel(channel *discordgo.Channel) {
	newNode := createPrivateChannelNode(channel)
	if !readstate.HasBeenRead(channel, channel.LastMessageID) {
		privateList.privateChannelStates[newNode] = unread
		if tview.IsVtxxx {
			newNode.SetBlinking(true)
		} else {
			newNode.SetColor(config.GetTheme().AttentionColor)
		}
	}
	privateList.chatsNode.AddChild(newNode)
}

func createPrivateChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := tview.NewTreeNode(discordutil.GetPrivateChannelName(channel))
	channelNode.SetReference(channel.ID)
	return channelNode
}

// AddOrUpdateFriend either adds a friend or updates the node if it is
// already present.
func (privateList *PrivateChatList) AddOrUpdateFriend(user *discordgo.User) {
	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok {
			channel, stateError := privateList.state.Channel(referenceChannelID)
			if stateError == nil && channel.Type == discordgo.ChannelTypeDM {
				if channel.Recipients[0].ID == user.ID {
					node.SetText(discordutil.GetUserName(user))
					return
				}
			}
		}
	}

	for _, node := range privateList.friendsNode.GetChildren() {
		referenceUserID, ok := node.GetReference().(string)
		if ok && referenceUserID == user.ID {
			node.SetText(discordutil.GetUserName(user))
			return
		}
	}

	privateList.addFriend(user)
}

func (privateList *PrivateChatList) addFriend(user *discordgo.User) {
	friendNode := tview.NewTreeNode(user.Username)
	friendNode.SetReference(user.ID)
	privateList.friendsNode.AddChild(friendNode)
}

// RemoveFriend removes a friend node if present. This will not trigger any
// action on the channel list.
func (privateList *PrivateChatList) RemoveFriend(userID string) {
	newChildren := make([]*tview.TreeNode, 0)

	for _, node := range privateList.friendsNode.GetChildren() {
		referenceUserID, ok := node.GetReference().(string)
		if !ok || ok && userID != referenceUserID {
			newChildren = append(newChildren, node)
		}
	}

	privateList.friendsNode.SetChildren(newChildren)
}

// RemoveChannel removes a channel node if present.
func (privateList *PrivateChatList) RemoveChannel(channel *discordgo.Channel) {
	newChildren := make([]*tview.TreeNode, 0)
	channelID := channel.ID

	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if !ok || ok && channelID != referenceChannelID {
			newChildren = append(newChildren, node)
		} else {
			delete(privateList.privateChannelStates, node)
		}
	}

	userID := channel.Recipients[0].ID

	for _, relationship := range privateList.state.Relationships {
		if relationship.Type == discordgo.RelationTypeFriend &&
			relationship.User.ID == userID {
			privateList.AddOrUpdateFriend(relationship.User)
			break
		}
	}

	privateList.chatsNode.SetChildren(newChildren)
}

// MarkAsUnread marks the channel as unread, coloring it red.
func (privateList *PrivateChatList) MarkAsUnread(channelID string) {
	node := tviewutil.GetNodeByReference(channelID, privateList.internalTreeView)
	if node != nil {
		privateList.privateChannelStates[node] = unread
		privateList.setNotificationCount(privateList.amountOfUnreadChannels())
		if tview.IsVtxxx {
			node.SetBlinking(true)
		} else {
			node.SetColor(config.GetTheme().AttentionColor)
		}
	}
}

func (privateList *PrivateChatList) amountOfUnreadChannels() int {
	var amount int
	for _, node := range privateList.chatsNode.GetChildren() {
		if privateList.privateChannelStates[node] == unread {
			amount++
		}
	}

	return amount
}

// MarkAsRead marks a channel as read.
func (privateList *PrivateChatList) MarkAsRead(channelID string) {
	node := tviewutil.GetNodeByReference(channelID, privateList.internalTreeView)
	if node != nil {
		privateList.setNotificationCount(privateList.amountOfUnreadChannels())
		privateList.privateChannelStates[node] = read
		if tview.IsVtxxx {
			node.SetBlinking(false)
			node.SetUnderline(false)
		} else {
			node.SetColor(config.GetTheme().PrimaryTextColor)
		}
	}
}

// Reorder resorts the list of private chats according to their last
// message times.
func (privateList *PrivateChatList) Reorder() {
	children := privateList.chatsNode.GetChildren()
	sort.Slice(children, func(a, b int) bool {
		nodeA := children[a]
		nodeB := children[b]

		referenceA, ok := nodeA.GetReference().(string)
		if !ok {
			return false
		}
		referenceB, ok := nodeB.GetReference().(string)
		if !ok {
			return true
		}

		channelA, stateError := privateList.state.Channel(referenceA)
		if stateError != nil {
			return false
		}
		channelB, stateError := privateList.state.Channel(referenceB)
		if stateError != nil {
			return true
		}

		return discordutil.CompareChannels(channelA, channelB)
	})
}

// MarkAsLoaded marks a channel as loaded, coloring it blue. If
// a different channel had loaded before, it's set to read.
func (privateList *PrivateChatList) MarkAsLoaded(channelID string) {
	for node, state := range privateList.privateChannelStates {
		if state == loaded {
			privateList.privateChannelStates[node] = read
			if tview.IsVtxxx {
				node.SetBlinking(false)
				node.SetUnderline(false)
			} else {
				node.SetColor(config.GetTheme().PrimaryTextColor)
			}
			break
		}
	}

	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			privateList.privateChannelStates[node] = loaded
			if tview.IsVtxxx {
				node.SetUnderline(true)
			} else {
				node.SetColor(tview.Styles.ContrastBackgroundColor)
			}
			break
		}
	}

	privateList.setNotificationCount(privateList.amountOfUnreadChannels())
}

// SetOnFriendSelect sets the handler that decides what happens when a friend
// node gets selected.
func (privateList *PrivateChatList) SetOnFriendSelect(handler func(userID string)) {
	privateList.onFriendSelect = handler
}

// SetOnChannelSelect sets the handler that decides what happens when a
// channel node gets selected.
func (privateList *PrivateChatList) SetOnChannelSelect(handler func(channelID string)) {
	privateList.onChannelSelect = handler
}

// Load loads all present data (chats, groups and friends).
func (privateList *PrivateChatList) Load() {
	privateChannels := make([]*discordgo.Channel, len(privateList.state.PrivateChannels))
	copy(privateChannels, privateList.state.PrivateChannels)
	discordutil.SortPrivateChannels(privateChannels)

	for _, channel := range privateChannels {
		privateList.addChannel(channel)
	}

FRIEND_LOOP:
	for _, friend := range privateList.state.Relationships {
		if friend.Type != discordgo.RelationTypeFriend {
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

		privateList.addFriend(friend.User)
	}

	privateList.internalTreeView.SetCurrentNode(privateList.chatsNode)
	privateList.setNotificationCount(privateList.amountOfUnreadChannels())
}

// GetComponent returns the TreeView component that is used.
// This component is the top-level container of this struct.
func (privateList *PrivateChatList) GetComponent() *tview.TreeView {
	return privateList.internalTreeView
}

//SetInputCapture delegates to tviews SetInputCapture
func (privateList *PrivateChatList) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) {
	privateList.internalTreeView.SetInputCapture(capture)
}
