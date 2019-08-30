package discordutil

import (
	"math/rand"
	"sort"

	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

var (
	botPrefix = tview.Escape("[BOT]")
	botColor  = "#9496fc"

	colors = []string{
		"#d8504e",
		"#d87e4e",
		"#d8a54e",
		"#d8c64e",
		"#b8d84e",
		"#91d84e",
		"#67d84e",
		"#4ed87c",
		"#4ed8aa",
		"#4ed8cf",
		"#4eb6d8",
		"#4e57d8",
		"#754ed8",
		"#a34ed8",
		"#cf4ed8",
		"#d84e9c",
	}
	//global state that persist during a session.
	userColorCache = make(map[string]string)

	lastRandomNumber = -1
)

// GetUserColor gets the users color for this session. If no color can be found
// a new color will be a generated and cached.
func GetUserColor(user *discordgo.User) string {
	if user.Bot {
		return botColor
	}

	color, ok := userColorCache[user.ID]
	if ok {
		return color
	}

	newColor := getRandomColorString()
	userColorCache[user.ID] = newColor
	return newColor
}

func getRandomColorString() string {
	randIndex := rand.Intn(len(colors))
	if randIndex == lastRandomNumber {
		return getRandomColorString()
	}

	lastRandomNumber = randIndex

	return colors[randIndex]
}

// GetMemberName returns the name to use for representing this user. This is
// either the username or the nickname. In case the member is a bot, the bot
// prefix will be prepended.
func GetMemberName(member *discordgo.Member) string {
	if member.Nick != "" {
		return getUserName(member.Nick, member.User.Bot)
	}

	return GetUserName(member.User)
}

// GetUserName returns the users username, prepending the bot prefix in case he
// is a bot.
func GetUserName(user *discordgo.User) string {
	return getUserName(user.Username, user.Bot)
}

func getUserName(name string, bot bool) string {
	discordName := tview.Escape(name)

	if bot {
		return botPrefix + discordName
	}

	return discordName
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
			return false
		}

		if firstRole == nil {
			return true
		}

		return firstRole.Position > secondRole.Position
	})
}

// IsBlocked checks whether the state contains any relationship that says the
// given user has been blocked.
func IsBlocked(state *discordgo.State, user *discordgo.User) bool {
	for _, relationship := range state.Relationships {
		if relationship.User.ID == user.ID &&
			relationship.Type == discordgo.RelationTypeBlocked {
			return true
		}
	}

	return false
}
