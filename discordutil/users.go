package discordutil

import (
	"math/rand"
	"sort"

	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/cordless/tview"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

var (
	botPrefix = tview.Escape("[BOT]")

	//global state that persist during a session.
	userColorCache = make(map[string]string)

	lastRandomNumber = -1
	randColorLength  = len(config.GetTheme().RandomUserColors)
)

// GetUserColor gets the users color for this session. If no color can be found
// a new color will be a generated and cached.
func GetUserColor(user *discordgo.User) string {
	if user.Bot {
		return tviewutil.ColorToHex(config.GetTheme().BotColor)
	}

	//Avoid unnecessarily retrieving and caching colors
	if !config.Current.UseRandomUserColors || randColorLength == 0 {
		return tviewutil.ColorToHex(config.GetTheme().DefaultUserColor)
	} else if randColorLength == 1 {
		return tviewutil.ColorToHex(config.GetTheme().RandomUserColors[0])
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
	randIndex := rand.Intn(randColorLength)
	if randIndex == lastRandomNumber {
		return getRandomColorString()
	}

	lastRandomNumber = randIndex

	return tviewutil.ColorToHex(config.GetTheme().RandomUserColors[randIndex])
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
	discordName := tviewutil.Escape(name)

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
