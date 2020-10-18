package ui

import (
	"sort"
	"sync"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/discordutil"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"
	"github.com/gdamore/tcell"
)

// UserTree represents the visual list of users in a guild.
type UserTree struct {
	*sync.Mutex

	internalTreeView *tview.TreeView

	state *discordgo.State

	userNodes map[string]*tview.TreeNode
	roleNodes map[string]*tview.TreeNode
	roles     []*discordgo.Role

	loadedFor interface{}
}

// NewUserTree creates a new pre-configured UserTree that is empty.
func NewUserTree(state *discordgo.State) *UserTree {
	userTree := &UserTree{
		state:            state,
		internalTreeView: tview.NewTreeView(),
		Mutex:            &sync.Mutex{},
	}

	userTree.internalTreeView.
		SetVimBindingsEnabled(config.Current.OnTypeInListBehaviour == config.DoNothingOnTypeInList).
		SetRoot(tview.NewTreeNode("")).
		SetTopLevel(1).
		SetCycleSelection(true).
		SetBorder(true)

	return userTree
}

// Clear removes all nodes and data out of the view.
func (userTree *UserTree) Clear() {
	userTree.Lock()
	defer userTree.Unlock()
	userTree.clear()
}

func (userTree *UserTree) rootNode() *tview.TreeNode {
	return userTree.internalTreeView.GetRoot()
}

func (userTree *UserTree) clear() {
	for _, roleNode := range userTree.roleNodes {
		roleNode.ClearChildren()
	}
	userTree.rootNode().ClearChildren()
	userTree.loadedFor = nil

	// After clearing, we don't reallocate anything, since we don't know
	// whether we actually want to repopulate the tree.
	userTree.userNodes = nil
	userTree.roleNodes = nil
	userTree.roles = nil
}

// LoadGroup loads all users for a group-channel.
func (userTree *UserTree) LoadGroup(channelID string) error {
	userTree.Lock()
	defer userTree.Unlock()

	channel, stateError := userTree.state.PrivateChannel(channelID)
	if stateError != nil {
		return stateError
	}

	//Already loaded
	if channel == userTree.loadedFor {
		return nil
	}

	userTree.clear()
	userTree.userNodes = make(map[string]*tview.TreeNode)
	userTree.roleNodes = make(map[string]*tview.TreeNode)

	userTree.addOrUpdateUsers(channel.Recipients)

	userTree.loadedFor = channel
	userTree.selectFirstNode()

	return nil
}

// LoadGuild will load all available roles of the guild and then load all
// available members. Afterwards the first available node will be selected.
func (userTree *UserTree) LoadGuild(guildID string) error {
	userTree.Lock()
	defer userTree.Unlock()

	guild, stateError := userTree.state.Guild(guildID)
	if stateError != nil {
		return stateError
	}

	//Already loaded
	if guild == userTree.loadedFor {
		return nil
	}

	userTree.clear()
	userTree.userNodes = make(map[string]*tview.TreeNode)
	userTree.roleNodes = make(map[string]*tview.TreeNode)

	guildRoles, roleLoadError := userTree.loadGuildRoles(guild)
	if roleLoadError != nil {
		return roleLoadError
	}
	userTree.roles = guildRoles

	userLoadError := userTree.loadGuildMembers(guild)
	if userLoadError != nil {
		return userLoadError
	}

	userTree.loadedFor = guild
	userTree.selectFirstNode()

	return nil
}

func (userTree *UserTree) selectFirstNode() {
	if userTree.internalTreeView.GetCurrentNode() == nil {
		userNodes := userTree.rootNode().GetChildren()
		if len(userNodes) > 0 {
			userTree.internalTreeView.SetCurrentNode(userTree.rootNode().GetChildren()[0])
		}
	}
}

func (userTree *UserTree) loadGuildMembers(guild *discordgo.Guild) error {
	members, stateError := userTree.state.Members(guild.ID)
	if stateError != nil {
		return stateError
	}

	userTree.addOrUpdateMembers(members)

	return nil
}

func (userTree *UserTree) loadGuildRoles(guild *discordgo.Guild) ([]*discordgo.Role, error) {
	guildRoles := guild.Roles

	sort.Slice(guildRoles, func(a, b int) bool {
		return guildRoles[a].Position > guildRoles[b].Position
	})

	for _, role := range guildRoles {
		if role.Hoist {
			roleNode := tview.NewTreeNode("[" + discordutil.GetRoleColor(role) +
				"]" + tviewutil.Escape(role.Name))
			roleNode.SetSelectable(false)
			userTree.roleNodes[role.ID] = roleNode
			userTree.rootNode().AddChild(roleNode)
		}
	}

	return guildRoles, nil
}

// AddOrUpdateMember adds the passed member to the tree, unless it is
// already part of the tree, in that case the nodes name is updated.
func (userTree *UserTree) AddOrUpdateMember(member *discordgo.Member) {
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	userTree.addOrUpdateMember(member)
}

func (userTree *UserTree) addOrUpdateMember(member *discordgo.Member) {
	nameToUse := "[" + discordutil.GetMemberColor(userTree.state, member) +
		"]" + discordutil.GetMemberName(member)

	userNode, contains := userTree.userNodes[member.User.ID]
	if contains && userNode != nil {
		userNode.SetText(nameToUse)
		return
	}

	userNode = tview.NewTreeNode(nameToUse)
	userTree.userNodes[member.User.ID] = userNode

	discordutil.SortUserRoles(member.Roles, userTree.roles)

	for _, userRole := range member.Roles {
		roleNode, exists := userTree.roleNodes[userRole]
		if exists && roleNode != nil {
			roleNode.AddChild(userNode)
			return
		}
	}

	userTree.rootNode().AddChild(userNode)
}

// AddOrUpdateUser adds a user to the tree, unless the user already exists,
// in that case the users node gets updated.
func (userTree *UserTree) AddOrUpdateUser(user *discordgo.User) {
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	userTree.addOrUpdateUser(user)
}

func (userTree *UserTree) addOrUpdateUser(user *discordgo.User) {
	nameToUse := "[" + discordutil.GetUserColor(user) +
		"]" + discordutil.GetUserName(user)

	userNode, contains := userTree.userNodes[user.ID]
	if contains && userNode != nil {
		userNode.SetText(nameToUse)
		return
	}

	userNode = tview.NewTreeNode(nameToUse)
	userTree.userNodes[user.ID] = userNode
	userTree.rootNode().AddChild(userNode)
}

// AddOrUpdateUsers adds users to the tree, unless they already exists, in that
// case the users nodes gets updated.
func (userTree *UserTree) AddOrUpdateUsers(users []*discordgo.User) {
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	userTree.addOrUpdateUsers(users)
}

func (userTree *UserTree) addOrUpdateUsers(users []*discordgo.User) {
	for _, user := range users {
		userTree.addOrUpdateUser(user)
	}
}

// AddOrUpdateMembers adds the all passed members to the tree, unless a node is
// already part of the tree, in that case the nodes name is updated.
func (userTree *UserTree) AddOrUpdateMembers(members []*discordgo.Member) {
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	userTree.addOrUpdateMembers(members)
}

func (userTree *UserTree) addOrUpdateMembers(members []*discordgo.Member) {
	for _, member := range members {
		userTree.addOrUpdateMember(member)
	}
}

// RemoveMember finds and removes a node from the tree.
func (userTree *UserTree) RemoveMember(member *discordgo.Member) {
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	userTree.removeMember(member)
}

func (userTree *UserTree) removeMember(member *discordgo.Member) {
	userNode, contains := userTree.userNodes[member.User.ID]
	if contains {
		userTree.rootNode().Walk(func(node, parent *tview.TreeNode) bool {
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
	userTree.Lock()
	defer userTree.Unlock()
	if !userTree.isLoaded() {
		return
	}
	for _, member := range members {
		userTree.removeMember(member)
	}
}

//SetInputCapture delegates to tviews SetInputCapture
func (userTree *UserTree) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) {
	userTree.internalTreeView.SetInputCapture(capture)
}

// isLoaded indicates whether the UserList is currently displaying any data.
func (userTree *UserTree) isLoaded() bool {
	return userTree.loadedFor != nil
}
