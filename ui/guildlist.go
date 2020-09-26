package ui

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/readstate"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

// GuildList is the UI component to hold all user guilds and allow loading
// one of them.
type GuildList struct {
	*tview.TreeView
	onGuildSelect func(node *tview.TreeNode, guildID string)
}

// NewGuildList creates and initializes a ready to use GuildList.
func NewGuildList(guilds []*discordgo.Guild) *GuildList {
	guildList := &GuildList{
		TreeView: tview.NewTreeView(),
	}

	guildList.
		SetVimBindingsEnabled(config.Current.OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetCycleSelection(true).
		SetTopLevel(1).
		SetBorder(true).
		SetIndicateOverflow(true)

	root := tview.NewTreeNode("")
	guildList.SetRoot(root)
	guildList.SetTitle("Servers").SetTitleAlign(tview.AlignLeft)
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

		guildList.UpdateNodeState(guild, guildNode, false)

		guildNode.SetSelectable(true)
	}

	if len(root.GetChildren()) > 0 {
		guildList.SetCurrentNode(root)
	}

	return guildList
}

// UpdateNodeStateByGuild updates the state of a guilds node accordingly
// to its readstate, unless the node is selected.
//
// FIXME selected should probably be removed here, but bugs will occur
// so I'll do it someday ... :D
func (g *GuildList) UpdateNodeStateByGuild(guild *discordgo.Guild, selected bool) {
	for _, node := range g.GetRoot().GetChildren() {
		if node.GetReference().(string) == guild.ID {
			g.UpdateNodeState(guild, node, selected)
			break
		}
	}
}

// UpdateNodeState updates the state of a node accordingly to its
// readstate, unless the node is selected.
//
// FIXME selected should probably be removed here, but bugs will occur
// so I'll do it someday ... :D
func (g *GuildList) UpdateNodeState(guild *discordgo.Guild, node *tview.TreeNode, selected bool) {
	if selected {
		if vtxxx {
			node.SetAttributes(tcell.AttrUnderline)
		} else {
			node.SetColor(tview.Styles.ContrastBackgroundColor)
		}
	} else {
		if !readstate.HasGuildBeenRead(guild.ID) {
			if vtxxx {
				node.SetAttributes(tcell.AttrBlink)
			} else {
				node.SetColor(config.GetTheme().AttentionColor)
			}
		} else {
			node.SetAttributes(tcell.AttrNone)
			node.SetColor(tview.Styles.PrimaryTextColor)
		}
	}

	if readstate.HasGuildBeenMentioned(guild.ID) {
		node.SetText("(@) " + tviewutil.Escape(guild.Name))
	} else {
		node.SetText(tviewutil.Escape(guild.Name))
	}
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

func (g *GuildList) setNotificationCount(count int) {
	if count == 0 {
		g.SetTitle("Servers")
	} else {
		g.SetTitle(fmt.Sprintf("Servers[%s](%d)", tviewutil.ColorToHex(config.GetTheme().AttentionColor), count))
	}
}

func (g *GuildList) amountOfUnreadGuilds() int {
	var unreadCount int
	for _, child := range g.GetRoot().GetChildren() {
		if !readstate.HasGuildBeenRead((child.GetReference()).(string)) {
			unreadCount++
		}
	}

	return unreadCount
}

// UpdateUnreadGuildCount finds the number of guilds containing unread
// channels and updates the title accordingly.
func (g *GuildList) UpdateUnreadGuildCount() {
	g.setNotificationCount(g.amountOfUnreadGuilds())
}
