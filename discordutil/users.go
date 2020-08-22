package discordutil

import (
	"math/rand"
	"sort"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"

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

// GetMemberColor gets the members color according the members role.
// If role colors aren't enabled, or the member is a bot, we fallthrough
// to GetUserColor.
func GetMemberColor(state *discordgo.State, member *discordgo.Member) string {
	if len(member.Roles) > 0 && !member.User.Bot &&
		config.Current.UserColors == config.RoleColor {

		//Since roles aren't sorted, we need to find the smallest one.
		var roleToUse *discordgo.Role
		for _, memberRole := range member.Roles {
			role, _ := state.Role(member.GuildID, memberRole)
			if roleToUse == nil || role.Position > roleToUse.Position {
				roleToUse = role
			}
		}

		if roleToUse != nil {
			if color := GetRoleColor(roleToUse); color != "" {
				return color
			}
		}
	}

	return GetUserColor(member.User)
}

// GetUserColor gers a user color according to the configuration.
// If "random" is the setting, then a new random color is retrieved
// and cached for this session and this user.
func GetUserColor(user *discordgo.User) string {
	//Despite user settings, bots always get a color.
	if user.Bot {
		return tviewutil.ColorToHex(config.GetTheme().BotColor)
	}

	switch config.Current.UserColors {
	case config.RandomColor:
		//Avoid unnecessarily retrieving and caching colors and fallthrough
		//to using single color instead.
		if randColorLength != 0 {
			color, ok := userColorCache[user.ID]
			if ok {
				return color
			}

			newColor := getRandomColorString()
			userColorCache[user.ID] = newColor
			return newColor
		}

		fallthrough
	case config.SingleColor:
		return tviewutil.ColorToHex(config.GetTheme().DefaultUserColor)
	case config.NoColor:
		fallthrough
	default:
		return tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)
	}
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
	if config.Current.ShowNicknames && member.Nick != "" {
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
