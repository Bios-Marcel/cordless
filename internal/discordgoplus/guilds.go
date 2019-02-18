package discordgoplus

import (
	"sort"

	"github.com/Bios-Marcel/discordgo"
)

func LoadGuilds(session *discordgo.Session) ([]*discordgo.UserGuild, error) {
	guilds := make([]*discordgo.UserGuild, 0)
	var beforeID string

	for {
		newGuilds, discordError := session.UserGuilds(100, beforeID, "")
		if discordError != nil {
			return nil, discordError
		}

		if len(newGuilds) != 0 {
			guilds = append(newGuilds, guilds...)
			if len(newGuilds) == 100 {
				beforeID = newGuilds[0].ID
			} else {
				return guilds, nil
			}
		} else {
			return guilds, nil
		}
	}
}

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
