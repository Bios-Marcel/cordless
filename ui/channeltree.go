package ui

import (
	"os"
	"regexp"
	"sort"
	"sync"

	"github.com/gdamore/tcell"

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

func checkVT() bool {
	VTxxx, err := regexp.MatchString("(vt)[0-9]+", os.Getenv("TERM"))
	if err != nil {
		panic(err)
	}
	return VTxxx
}

var vtxxx = checkVT()

// ChannelTree is the component that displays the channel hierarchy of the
// currently loaded guild and allows interactions with those channels.
type ChannelTree struct {
	*tview.TreeView

	state *discordgo.State

	onChannelSelect func(channelID string)
	channelStates   map[*tview.TreeNode]channelState
	channelPosition map[string]int

	mutex *sync.Mutex
}

// NewChannelTree creates a new ready-to-be-used ChannelTree
func NewChannelTree(state *discordgo.State) *ChannelTree {
	channelTree := &ChannelTree{
		state:           state,
		TreeView:        tview.NewTreeView(),
		channelStates:   make(map[*tview.TreeNode]channelState),
		channelPosition: make(map[string]int),
		mutex:           &sync.Mutex{},
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
	guild, cacheError := channelTree.state.Guild(guildID)
	if cacheError != nil {
		return cacheError
	}

	channelTree.Clear()

	channels := guild.Channels
	sort.Slice(channels, func(a, b int) bool {
		return channels[a].Position < channels[b].Position
	})

	// Top level channel
	state := channelTree.state
	for _, channel := range channels {
		if (channel.Type != discordgo.ChannelTypeGuildText && channel.Type != discordgo.ChannelTypeGuildNews) ||
			channel.ParentID != "" || !discordutil.HasReadMessagesPermission(channel.ID, state) {
			continue
		}
		createTopLevelChannelNodes(channelTree, channel)
	}
	// Categories
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
					// therefore we add the category and jump to the next
					createChannelCategoryNode(channelTree, channel)
					continue CATEGORY_LOOP
				}
				childless = false
			}
		}

		//If the category is childless, we want to add it anyway.
		if childless {
			createChannelCategoryNode(channelTree, channel)
		}
	}
	// Second level channel
	for _, channel := range channels {
		if (channel.Type != discordgo.ChannelTypeGuildText && channel.Type != discordgo.ChannelTypeGuildNews) ||
			channel.ParentID == "" || !discordutil.HasReadMessagesPermission(channel.ID, state) {
			continue
		}
		createSecondLevelChannelNodes(channelTree, channel)
	}
	channelTree.SetCurrentNode(channelTree.GetRoot())
	return nil
}

func createTopLevelChannelNodes(channelTree *ChannelTree, channel *discordgo.Channel) {
	channelNode := channelTree.createTextChannelNode(channel)
	channelTree.GetRoot().AddChild(channelNode)
}

func createChannelCategoryNode(channelTree *ChannelTree, channel *discordgo.Channel) {
	channelNode := createChannelNode(channel)
	channelNode.SetSelectable(false)
	channelTree.GetRoot().AddChild(channelNode)
}

func createSecondLevelChannelNodes(channelTree *ChannelTree, channel *discordgo.Channel) {
	for _, node := range channelTree.GetRoot().GetChildren() {
		channelID, ok := node.GetReference().(string)
		if ok && channelID == channel.ParentID {
			channelNode := channelTree.createTextChannelNode(channel)
			node.AddChild(channelNode)
			break
		}
	}
}

func createChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := tview.NewTreeNode(channel.Name)
	var prefixes string
	if channel.NSFW {
		prefixes += tviewutil.Escape("🔞")
	}

	// Adds a padlock prefix if the channel if not readable by the everyone group
	if config.Current.IndicateChannelAccessRestriction {
		for _, permission := range channel.PermissionOverwrites {
			if permission.Type == "role" && permission.ID == channel.GuildID && permission.Deny&discordgo.PermissionViewChannel == discordgo.PermissionViewChannel {
				prefixes += tviewutil.Escape("\U0001F512")
			}
		}
	}

	channelNode.SetPrefix(prefixes)

	channelNode.SetReference(channel.ID)
	return channelNode
}

func (channelTree *ChannelTree) createTextChannelNode(channel *discordgo.Channel) *tview.TreeNode {
	channelNode := createChannelNode(channel)

	if !readstate.HasBeenRead(channel, channel.LastMessageID) {
		channelTree.channelStates[channelNode] = channelUnread
		if vtxxx {
			channelNode.SetAttributes(tcell.AttrBlink)
		} else {
			channelNode.SetColor(config.GetTheme().AttentionColor)
		}
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
			//TODO Do the moving somehow
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
		channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
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

		parent.SetChildren(append(children[:childIndex], children[childIndex+1:]...))
	}
}

// MarkChannelAsUnread marks a channel as unread.
func (channelTree *ChannelTree) MarkChannelAsUnread(channelID string) {
	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channelTree.channelStates[node] = channelUnread
			if vtxxx {
				node.SetAttributes(tcell.AttrBlink)
			} else {
				node.SetColor(config.GetTheme().AttentionColor)
			}
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
				node.SetText(tviewutil.Escape(channel.Name))
			}

			if channelTree.channelStates[node] != channelLoaded {
				channelTree.channelStates[node] = channelRead
				if vtxxx {
					node.SetAttributes(tcell.AttrNone)
				} else {
					node.SetColor(config.GetTheme().PrimaryTextColor)
				}
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
			channelTree.markNodeAsMentioned(node, channelID)
			return false
		}

		return true
	})
}

func (channelTree *ChannelTree) markNodeAsMentioned(node *tview.TreeNode, channelID string) {
	channelTree.channelStates[node] = channelMentioned
	channel, stateError := channelTree.state.Channel(channelID)
	if stateError == nil {
		node.SetText("(@) " + tviewutil.Escape(channel.Name))
	}
	if vtxxx {
		node.SetAttributes(tcell.AttrBlink)
	} else {
		node.SetColor(config.GetTheme().AttentionColor)
	}
}

// MarkChannelAsLoaded marks a channel as loaded and therefore marks all other
// channels as either unread, read or mentioned.
func (channelTree *ChannelTree) MarkChannelAsLoaded(channelID string) {
	for node, state := range channelTree.channelStates {
		if state == channelLoaded {
			channelTree.channelStates[node] = channelRead
			if vtxxx {
				node.SetAttributes(tcell.AttrNone)
			} else {
				node.SetColor(config.GetTheme().PrimaryTextColor)
			}
			break
		}
	}

	channelTree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		referenceChannelID, ok := node.GetReference().(string)
		if ok && referenceChannelID == channelID {
			channelTree.channelStates[node] = channelLoaded
			channel, stateError := channelTree.state.Channel(channelID)
			if stateError == nil {
				node.SetText(tviewutil.Escape(channel.Name))
			}
			if vtxxx {
				node.SetAttributes(tcell.AttrUnderline)
			} else {
				node.SetColor(tview.Styles.ContrastBackgroundColor)
			}
			return false
		}

		return true
	})
}

// SetOnChannelSelect sets the handler that reacts to channel selection events.
func (channelTree *ChannelTree) SetOnChannelSelect(handler func(channelID string)) {
	channelTree.onChannelSelect = handler
}

// Lock will lock the ChannelTree, allowing other callers to prevent race
// conditions.
func (channelTree *ChannelTree) Lock() {
	channelTree.mutex.Lock()
}

// Unlock unlocks the previously locked ChannelTree.
func (channelTree *ChannelTree) Unlock() {
	channelTree.mutex.Unlock()
}
