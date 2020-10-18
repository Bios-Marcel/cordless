package readstate

import (
	"strconv"
	"sync"
	"time"

	"github.com/Bios-Marcel/discordgo"
)

var (
	data           = make(map[string]uint64)
	mentions       = make(map[string]bool)
	readStateMutex = &sync.Mutex{}
	timerMutex     = &sync.Mutex{}
	ackTimers      = make(map[string]*time.Timer)
	state          *discordgo.State
)

// Load loads the locally saved readmarkers returning an error if this failed.
func Load(sessionState *discordgo.State) {
	state = sessionState

	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	for _, channelState := range sessionState.ReadState {
		lastMessageID := channelState.GetLastMessageID()
		if lastMessageID == "" {
			continue
		}

		parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
		if parseError != nil {
			continue
		}

		data[channelState.ID] = parsed

		if channelState.MentionCount > 0 {
			mentions[channelState.ID] = true
		}
	}

}

// ClearReadStateFor clears all entries for the given Channel.
func ClearReadStateFor(channelID string) {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	timerMutex.Lock()
	defer timerMutex.Unlock()

	delete(data, channelID)
	delete(ackTimers, channelID)
}

// UpdateReadLocal can be used to locally update the data without sending
// anything to the Discord API. The update will only be applied if the new
// message ID is greater than the old one.
func UpdateReadLocal(channelID string, lastMessageID string) bool {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	delete(mentions, channelID)

	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return false
	}

	old, isPresent := data[channelID]
	if !isPresent || old < parsed {
		data[channelID] = parsed
		return true
	}

	return false
}

// UpdateRead tells the discord server that a channel has been read. If the
// channel has already been read and this method was called needlessly, then
// this will be a No-OP.
func UpdateRead(session *discordgo.Session, channel *discordgo.Channel, lastMessageID string) error {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	delete(mentions, channel.ID)

	// Avoid unnecessary traffic
	if hasBeenReadWithoutLocking(channel, lastMessageID) {
		return nil
	}

	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return parseError
	}

	data[channel.ID] = parsed

	_, ackError := session.ChannelMessageAck(channel.ID, lastMessageID, "")
	return ackError
}

// UpdateReadBuffered triggers an acknowledgement after a certain amount of
// seconds. If this message is called again during that time, the timer will
// be reset. This avoid unnecessarily many calls to the Discord servers.
func UpdateReadBuffered(session *discordgo.Session, channel *discordgo.Channel, lastMessageID string) {
	timerMutex.Lock()
	timerMutex.Unlock()

	ackTimer := ackTimers[channel.ID]
	if ackTimer == nil {
		newTimer := time.NewTimer(4 * time.Second)
		ackTimers[channel.ID] = newTimer
		go func() {
			<-newTimer.C
			ackTimers[channel.ID] = nil
			UpdateRead(session, channel, lastMessageID)
		}()
	} else {
		ackTimer.Reset(4 * time.Second)
	}
}

// IsGuildMuted returns whether the user muted the given guild.
func IsGuildMuted(guildID string) bool {
	for _, settings := range state.UserGuildSettings {
		if settings.GuildID == guildID {
			if settings.Muted && isStillMuted(settings.MuteConfig) {
				return true
			}

			break
		}
	}

	return false
}

// HasGuildBeenRead returns true if the guild has no unread messages or is
// muted.
func HasGuildBeenRead(guildID string) bool {
	if IsGuildMuted(guildID) {
		return true
	}

	realGuild, cacheError := state.Guild(guildID)
	if cacheError == nil {
		readStateMutex.Lock()
		defer readStateMutex.Unlock()

		for _, channel := range realGuild.Channels {
			if !hasReadMessagesPermission(channel.ID, state) {
				continue
			}

			if !hasBeenReadWithoutLocking(channel, channel.LastMessageID) {
				return false
			}
		}
	}

	return true
}

//HACK Had to copy this from discordutil/channel.go due to import cycle.
func hasReadMessagesPermission(channelID string, state *discordgo.State) bool {
	userPermissions, err := state.UserChannelPermissions(state.User.ID, channelID)
	if err != nil {
		// Unable to access channel permissions.
		return false
	}
	return (userPermissions & discordgo.PermissionViewChannel) > 0
}

// HasGuildBeenMentioned checks whether any channel in the guild mentioned
// the currently logged in user.
func HasGuildBeenMentioned(guildID string) bool {
	if IsGuildMuted(guildID) {
		return false
	}

	realGuild, cacheError := state.Guild(guildID)
	if cacheError == nil {
		readStateMutex.Lock()
		defer readStateMutex.Unlock()

		for _, channel := range realGuild.Channels {
			if hasBeenMentionedWithoutLocking(channel.ID) {
				return true
			}
		}
	}

	return false
}

func isStillMuted(config *discordgo.MuteConfig) bool {
	if config == nil || config.EndTime == "" {
		//This means permanently muted; I think!
		//We make the assumption that this function is only
		//called if "Muted" is set to "true". Therefore no timeframe means
		//we must be permanently muted.
		return true
	}

	muteEndTime, parseError := config.EndTime.Parse()
	if parseError != nil {
		panic(parseError)
	}

	return time.Now().UTC().Before(muteEndTime)
}

func isChannelMuted(channel *discordgo.Channel) bool {
	//optimization for the case of guild channels, as the handling for
	//private channels will be unnecessarily slower.
	if channel.GuildID == "" {
		return IsPrivateChannelMuted(channel)
	}

	return IsGuildChannelMuted(channel)
}

// IsGuildChannelMuted checks whether a guild channel has been set to silent.
func IsGuildChannelMuted(channel *discordgo.Channel) bool {
	if isGuildChannelMuted(channel.GuildID, channel.ID) {
		return true
	}

	//Check if Parent (CATEGORY) is muted
	if channel.ParentID != "" && isGuildChannelMuted(channel.GuildID, channel.ParentID) {
		return true
	}

	return false
}

func isGuildChannelMuted(guildID, channelID string) bool {
	for _, settings := range state.UserGuildSettings {
		if settings.GetGuildID() == guildID {
			for _, override := range settings.ChannelOverrides {
				if override.ChannelID == channelID {
					if override.Muted && isStillMuted(override.MuteConfig) {
						return true
					}

					break
				}
			}

			break
		}
	}

	return false
}

// IsPrivateChannelMuted checks whether a private channel has been set to
// silent.
func IsPrivateChannelMuted(channel *discordgo.Channel) bool {
	for _, settings := range state.UserGuildSettings {
		//Discord holds the mute settings for private channels in the user-guildsettings
		//but for an empty Guild ID. Doesn't really make sense, but ¯\_(ツ)_/¯
		if settings.GetGuildID() == "" {
			for _, override := range settings.ChannelOverrides {
				if override.ChannelID == channel.ID {
					if override.Muted && isStillMuted(override.MuteConfig) {
						return true
					}

					break
				}
			}

			//No break here, since it can happen that there are multiple
			//instances of UserGuildSettings for non guilds ... don't ask
			//me why ...
		}
	}

	return false
}

// HasBeenRead checks whether the passed channel has an unread Message or not.
func HasBeenRead(channel *discordgo.Channel, lastMessageID string) bool {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	return hasBeenReadWithoutLocking(channel, lastMessageID)
}

// HasBeenMentioned checks whether the currently logged in user has been
// mentioned in this channel.
func HasBeenMentioned(channelID string) bool {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	return hasBeenMentionedWithoutLocking(channelID)
}

func hasBeenMentionedWithoutLocking(channelID string) bool {
	mentioned, ok := mentions[channelID]
	return ok && mentioned
}

// MarkAsMentioned sets the given channel ID to mentioned.
func MarkAsMentioned(channelID string) {
	readStateMutex.Lock()
	defer readStateMutex.Unlock()

	mentions[channelID] = true
}

// hasBeenReadWithoutLocking checks whether the passed channel has an unread Message or not.
// The difference to HasBeenRead is, that no locking happens. This is inteded to be used
// for recursive calls to this method and avoiding lock overhead and deadlocks.
func hasBeenReadWithoutLocking(channel *discordgo.Channel, lastMessageID string) bool {
	if lastMessageID == "" {
		return true
	}

	if isChannelMuted(channel) {
		return true
	}

	// If there was no message, lastMessageID would've been empty, therefore
	// this check only makes sense if the cache is filled already.
	if len(channel.Messages) > 0 {
		lastMessage := channel.Messages[len(channel.Messages)-1]
		//I once had a crash here running into a nil-dereference, so I assume the author must've been null.
		if lastMessage != nil && lastMessage.Author != nil && lastMessage.Author.ID == state.User.ID {
			return true
		}
	}

	data, present := data[channel.ID]
	if !present {
		//We return true as there are too many false-positive otherwise and damn, that shit is annoying.
		return true
	}

	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return true
	}

	return data >= parsed
}
