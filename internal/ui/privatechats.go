package ui

import (
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

type nodeState int

const (
	loaded nodeState = iota
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

	onChannelSelect func(node *tview.TreeNode, channelID string)
	onFriendSelect  func(userID string)
	nodeStates      map[*tview.TreeNode]nodeState

	userChannels map[string]*tview.TreeNode
	friends      map[string]*tview.TreeNode
}

// NewPrivateChatList creates a new ready to use private chat list.
func NewPrivateChatList(state *discordgo.State) *PrivateChatList {
	privateList := &PrivateChatList{
		state: state,

		internalTreeView: tview.NewTreeView(),
		chatsNode:        tview.NewTreeNode("Chats"),
		friendsNode:      tview.NewTreeNode("Friends"),

		nodeStates: make(map[*tview.TreeNode]nodeState, 0),
	}

	privateList.internalTreeView.
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetCycleSelection(true).
		SetSelectedFunc(privateList.onNodeSelected).
		SetBorder(true)

	privateList.internalTreeView.GetRoot().
		AddChild(privateList.chatsNode).
		AddChild(privateList.friendsNode)

	return privateList
}

func (privateList *PrivateChatList) onNodeSelected(node *tview.TreeNode) {
	if node.GetParent() == privateList.chatsNode {
		if privateList.onChannelSelect != nil {
			channelID, ok := node.GetReference().(string)
			if ok {
				privateList.onChannelSelect(node, channelID)
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
			node.SetText(discordgoplus.GetPrivateChannelName(channel))
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
	newChildren := append([]*tview.TreeNode{createChannelNode(channel)}, privateList.chatsNode.GetChildren()...)
	privateList.chatsNode.SetChildren(newChildren)
}

func (privateList *PrivateChatList) addChannel(channel *discordgo.Channel) {
	privateList.chatsNode.AddChild(createChannelNode(channel))
}

func createChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := tview.NewTreeNode(discordgoplus.GetPrivateChannelName(channel))
	channelNode.SetReference(channel.ID)
	return channelNode
}

// AddOrUpdateFriend either adds a friend or updates the node if it is
// already present.
func (privateList *PrivateChatList) AddOrUpdateFriend(user *discordgo.User) {
	for _, node := range privateList.chatsNode.GetChildren() {
		refrenceChannelID, ok := node.GetReference().(string)
		if ok {
			channel, stateError := privateList.state.Channel(refrenceChannelID)
			if stateError == nil && channel.Type == discordgo.ChannelTypeDM {
				if channel.Recipients[0].ID == user.ID {
					node.SetText(discordgoplus.GetUserName(user, &userColor))
					return
				}
			}
		}
	}

	for _, node := range privateList.friendsNode.GetChildren() {
		referenceUserID, ok := node.GetReference().(string)
		if ok && referenceUserID == user.ID {
			node.SetText(discordgoplus.GetUserName(user, &userColor))
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
			delete(privateList.nodeStates, node)
		}
	}

	userID := channel.Recipients[0].ID

	for _, relationship := range privateList.state.Relationships {
		if relationship.Type == discordgoplus.RelationTypeFriend &&
			relationship.User.ID == userID {
			privateList.AddOrUpdateFriend(relationship.User)
			break
		}
	}

	privateList.chatsNode.SetChildren(newChildren)
}

// MarkChannelAsUnread marks the channel as unread, coloring it red.
func (privateList *PrivateChatList) MarkChannelAsUnread(channel *discordgo.Channel) {
	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channel.ID {
			privateList.nodeStates[node] = unread
			node.SetColor(tcell.ColorRed)
			break
		}
	}
}

// MarkChannelAsRead marks a channel as read, coloring it white.
func (privateList *PrivateChatList) MarkChannelAsRead(channel *discordgo.Channel) {
	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channel.ID {
			privateList.nodeStates[node] = read
			node.SetColor(tcell.ColorWhite)
			break
		}
	}
}

// MarkChannelAsLoaded marks a channel as loaded, coloring it blue (teal). If
// a different channel had loaded before, it's set to read.
func (privateList *PrivateChatList) MarkChannelAsLoaded(channel *discordgo.Channel) {
	for node, state := range privateList.nodeStates {
		if state == loaded {
			privateList.nodeStates[node] = read
			node.SetColor(tcell.ColorWhite)
			break
		}
	}

	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channel.ID {
			privateList.nodeStates[node] = loaded
			node.SetColor(tcell.ColorTeal)
			break
		}
	}
}

// SetOnFriendSelect sets the handler that decides what happens when a friend
// node gets selected.
func (privateList *PrivateChatList) SetOnFriendSelect(handler func(userID string)) {
	privateList.onFriendSelect = handler
}

// SetOnChannelSelect sets the handler that decides what happens when a
// channel node gets selected.
func (privateList *PrivateChatList) SetOnChannelSelect(handler func(node *tview.TreeNode, channelID string)) {
	privateList.onChannelSelect = handler
}

// Load loads all present data (chats, groups and friends).
func (privateList *PrivateChatList) Load() error {
	privateChannels := make([]*discordgo.Channel, len(privateList.state.PrivateChannels))
	copy(privateChannels, privateList.state.PrivateChannels)
	discordgoplus.SortPrivateChannels(privateChannels)

	for _, channel := range privateChannels {
		privateList.addChannel(channel)
	}

FRIEND_LOOP:
	for _, friend := range privateList.state.Relationships {
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

		privateList.addFriend(friend.User)
	}

	privateList.internalTreeView.SetCurrentNode(privateList.chatsNode)

	return nil
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
