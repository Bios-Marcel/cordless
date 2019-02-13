package ui

import (
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

// PrivateChatList holds the nodes and handlers for the private view. That
// is the one responsible for managing private chats and friends.
type PrivateChatList struct {
	internalTreeView *tview.TreeView

	state *discordgo.State

	chatsNode   *tview.TreeNode
	friendsNode *tview.TreeNode

	onChannelSelect func(channelID string)
	onFriendSelect  func(userID string)

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
	}

	privateList.internalTreeView.
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetCycleSelection(true).
		SetSelectedFunc(privateList.onNodeSelected)

	privateList.chatsNode.SetSelectable(false)
	privateList.friendsNode.SetSelectable(false)

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
			node.SetText(discordgoplus.GetPrivateChannelName(channel))
			return
		}
	}

	privateList.addChannel(channel)
}

func (privateList *PrivateChatList) addChannel(channel *discordgo.Channel) {
	channelNode := tview.NewTreeNode(discordgoplus.GetPrivateChannelName(channel))
	channelNode.SetReference(channel.ID)
	privateList.chatsNode.AddChild(channelNode)
}

// AddOrUpdateFriend either adds a friend or updates the node if it is
// already present
func (privateList *PrivateChatList) AddOrUpdateFriend(user *discordgo.User) {
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

// RemoveFriend removes a friend node if present.
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
func (privateList *PrivateChatList) RemoveChannel(channelID string) {
	newChildren := make([]*tview.TreeNode, 0)

	for _, node := range privateList.chatsNode.GetChildren() {
		referenceChannelID, ok := node.GetReference().(string)
		if !ok || ok && channelID != referenceChannelID {
			newChildren = append(newChildren, node)
		}
	}

	privateList.chatsNode.SetChildren(newChildren)
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
func (privateList *PrivateChatList) Load() error {
	privateChannels := make([]*discordgo.Channel, len(privateList.state.PrivateChannels))
	copy(privateChannels, privateList.state.PrivateChannels)
	discordgoplus.SortPrivateChannels(privateChannels)

	for _, channel := range privateChannels {
		privateList.addChannel(channel)
	}

	for _, friend := range privateList.state.Relationships {
		if friend.Type != discordgoplus.RelationTypeFriend {
			continue
		}

		//TODO Add filter logic or not?

		privateList.addFriend(friend.User)
	}

	return nil
}
