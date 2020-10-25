package ui

import (
	"sort"
	"sync"

	"github.com/Bios-Marcel/cordless/ui/tviewutil"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/readstate"
)

type channelState int

const (
	channelLoaded channelState = iota
	channelUnread
	channelMentioned
	channelRead
)

var (
	mentionedIndicator = "(@)"
	nsfwIndicator      = tviewutil.Escape("🔞")
	lockedIndicator    = tviewutil.Escape("\U0001F512")
)

// ChannelTree is the component that displays the channel hierarchy of the
// currently loaded guild and allows interactions with those channels.
type ChannelTree struct {
	*tview.TreeView
	*sync.Mutex

	state *discordgo.State

	onChannelSelect func(channelID string)
	channelStates   map[*tview.TreeNode]channelState
	channelPosition map[string]int
	prefixes        map[string][]string
}

// NewChannelTree creates a new ready-to-be-used ChannelTree
func NewChannelTree(state *discordgo.State) *ChannelTree {
	channelTree := &ChannelTree{
		state:           state,
		TreeView:        tview.NewTreeView(),
		channelStates:   make(map[*tview.TreeNode]channelState),
		channelPosition: make(map[string]int),
		prefixes:        make(map[string][]string),
		Mutex:           &sync.Mutex{},
	}

	channelTree.
		SetVimBindingsEnabled(config.Current.OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true).
		SetTopLevel(1).
		SetBorder(true).
		SetIndicateOverflow(true)

	channelTree.SetRoot(tview.NewTreeNode(""))
	channelTree.SetSelectedFunc(func(node *tview.TreeNode) {
		channelID, ok := node.GetReference().(string)
		if ok && channelTree.onChannelSelect != nil {
			channelTree.onChannelSelect(channelID)
		}
	})

	return channelTree
}

// Clear resets all current state.
func (channelTree *ChannelTree) Clear() {
	channelTree.channelStates = make(map[*tview.TreeNode]channelState)
	channelTree.channelPosition = make(map[string]int)
	channelTree.GetRoot().ClearChildren()
}

// LoadGuild accesses the state in order to load all locally present channels
// for the passed guild.
func (channelTree *ChannelTree) LoadGuild(guildID string) error {
	state := channelTree.state
	guild, cacheError := state.Guild(guildID)
	if cacheError != nil {
		return cacheError
	}

	channelTree.Clear()

	channels := guild.Channels
	sort.Slice(channels, func(a, b int) bool {
		return channels[a].Position < channels[b].Position
	})

	// Top level channel
	for _, channel := range channels {
		if (channel.Type != discordgo.ChannelTypeGuildText && channel.Type != discordgo.ChannelTypeGuildNews) ||
			channel.ParentID != "" || !discordutil.HasReadMessagesPermission(channel.ID, state) {
			continue
		}
		channelTree.createTopLevelChannelNodes(channel)
	}
	// Categories; Must be handled before second level channels, as the
	// categories serve as parents.
CATEGORY_LOOP:
	for _, channel := range channels {
		if channel.Type != discordgo.ChannelTypeGuildCategory || channel.ParentID != "" {
			continue
		}

		childless := true
		for _, potentialChild := range channels {
			if potentialChild.ParentID == channel.ID {
				if discordutil.HasReadMessagesPermission(potentialChild.ID, state) {
					//We have at least one child with read-permissions,
					//therefore we add the category as the channel will need
					//a parent.
					channelTree.createChannelCategoryNode(channel)
					continue CATEGORY_LOOP
				}

				//Has at least once child-channel, so we don't need to add a
				//category later on, if none of the child-channels is
				//accessible by the currently logged on user.
				childless = false
			}
		}

		//If the category is childless, we want to add it anyway.
		if childless {
			channelTree.createChannelCategoryNode(channel)
		}
	}
	// Second level channel
	for _, channel := range channels {
		//Only Text and News are supported. If new channel types are
		//added, support first needs to be confirmed or implemented. This is
		//in order to avoid faulty runtime behaviour.
		if (channel.Type != discordgo.ChannelTypeGuildText && channel.Type != discordgo.ChannelTypeGuildNews) ||
			channel.ParentID == "" || !discordutil.HasReadMessagesPermission(channel.ID, state) {
			continue
		}
		channelTree.createSecondLevelChannelNodes(channel)
	}
	channelTree.SetCurrentNode(channelTree.GetRoot())
	return nil
}

func (channelTree *ChannelTree) createTopLevelChannelNodes(channel *discordgo.Channel) {
	channelNode := channelTree.createTextChannelNode(channel)
	channelTree.GetRoot().AddChild(channelNode)
}

func (channelTree *ChannelTree) createChannelCategoryNode(channel *discordgo.Channel) {
	channelNode := channelTree.createChannelNode(channel)
	channelTree.GetRoot().AddChild(channelNode)
}

func (channelTree *ChannelTree) createSecondLevelChannelNodes(channel *discordgo.Channel) {
	parentNode := tviewutil.GetNodeByReference(channel.ParentID, channelTree.TreeView)
	if parentNode != nil {
		channelNode := channelTree.createTextChannelNode(channel)
		parentNode.AddChild(channelNode)
	}
}

func (channelTree *ChannelTree) createChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := tview.NewTreeNode(tviewutil.Escape(channel.Name))
	if channel.NSFW {
		channelNode.AddPrefix(nsfwIndicator)
	}

	// Adds a padlock prefix if the channel if not readable by the everyone group
	if config.Current.IndicateChannelAccessRestriction {
		for _, permission := range channel.PermissionOverwrites {
			if permission.Type == "role" && permission.ID == channel.GuildID && permission.Deny&discordgo.PermissionViewChannel == discordgo.PermissionViewChannel {
				channelNode.AddPrefix(lockedIndicator)
				break
			}
		}
	}

	channelNode.SetReference(channel.ID)
	return channelNode
}

func (channelTree *ChannelTree) createTextChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := channelTree.createChannelNode(channel)

	if !readstate.HasBeenRead(channel, channel.LastMessageID) {
		channelTree.channelStates[channelNode] = channelUnread
		channelTree.markNodeAsUnread(channelNode)
	}

	if readstate.HasBeenMentioned(channel.ID) {
		channelTree.markNodeAsMentioned(channelNode, channel.ID)
	}

	return channelNode
}

// AddOrUpdateChannel either adds a new node for the given channel or updates
// its current node.
func (channelTree *ChannelTree) AddOrUpdateChannel(channel *discordgo.Channel) {
	var updated bool
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		nodeChannelID, ok := node.GetReference().(string)
		if ok && nodeChannelID == channel.ID {
			//TODO Support Re-Parenting
			/*oldPosition := channelTree.channelPosition[channel.ID]
			oldParentID, parentOk := parent.GetReference().(string)
			if (!parentOk && channel.ParentID != "") || (oldPosition != channel.Position) ||
				(parentOk && channel.ParentID != oldParentID) {

			}*/

			updated = true
			node.SetText(tviewutil.Escape(channel.Name))

			return false
		}

		return true
	})

	if !updated {
		channelNode := channelTree.createChannelNode(channel)
		if channel.ParentID == "" {
			channelTree.GetRoot().AddChild(channelNode)
		} else {
			parentNode := tviewutil.GetNodeByReference(channel.ParentID, channelTree.TreeView)
			if parentNode != parentNode {
				parentNode.AddChild(channelNode)
			}
		}
	}
}

// RemoveChannel removes a channels node from the tree.
func (channelTree *ChannelTree) RemoveChannel(channel *discordgo.Channel) {
	channelID := channel.ID

	if channel.Type == discordgo.ChannelTypeGuildText {
		channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
			nodeChannelID, ok := node.GetReference().(string)
			if ok && nodeChannelID == channelID {
				channelTree.removeNode(node, parent, channelID)
				return false
			}

			return true
		})
	} else if channel.Type == discordgo.ChannelTypeGuildCategory {
		node := tviewutil.GetNodeByReference(channelID, channelTree.TreeView)
		if node != nil {
			oldChildren := node.GetChildren()
			node.SetChildren(make([]*tview.TreeNode, 0))
			channelTree.removeNode(node, channelTree.GetRoot(), channelID)
			channelTree.GetRoot().SetChildren(append(channelTree.GetRoot().GetChildren(), oldChildren...))
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

		parent.SetChildren(append(children[:childIndex], children[childIndex+1:]...))
	}
}

// MarkAsUnread marks a channel as unread.
func (channelTree *ChannelTree) MarkAsUnread(channelID string) {
	node := tviewutil.GetNodeByReference(channelID, channelTree.TreeView)
	if node != nil {
		channelTree.channelStates[node] = channelUnread
		channelTree.markNodeAsUnread(node)
	}
}

func (channelTree *ChannelTree) markNodeAsUnread(node *tview.TreeNode) {
	if tview.IsVtxxx {
		node.SetBlinking(true)
		node.SetUnderline(false)
	} else {
		node.SetColor(config.GetTheme().AttentionColor)
	}
}

// MarkAsRead marks a channel as read.
func (channelTree *ChannelTree) MarkAsRead(channelID string) {
	node := tviewutil.GetNodeByReference(channelID, channelTree.TreeView)
	if node != nil {
		channelTree.channelStates[node] = channelRead
		channelTree.markNodeAsRead(node)
	}
}

func (channelTree *ChannelTree) markNodeAsRead(node *tview.TreeNode) {
	if tview.IsVtxxx {
		node.SetBlinking(false)
		node.SetUnderline(false)
	} else {
		node.SetColor(config.GetTheme().PrimaryTextColor)
	}
	node.RemovePrefix(mentionedIndicator)
}

// MarkAsMentioned marks a channel as mentioned.
func (channelTree *ChannelTree) MarkAsMentioned(channelID string) {
	node := tviewutil.GetNodeByReference(channelID, channelTree.TreeView)
	if node != nil {
		channelTree.channelStates[node] = channelMentioned
		channelTree.markNodeAsMentioned(node, channelID)
	}
}

func (channelTree *ChannelTree) markNodeAsMentioned(node *tview.TreeNode, channelID string) {
	channelTree.markNodeAsUnread(node)
	node.AddPrefix(mentionedIndicator)
	node.SortPrefixes(channelTree.prefixSorter)
}

func (channelTree *ChannelTree) prefixSorter(a, b string) bool {
	if a == mentionedIndicator {
		return true
	} else if b == mentionedIndicator {
		return false
	} else if a == nsfwIndicator {
		return true
	} else if b == nsfwIndicator {
		return false
	} else if a == lockedIndicator {
		return true
	} else if b == lockedIndicator {
		return false
	}
	return false
}

// MarkAsLoaded marks a channel as loaded and therefore marks all other
// channels as either unread, read or mentioned.
func (channelTree *ChannelTree) MarkAsLoaded(channelID string) {
	for node, state := range channelTree.channelStates {
		if state == channelLoaded {
			channelTree.channelStates[node] = channelRead
			channelTree.markNodeAsRead(node)
			break
		}
	}

	node := tviewutil.GetNodeByReference(channelID, channelTree.TreeView)
	if node != nil {
		channelTree.channelStates[node] = channelLoaded
		channelTree.markNodeAsLoaded(node)
	}
}

func (channelTree *ChannelTree) markNodeAsLoaded(node *tview.TreeNode) {
	if tview.IsVtxxx {
		node.SetUnderline(true)
		node.SetBlinking(false)
	} else {
		node.SetColor(tview.Styles.ContrastBackgroundColor)
	}
	node.RemovePrefix(mentionedIndicator)
}

// SetOnChannelSelect sets the handler that reacts to channel selection events.
func (channelTree *ChannelTree) SetOnChannelSelect(handler func(channelID string)) {
	channelTree.onChannelSelect = handler
}
