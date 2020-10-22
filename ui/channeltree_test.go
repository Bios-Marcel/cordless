package ui

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"
	tcell "github.com/gdamore/tcell/v2"
)

func TestChannelTree(t *testing.T) {
	simScreen := tcell.NewSimulationScreen("UTF-8")

	simScreen.Init()
	simScreen.SetSize(10, 10)

	state := discordgo.NewState()

	c1 := &discordgo.Channel{
		ID:       "C1",
		Name:     "C1",
		Position: 2,
	}
	c2 := &discordgo.Channel{
		ID:       "C2",
		Name:     "C2",
		Position: 1,
	}
	c3 := &discordgo.Channel{
		ID:       "C3",
		Name:     "C3",
		Position: 3,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   "R1",
				Type: "role",
				Deny: discordgo.PermissionViewChannel,
			},
		},
	}
	g1 := &discordgo.Guild{
		ID:   "G1",
		Name: "G1",
		Channels: []*discordgo.Channel{
			c1,
			c2,
			c3,
		},
	}
	c1.GuildID = g1.ID
	c2.GuildID = g1.ID
	c3.GuildID = g1.ID

	stateError := state.GuildAdd(g1)
	if stateError != nil {
		t.Errorf("Error initializing state: %s", stateError)
	}

	stateError = state.ChannelAdd(c1)
	if stateError != nil {
		t.Errorf("Error initializing state: %s", stateError)
	}
	stateError = state.ChannelAdd(c2)
	if stateError != nil {
		t.Errorf("Error initializing state: %s", stateError)
	}

	state.User = &discordgo.User{
		ID: "U1",
	}

	r1 := &discordgo.Role{
		ID:          "R1",
		Name:        "Rollo",
		Permissions: discordgo.PermissionViewChannel,
	}
	state.RoleAdd("G1", r1)

	state.MemberAdd(&discordgo.Member{
		GuildID: g1.ID,
		User:    state.User,
		Roles:   []string{r1.ID},
	})

	tree := NewChannelTree(state)
	loadError := tree.LoadGuild("G1")

	if loadError != nil {
		t.Errorf("Error loading channeltree: %s", loadError)
	}

	tree.SetBorder(false)
	tree.SetRect(0, 0, 10, 10)

	tree.Draw(simScreen)

	expectCell('C', 0, 0, simScreen, t)
	expectCell('2', 1, 0, simScreen, t)

	expectCell('C', 0, 1, simScreen, t)
	expectCell('1', 1, 1, simScreen, t)

	expectCell(' ', 0, 2, simScreen, t)
	expectCell(' ', 1, 2, simScreen, t)
}

func expectCell(expected rune, column, row int, screen tcell.SimulationScreen, t *testing.T) {
	cell, _, _, _ := screen.GetContent(column, row)
	if cell != expected {
		t.Errorf("Cell missmatch. Was '%c' instead of '%c'.", cell, expected)
	}
}
