package readstate

import (
	"strconv"
	"sync"
	"time"

	"github.com/Bios-Marcel/discordgo"
)

var (
	data       = make(map[string]uint64)
	timerMutex = &sync.Mutex{}
	ackTimers  = make(map[string]*time.Timer)
)

// Load loads the locally saved readmarkers returing an error if this failed.
func Load(readState []*discordgo.ReadState) {
	for _, channelState := range readState {
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
func UpdateRead(session *discordgo.Session, channelID string, lastMessageID string) error {
	// Avoid unnecessary traffic
	if HasBeenRead(channelID, lastMessageID) {
		return nil
	}

	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return parseError
	}

	data[channelID] = parsed

	_, ackError := session.ChannelMessageAck(channelID, lastMessageID, "")
	return ackError
}

// UpdateReadBuffered triggers an acknowledgement after a certain amount of
// seconds. If this message is called again during that time, the timer will
// be reset. This avoid unnecessarily many calls to the Discord servers.
func UpdateReadBuffered(session *discordgo.Session, channelID string, lastMessageID string) {
	timerMutex.Lock()
	ackTimer := ackTimers[channelID]
	if ackTimer == nil {
		newTimer := time.NewTimer(4 * time.Second)
		ackTimers[channelID] = newTimer
		go func() {
			<-newTimer.C
			ackTimers[channelID] = nil
			UpdateRead(session, channelID, lastMessageID)
		}()
	} else {
		ackTimer.Reset(4 * time.Second)
	}
	timerMutex.Unlock()
}

// HasBeenRead checks whether the passed channel has an unread Message or not.
func HasBeenRead(channelID string, lastMessageID string) bool {
	if lastMessageID == "" {
		return true
	}

	data, present := data[channelID]
	if !present {
		return false
	}

	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return true
	}

	return data >= parsed
}
