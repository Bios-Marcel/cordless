package discordgoplus

import (
	"sort"

	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

var (
	botPrefix = tview.Escape("[BOT]")
)

// GetUserName returns a name for the user.
// Names will be escaped and bots will get the "[BOT]" prefix and a blue color.
func GetUserName(user *discordgo.User, color *string) string {
	var nameToUse string
	if user.Bot {
		nameToUse = "[blue]" + botPrefix
	} else {
		if color != nil {
			nameToUse = nameToUse + "[" + *color + "]"
		}
	}

	nameToUse = nameToUse + tview.Escape(user.Username)

	return nameToUse
}

// GetMemberName returns a name for the member using his nickname if available.
// Names will be escaped and bots get a special "[BOT]" prefix and a blue
// color.
func GetMemberName(member *discordgo.Member, color *string) string {
	var nameToUse string
	if member.User.Bot {
		nameToUse = "[blue]" + botPrefix
	} else {
		if color != nil {
			nameToUse = nameToUse + "[" + *color + "]"
		}
	}

	var discordName string
	if member.Nick != "" {
		discordName = member.Nick
	} else {
		discordName = member.User.Username
	}

	nameToUse = nameToUse + tview.Escape(discordName)

	return nameToUse
}

// SortUserRoles sorts an array of roleIDs according to the guilds roles.
func SortUserRoles(roles []string, guildRoles []*discordgo.Role) {
	sort.Slice(roles, func(a, b int) bool {
		firstIdentifier := roles[a]
		secondIdentifier := roles[b]

		var firstRole *discordgo.Role
		for _, role := range guildRoles {
			if role.ID == firstIdentifier {
				firstRole = role
				break
			}
		}

		var secondRole *discordgo.Role
		for _, role := range guildRoles {
			if role.ID == secondIdentifier {
				secondRole = role
				break
			}
		}

		if secondRole == nil {
			return true
		}

		return firstRole.Position > secondRole.Position
	})
}
