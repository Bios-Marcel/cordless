package readstate

import (
	"strconv"
	"sync"
	"time"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/discordutil"
)

var (
	data       = make(map[string]uint64)
	timerMutex = &sync.Mutex{}
	ackTimers  = make(map[string]*time.Timer)
	state      *discordgo.State
)

// Load loads the locally saved readmarkers returning an error if this failed.
func Load(sessionState *discordgo.State) {
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
	}

	state = sessionState
}

// ClearReadStateFor clears all entries for the given Channel.
func ClearReadStateFor(channelID string) {
	timerMutex.Lock()
	delete(data, channelID)
	delete(ackTimers, channelID)
	timerMutex.Unlock()
}

// UpdateReadLocal can be used to locally update the data without sending
// anything to the Discord API. The update will only be applied if the new
// message ID is greater than the old one.
func UpdateReadLocal(channelID string, lastMessageID string) bool {
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
	// Avoid unnecessary traffic
	if HasBeenRead(channel, lastMessageID) {
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
	timerMutex.Unlock()
}

// IsGuildMuted returns whether the user muted the given guild.
func IsGuildMuted(guildID string) bool {
	for _, settings := range state.UserGuildSettings {
		if settings.GuildID == guildID {
			if settings.Muted {
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
		for _, channel := range realGuild.Channels {
			if !discordutil.HasReadMessagesPermission(channel.ID, state) {
				continue
			}

			if !HasBeenRead(channel, channel.LastMessageID) {
				return false
			}
		}
	}

	return true
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
	for _, settings := range state.UserGuildSettings {
		if settings.GetGuildID() == channel.GuildID {
			for _, override := range settings.ChannelOverrides {
				if override.ChannelID == channel.ID {
					if override.Muted {
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
					if override.Muted {
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
