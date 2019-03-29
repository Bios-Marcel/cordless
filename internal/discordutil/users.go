package discordutil

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

	lastRandomNumber int64 = -1
)

// GetUserColor gets the users color for this session. If no color can be found
// a new color will be a generated and cached.
func GetUserColor(user *discordgo.User) string {
	if user.Bot {
		return "#9496fc"
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
	number, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return colors[0]
	}

	numberInt64 := number.Int64()
	if numberInt64 == lastRandomNumber {
		return getRandomColorString()
	}

	lastRandomNumber = numberInt64

	return colors[numberInt64]
}

// GetMemberName returns the name to use for representing this user. This is
// either the username or the nickname. In case the member is a bot, the bot
// prefix will be prepended.
func GetMemberName(member *discordgo.Member) string {
	var discordName string
	if member.Nick != "" {
		discordName = tview.Escape(member.Nick)
	} else {
		discordName = tview.Escape(member.User.Username)
	}

	if member.User.Bot {
		return botPrefix + discordName
	}

	return discordName
}

// GetUserName returns the users username, prepending the bot prefix in case he
// is a bot.
func GetUserName(user *discordgo.User) string {
	discordName := tview.Escape(user.Username)

	if user.Bot {
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
			return true
		}

		return firstRole.Position > secondRole.Position
	})
}
