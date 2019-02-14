package ui

import (
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
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

	onLoad        func(channelID string)
	channelStates map[*tview.TreeNode]channelState
}

// NewChannelTree creates a new ready-to-be-used ChannelTree
func NewChannelTree() *ChannelTree {
	channelTree := &ChannelTree{
		internalTreeView: tview.NewTreeView(),
		channelStates:    make(map[*tview.TreeNode]channelState),
	}

	return channelTree
}

// LoadGuild accesses the state in order to load all locally present channels
// for the passed guild.
func (channelTree *ChannelTree) LoadGuild(guildID string) {

}

// AddOrUpdateChannel either adds a new node for the given channel or updates
// its current node.
func (channelTree *ChannelTree) AddOrUpdateChannel(channel *discordgo.Channel) {

}

// RemoveChannel removes a channels node from the tree.
func (channelTree *ChannelTree) RemoveChannel(channelID string) {

}

// MarkChannelAsUnread marks a channel as unread.
func (channelTree *ChannelTree) MarkChannelAsUnread(channelID string) {

}

// MarkChannelAsMentioned marks a channel as mentioned.
func (channelTree *ChannelTree) MarkChannelAsMentioned(channelID string) {

}

// MarkChannelAsLoaded marks a channel as loaded and therefore marks all other
// channels as either unread, read or mentioned.
func (channelTree *ChannelTree) MarkChannelAsLoaded(channelID string) {

}

// SetOnLoad sets the handler that reacts to channel selection events.
func (channelTree *ChannelTree) SetOnLoad(handler func(channelID string)) {
	channelTree.onLoad = handler
}
