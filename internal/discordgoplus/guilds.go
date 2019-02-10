package discordgoplus

import (
	"sort"

	"github.com/Bios-Marcel/discordgo"
)

// SortGuilds sorts the guilds according to the users settings.
func SortGuilds(settings *discordgo.Settings, guilds []*discordgo.UserGuild) {
	sort.Slice(guilds, func(a, b int) bool {
		aFound := false
		for _, guild := range settings.GuildPositions {
			if aFound {
				if guild == guilds[b].ID {
					return true
				}
			} else {
				if guild == guilds[a].ID {
					aFound = true
				}
			}
		}

		return false
	})
}
