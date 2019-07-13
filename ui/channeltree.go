package ui

import (
	"sort"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

type channelState int

const (
	channelLoaded channelState = iota
	channelUnread
	channelMentioned
	channelRead
)

// ChannelTree is the component that displays the channel hierarchy of the
// currently loaded guild and allows interactions with those channels.
type ChannelTree struct {
	internalTreeView *tview.TreeView

	state *discordgo.State

	onChannelSelect func(channelID string)
	channelStates   map[*tview.TreeNode]channelState
	channelPosition map[string]int
}

// NewChannelTree creates a new ready-to-be-used ChannelTree
func NewChannelTree(state *discordgo.State) *ChannelTree {
	channelTree := &ChannelTree{
		state:            state,
		internalTreeView: tview.NewTreeView(),
		channelStates:    make(map[*tview.TreeNode]channelState),
		channelPosition:  make(map[string]int),
	}

	channelTree.internalTreeView.
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true).
		SetTopLevel(1).
		SetBorder(true)

	channelTree.internalTreeView.SetRoot(tview.NewTreeNode(""))
	channelTree.internalTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		channelID, ok := node.GetReference().(string)
		if ok && channelTree.onChannelSelect != nil {
			channelTree.onChannelSelect(channelID)
		}
	})

	return channelTree
}

// LoadGuild accesses the state in order to load all locally present channels
// for the passed guild.
func (channelTree *ChannelTree) LoadGuild(guildID string) error {
	guild, cacheError := channelTree.state.Guild(guildID)
	if cacheError != nil {
		return cacheError
	}

	channelTree.channelStates = make(map[*tview.TreeNode]channelState)
	channelTree.channelPosition = make(map[string]int)
	channelTree.internalTreeView.GetRoot().ClearChildren()

	channels := guild.Channels
	sort.Slice(channels, func(a, b int) bool {
		return channels[a].Position < channels[b].Position
	})

	// Top level channel
	state := channelTree.state
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildText || channel.ParentID != "" || !hasReadMessagesPermission(channel, state) {
			continue
		}
		createTopLevelChannelNodes(channelTree, channel)
	}
	// Categories
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildCategory || channel.ParentID != "" || !hasReadMessagesPermission(channel, state) {
			continue
		}
		createChannelCategoryNodes(channelTree, channel)
	}
	// Second level channel
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildText || channel.ParentID == "" || !hasReadMessagesPermission(channel, state) {
			continue
		}
		createSecondLevelChannelNodes(channelTree, channel)
	}
	channelTree.internalTreeView.SetCurrentNode(channelTree.internalTreeView.GetRoot())
	return nil
}

func createTopLevelChannelNodes(channelTree *ChannelTree, channel *discordgo.Channel) {
	channelNode := createChannelNode(channel)
	if !readstate.HasBeenRead(channel.ID, channel.LastMessageID) {
		channelTree.channelStates[channelNode] = channelUnread
		channelNode.SetColor(tcell.ColorRed)
	}
	channelTree.internalTreeView.GetRoot().AddChild(channelNode)
}

func createChannelCategoryNodes(channelTree *ChannelTree, channel *discordgo.Channel) {
	channelNode := createChannelNode(channel)
	channelNode.SetSelectable(false)
	channelTree.internalTreeView.GetRoot().AddChild(channelNode)
}

func createSecondLevelChannelNodes(channelTree *ChannelTree, channel *discordgo.Channel) {
	channelNode := createChannelNode(channel)
	for _, node := range channelTree.internalTreeView.GetRoot().GetChildren() {
		channelID, ok := node.GetReference().(string)
		if ok && channelID == channel.ParentID {
			if !readstate.HasBeenRead(channel.ID, channel.LastMessageID) {
				channelTree.channelStates[channelNode] = channelUnread
				channelNode.SetColor(tcell.ColorRed)
			}

			node.AddChild(channelNode)
			break
		}
	}
}

// Checks if the user has permission to view a specific channel.
func hasReadMessagesPermission(channel *discordgo.Channel, state *discordgo.State) bool {
	userPermissions, err := state.UserChannelPermissions(state.User.ID, channel.ID)
	if err != nil {
		// Unable to access channel permissions.
		return false
	}
	return (userPermissions & discordgo.PermissionReadMessages) > 0
}

func createChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := tview.NewTreeNode(discordutil.GetChannelNameForTree(channel))
	channelNode.SetReference(channel.ID)
	return channelNode
}

// AddOrUpdateChannel either adds a new node for the given channel or updates
// its current node.
func (channelTree *ChannelTree) AddOrUpdateChannel(channel *discordgo.Channel) {
	var updated bool
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		nodeChannelID, ok := node.GetReference().(string)
		if ok && nodeChannelID == channel.ID {
			//TODO Do the moving somehow
			/*oldPosition := channelTree.channelPosition[channel.ID]
			oldParentID, parentOk := parent.GetReference().(string)
			if (!parentOk && channel.ParentID != "") || (oldPosition != channel.Position) ||
				(parentOk && channel.ParentID != oldParentID) {

			}*/

			updated = true
			node.SetText(discordutil.GetChannelNameForTree(channel))

			return false
		}

		return true
	})

	if !updated {
		channelNode := createChannelNode(channel)
		if channel.ParentID == "" {
			channelTree.GetRoot().AddChild(channelNode)
		} else {
			for _, node := range channelTree.GetRoot().GetChildren() {
				channelID, ok := node.GetReference().(string)
				if ok && channelID == channel.ParentID {
					node.AddChild(channelNode)
				}
			}
		}
	}
}

// RemoveChannel removes a channels node from the tree.
func (channelTree *ChannelTree) RemoveChannel(channel *discordgo.Channel) {
	channelID := channel.ID

	if channel.Type == discordgo.ChannelTypeGuildText {
		channelTree.internalTreeView.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
			nodeChannelID, ok := node.GetReference().(string)
			if ok && nodeChannelID == channelID {
				channelTree.removeNode(node, parent, channelID)
				return false
			}

			return true
		})
	} else if channel.Type == discordgo.ChannelTypeGuildCategory {
		for _, node := range channelTree.GetRoot().GetChildren() {
			nodeChannelID, ok := node.GetReference().(string)
			if ok && nodeChannelID == channelID {
				oldChildren := node.GetChildren()
				node.SetChildren(make([]*tview.TreeNode, 0))
				channelTree.removeNode(node, channelTree.GetRoot(), channelID)
				channelTree.GetRoot().SetChildren(append(channelTree.GetRoot().GetChildren(), oldChildren...))
				break
			}
		}
	}
}

func (channelTree *ChannelTree) removeNode(node, parent *tview.TreeNode, channelID string) {
	delete(channelTree.channelStates, node)
	delete(channelTree.channelPosition, channelID)
	children := parent.GetChildren()
	if len(children) == 1 {
		parent.SetChildren(make([]*tview.TreeNode, 0))
	} else {
		var childIndex int
		for index, child := range children {
			if child == node {
				childIndex = index
			}
		}

		if childIndex == 0 {
			parent.SetChildren(children[1:])
		} else if childIndex == len(children)-1 {
			parent.SetChildren(children[:len(children)-1])
		} else {
			parent.SetChildren(append(children[:childIndex], children[childIndex+1:]...))
		}
	}
}

// GetRoot returns the root node of the treeview.
func (channelTree *ChannelTree) GetRoot() *tview.TreeNode {
	return channelTree.internalTreeView.GetRoot()
}

// MarkChannelAsUnread marks a channel as unread.
func (channelTree *ChannelTree) MarkChannelAsUnread(channelID string) {
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channelTree.channelStates[node] = channelUnread
			node.SetColor(tcell.ColorRed)

			return false
		}

		return true
	})
}

// MarkChannelAsRead marks a channel as read if it's not loaded already.
func (channelTree *ChannelTree) MarkChannelAsRead(channelID string) {
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channel, stateError := channelTree.state.Channel(channelID)
			if stateError == nil {
				node.SetText(discordutil.GetChannelNameForTree(channel))
			}

			if channelTree.channelStates[node] != channelLoaded {
				channelTree.channelStates[node] = channelRead
				node.SetColor(tcell.ColorWhite)
			}

			return false
		}

		return true
	})
}

// MarkChannelAsMentioned marks a channel as mentioned.
func (channelTree *ChannelTree) MarkChannelAsMentioned(channelID string) {
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channelTree.channelStates[node] = channelMentioned
			channel, stateError := channelTree.state.Channel(channelID)
			if stateError == nil {
				node.SetText("(@You) " + discordutil.GetChannelNameForTree(channel))
			}
			node.SetColor(tcell.ColorRed)

			return false
		}

		return true
	})
}

// MarkChannelAsLoaded marks a channel as loaded and therefore marks all other
// channels as either unread, read or mentioned.
func (channelTree *ChannelTree) MarkChannelAsLoaded(channelID string) {
	for node, state := range channelTree.channelStates {
		if state == channelLoaded {
			channelTree.channelStates[node] = channelRead
			node.SetColor(tcell.ColorWhite)
			break
		}
	}

	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channelTree.channelStates[node] = channelLoaded
			channel, stateError := channelTree.state.Channel(channelID)
			if stateError == nil {
				node.SetText(discordutil.GetChannelNameForTree(channel))
			}
			node.SetColor(tview.Styles.ContrastBackgroundColor)
			return false
		}

		return true
	})
}

// SetOnChannelSelect sets the handler that reacts to channel selection events.
func (channelTree *ChannelTree) SetOnChannelSelect(handler func(channelID string)) {
	channelTree.onChannelSelect = handler
}
