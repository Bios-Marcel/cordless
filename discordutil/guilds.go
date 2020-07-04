package discordutil

import (
	"log"
	"sort"
	"strings"

	"github.com/Bios-Marcel/discordgo"
)

// GuildLoader reflects an instance that allows loading guilds from a discord backend.
type GuildLoader interface {
	UserGuilds(int, string, string) ([]*discordgo.UserGuild, error)
}

// LoadGuilds loads all guilds the current user is part of.
func LoadGuilds(guildLoader GuildLoader) ([]*discordgo.UserGuild, error) {
	guilds := make([]*discordgo.UserGuild, 0)
	var beforeID string

	for {
		newGuilds, discordError := guildLoader.UserGuilds(100, beforeID, "")
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
func SortGuilds(settings *discordgo.Settings, guilds []*discordgo.Guild) {
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

// FindEmojiInGuild searches for a fitting emoji. Fitting means the correct name
// (case insensitive), not animated and the correct permissions. If the result
// is an empty string, it means no result was found.
func FindEmojiInGuild(session *discordgo.Session, guild *discordgo.Guild, omitGWCheck bool, emojiSequence string) string {
	for _, emoji := range guild.Emojis {
		if emoji.Animated {
			continue
		}


		if strings.EqualFold(emoji.Name, emojiSequence) && (omitGWCheck || strings.HasPrefix(emoji.Name, "GW")) {
			if len(emoji.Roles) != 0 {
				selfMember, cacheError := session.State.Member(guild.ID, session.State.User.ID)
				if cacheError != nil {
					selfMember, discordError := session.GuildMember(guild.ID, session.State.User.ID)
					if discordError != nil {
						log.Println(discordError)
						continue
					}

					session.State.MemberAdd(selfMember)
				}

				if selfMember != nil {
					for _, emojiRole := range emoji.Roles {
						for _, selfRole := range selfMember.Roles {
							if selfRole == emojiRole {
								return "<:" + emoji.Name + ":" + emoji.ID + ">"
							}
						}
					}
				}
			}

			return "<:" + emoji.Name + ":" + emoji.ID + ">"
		}
	}

	return ""
}
