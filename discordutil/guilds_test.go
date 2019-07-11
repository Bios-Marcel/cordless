package discordutil

import (
	"testing"

	"github.com/Bios-Marcel/discordgo"
)

func TestSortGuilds(t *testing.T) {
	//The IDs have random length and don't mean much.
	guildOneID := "98122541287"
	guildTwoID := "12450501965"
	guildThreeID := "1086518z963"
	guildFourID := "19651241842"
	settings := &discordgo.Settings{
		GuildPositions: []string{
			guildOneID,
			guildTwoID,
			guildThreeID,
			guildFourID,
		},
	}

	guilds := []*discordgo.Guild{
		&discordgo.Guild{ID: guildThreeID},
		&discordgo.Guild{ID: guildOneID},
		&discordgo.Guild{ID: guildFourID},
		&discordgo.Guild{ID: guildTwoID},
	}

	SortGuilds(settings, guilds)

	if guilds[0].ID != guildOneID {
		t.Errorf("The first guild should've been %s, but was %s", guildOneID, guilds[0].ID)
	}

	if guilds[1].ID != guildTwoID {
		t.Errorf("The second guild should've been %s, but was %s", guildTwoID, guilds[1].ID)
	}

	if guilds[2].ID != guildThreeID {
		t.Errorf("The third guild should've been %s, but was %s", guildThreeID, guilds[2].ID)
	}

	if guilds[3].ID != guildFourID {
		t.Errorf("The fourth guild should've been %s, but was %s", guildFourID, guilds[3].ID)
	}
}
