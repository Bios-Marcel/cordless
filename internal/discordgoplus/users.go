package discordgoplus

import (
	"crypto/rand"
	"math/big"
	"sort"

	"github.com/Bios-Marcel/discordgo"
	"github.com/Bios-Marcel/tview"
)

var (
	botPrefix = tview.Escape("[BOT]")

	colors = []string{
		"maroon",
		"green",
		"olive",
		"navy",
		"purple",
		"teal",
		"red",
		"lime",
		"yellow",
		"blue",
		"fuchsia",
		"aqua",
	}
	//global state that persist during a session.
	userColorCache = make(map[string]string)

	lastRandomNumber int64 = -1
)

// GetUserName returns a name for the user.
// Names will be escaped and bots will get the "[BOT]" prefix and a blue color.
func GetUserName(user *discordgo.User) string {
	color := GetUserColor(user.ID)
	var nameToUse string
	if user.Bot {
		nameToUse = "[blue]" + botPrefix
	} else {
		if color != "" {
			nameToUse = nameToUse + "[" + color + "]"
		}
	}

	nameToUse = nameToUse + tview.Escape(user.Username)

	return nameToUse
}

// GetUserColor gets the users color for this session. If no color can be found
// a new color will be a generated and cached.
func GetUserColor(userID string) string {
	color, ok := userColorCache[userID]
	if ok {
		return color
	}

	newColor := getRandomColorString()
	userColorCache[userID] = newColor
	return newColor
}

func getRandomColorString() string {
	number, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors) )))
	if err != nil {
		return colors[0]
	}

	numberInt64 := number.Int64()
	if numberInt64 == lastRandomNumber {
		return getRandomColorString()
	}

	return colors[numberInt64]
}

// GetMemberName returns a name for the member using his nickname if available.
// Names will be escaped and bots get a special "[BOT]" prefix and a blue
// color.
func GetMemberName(member *discordgo.Member) string {
	var discordName string
	if member.Nick != "" {
		discordName = tview.Escape(member.Nick)
	} else {
		discordName = tview.Escape(member.User.Username)
	}

	if member.User.Bot {
		return "[blue]" + botPrefix + discordName
	}

	return "[" + GetUserColor(member.User.ID) + "]" + discordName
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
