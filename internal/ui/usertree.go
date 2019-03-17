package ui

import (
	"sort"

	"github.com/Bios-Marcel/cordless/internal/config"
	"github.com/Bios-Marcel/cordless/internal/discordgoplus"
	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

// UserTree represents the visual list of users in a guild.
type UserTree struct {
	internalTreeView *tview.TreeView
	rootNode         *tview.TreeNode

	state *discordgo.State

	userNodes map[string]*tview.TreeNode

	roleNodes map[string]*tview.TreeNode
	roles     []*discordgo.Role
}

// NewUserTree creates a new pre-configured UserTree that is empty.
func NewUserTree(state *discordgo.State) *UserTree {
	userTree := &UserTree{
		state:            state,
		userNodes:        make(map[string]*tview.TreeNode),
		roleNodes:        make(map[string]*tview.TreeNode),
		roles:            make([]*discordgo.Role, 0),
		rootNode:         tview.NewTreeNode(""),
		internalTreeView: tview.NewTreeView(),
	}

	userTree.internalTreeView.
		SetVimBindingsEnabled(config.GetConfig().OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetRoot(userTree.rootNode).
		SetTopLevel(1).
		SetCycleSelection(true)
	userTree.internalTreeView.SetBorder(true)

	return userTree
}

// Clear removes all nodes and data out of the view.
func (userTree *UserTree) Clear() {
	for _, roleNode := range userTree.roleNodes {
		roleNode.ClearChildren()
	}

	userTree.userNodes = make(map[string]*tview.TreeNode)
	userTree.roleNodes = make(map[string]*tview.TreeNode)
	userTree.roles = make([]*discordgo.Role, 0)

	userTree.rootNode.ClearChildren()
}

// LoadGroup loads all users for a group-channel.
func (userTree *UserTree) LoadGroup(channelID string) error {
	userTree.Clear()

	channel, stateError := userTree.state.PrivateChannel(channelID)
	if stateError != nil {
		return stateError
	}

	userTree.AddOrUpdateUsers(channel.Recipients)

	userTree.selectFirstNode()

	return nil
}

// LoadGuild will load all available roles of the guild and then load all
// available members. Afterwards the first available node will be selected.
func (userTree *UserTree) LoadGuild(guildID string) error {
	userTree.Clear()

	guildRoles, roleLoadError := userTree.loadGuildRoles(guildID)
	if roleLoadError != nil {
		return roleLoadError
	}
	userTree.roles = guildRoles

	userLoadError := userTree.loadGuildMembers(guildID)
	if userLoadError != nil {
		return userLoadError
	}

	userTree.selectFirstNode()

	return nil
}

func (userTree *UserTree) selectFirstNode() {
	if userTree.internalTreeView.GetCurrentNode() == nil {
		userNodes := userTree.rootNode.GetChildren()
		if len(userNodes) > 0 {
			userTree.internalTreeView.SetCurrentNode(userTree.rootNode.GetChildren()[0])
		}
	}
}

func (userTree *UserTree) loadGuildMembers(guildID string) error {
	members, stateError := userTree.state.Members(guildID)
	if stateError != nil {
		return stateError
	}

	userTree.AddOrUpdateMembers(members)

	return nil
}

func (userTree *UserTree) loadGuildRoles(guildID string) ([]*discordgo.Role, error) {
	guild, stateError := userTree.state.Guild(guildID)
	if stateError != nil {
		return nil, stateError
	}

	guildRoles := guild.Roles

	sort.Slice(guildRoles, func(a, b int) bool {
		return guildRoles[a].Position > guildRoles[b].Position
	})

	for _, role := range guildRoles {
		if role.Hoist {
			roleNode := tview.NewTreeNode(role.Name)
			roleNode.SetSelectable(false)
			userTree.roleNodes[role.ID] = roleNode
			userTree.rootNode.AddChild(roleNode)
		}
	}

	return guildRoles, nil
}

// AddOrUpdateMember adds the passed member to the tree, unless it is
// already part of the tree, in that case the nodes name is updated.
func (userTree *UserTree) AddOrUpdateMember(member *discordgo.Member) {
	nameToUse := discordgoplus.GetMemberName(member)
	if config.GetConfig().UseRandomUserColors {
		nameToUse = "[" + discordgoplus.GetUserColor(member.User) + "]" + nameToUse
	}

	userNode, contains := userTree.userNodes[member.User.ID]
	if contains && userNode != nil {
		userNode.SetText(nameToUse)
		return
	}

	userNode = tview.NewTreeNode(nameToUse)
	userTree.userNodes[member.User.ID] = userNode

	discordgoplus.SortUserRoles(member.Roles, userTree.roles)

	for _, userRole := range member.Roles {
		roleNode, exists := userTree.roleNodes[userRole]
		if exists && roleNode != nil {
			roleNode.AddChild(userNode)
			return
		}
	}

	userTree.rootNode.AddChild(userNode)
}

// AddOrUpdateUser adds a user to the tree, unless the user already exists,
// in that case the users node gets updated.
func (userTree *UserTree) AddOrUpdateUser(user *discordgo.User) {
	nameToUse := discordgoplus.GetUserName(user)
	if config.GetConfig().UseRandomUserColors {
		nameToUse = "[" + discordgoplus.GetUserColor(user) + "]" + nameToUse
	}

	userNode, contains := userTree.userNodes[user.ID]
	if contains && userNode != nil {
		userNode.SetText(nameToUse)
		return
	}

	userNode = tview.NewTreeNode(nameToUse)
	userTree.userNodes[user.ID] = userNode
	userTree.rootNode.AddChild(userNode)
}

// AddOrUpdateUsers adds users to the tree, unless they already exists, in that
// case the users nodes gets updated.
func (userTree *UserTree) AddOrUpdateUsers(users []*discordgo.User) {
	for _, user := range users {
		userTree.AddOrUpdateUser(user)
	}
}

// AddOrUpdateMembers adds the all passed members to the tree, unless a node is
// already part of the tree, in that case the nodes name is updated.
func (userTree *UserTree) AddOrUpdateMembers(members []*discordgo.Member) {
	for _, member := range members {
		userTree.AddOrUpdateMember(member)
	}
}

// RemoveMember finds and removes a node from the tree.
func (userTree *UserTree) RemoveMember(member *discordgo.Member) {
	userNode, contains := userTree.userNodes[member.User.ID]
	if contains {
		userTree.rootNode.Walk(func(node, parent *tview.TreeNode) bool {
			if node == userNode {
				if len(parent.GetChildren()) == 1 {
					parent.SetChildren(make([]*tview.TreeNode, 0))
				} else {
					indexToDelete := -1
					for index, child := range parent.GetChildren() {
						if child == node {
							indexToDelete = index
							break
						}
					}

					if indexToDelete == 0 {
						parent.SetChildren(parent.GetChildren()[1:])
					} else if indexToDelete == len(parent.GetChildren())-1 {
						parent.SetChildren(parent.GetChildren()[:len(parent.GetChildren())-1])
					} else {
						parent.SetChildren(append(parent.GetChildren()[0:indexToDelete],
							parent.GetChildren()[indexToDelete+1:]...))
					}
				}

				return false
			}

			return true
		})
	}
}

// RemoveMembers finds and removes all passed members from the tree.
func (userTree *UserTree) RemoveMembers(members []*discordgo.Member) {
	for _, member := range members {
		userTree.RemoveMember(member)
	}
}

//SetInputCapture delegates to tviews SetInputCapture
func (userTree *UserTree) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) {
	userTree.internalTreeView.SetInputCapture(capture)
}
