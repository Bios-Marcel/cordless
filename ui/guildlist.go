package ui

import (
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

// GuildList is the UI component to hold all user guilds and allow loading
// one of them.
type GuildList struct {
	*tview.TreeView
	onGuildSelect func(node *tview.TreeNode, guildID string)
}

// NewGuildList creates and initializes a ready to use GuildList.
func NewGuildList(guilds []*discordgo.Guild, window *Window) *GuildList {
	guildList := &GuildList{
		TreeView: tview.NewTreeView(),
	}

	guildList.
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true).
		SetTopLevel(1).
		SetBorder(true).
		SetIndicateOverflow(true)

	root := tview.NewTreeNode("")
	guildList.SetRoot(root)
	guildList.SetSelectedFunc(func(node *tview.TreeNode) {
		guildID, ok := node.GetReference().(string)
		if ok && guildList.onGuildSelect != nil {
			guildList.onGuildSelect(node, guildID)
		}
	})

	for _, guild := range guilds {
		//Guilds with an empty name are incomplete and we still have to wait
		//for the respective GuildCreate event to be sent to us.
		if guild.Name == "" {
			continue
		}
		guildNode := tview.NewTreeNode(tviewutil.Escape(guild.Name))
		guildNode.SetReference(guild.ID)
		root.AddChild(guildNode)

		window.updateServerReadStatus(guild.ID, guildNode, false)

		guildNode.SetSelectable(true)
	}

	if len(root.GetChildren()) > 0 {
		guildList.SetCurrentNode(root)
	}

	return guildList
}

// SetOnGuildSelect sets the handler for when a guild is selected.
func (g *GuildList) SetOnGuildSelect(handler func(node *tview.TreeNode, guildID string)) {
	g.onGuildSelect = handler
}

// RemoveGuild removes the node that refers to the given guildID.
func (g *GuildList) RemoveGuild(guildID string) {
	children := g.GetRoot().GetChildren()
	indexToRemove := -1
	for index, node := range children {
		if node.GetReference() == guildID {
			indexToRemove = index
			break
		}
	}

	if indexToRemove != -1 {
		g.GetRoot().SetChildren(append(children[:indexToRemove], children[indexToRemove+1:]...))
	}
}

// AddGuild adds a new node that references the given guildID and shows the
// given name.
func (g *GuildList) AddGuild(guildID, name string) {
	node := tview.NewTreeNode(tviewutil.Escape(name))
	node.SetReference(guildID)
	g.GetRoot().AddChild(node)
}

// UpdateName updates the name of the guild with the given ID.
func (g *GuildList) UpdateName(guildID, newName string) {
	for _, node := range g.GetRoot().GetChildren() {
		if node.GetReference() == guildID {
			node.SetText(tviewutil.Escape(newName))
			break
		}
	}
}
