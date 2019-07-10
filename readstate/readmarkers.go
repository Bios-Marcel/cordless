package readstate

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/discordgo"
)

var (
	readMarkers         = &ReadMarker{Data: make(map[string]uint64)}
	readmarkersFilePath string
	flushWaitGroup      = &sync.WaitGroup{}
)

// ReadMarker contains the data that matches channel ids to their last read
// message.
type ReadMarker struct {
	Data map[string]uint64
}

func init() {
	configDir, configDirErr := config.GetConfigDirectory()
	if configDirErr == nil {
		readmarkersFilePath = filepath.Join(configDir, "readmarkers.json")
	}

	go func(waitGroup *sync.WaitGroup) {
		for {
			waitGroup.Add(75)
			waitGroup.Wait()
			Flush()
		}
	}(flushWaitGroup)
}

// Load loads the locally saved readmarkers returing an error if this failed.
func Load() error {
	if readmarkersFilePath == "" {
		return errors.New("error loading data, filepath empty")
	}

	readmarkersFile, openError := os.Open(readmarkersFilePath)

	if os.IsNotExist(openError) {
		return nil
	}

	if openError != nil {
		return openError
	}

	defer readmarkersFile.Close()
	decoder := json.NewDecoder(readmarkersFile)
	decodeError := decoder.Decode(readMarkers)

	//io.EOF would mean empty, therefore we use defaults.
	if decodeError != nil && decodeError != io.EOF {
		return decodeError
	}

	return nil
}

// UpdateRead updates the local data for the passed channel using the passed
// Message ID. It also calls Done on the waitgroup for flushing.
func UpdateRead(channelID string, lastMessageID string) error {
	parsed, parseError := strconv.ParseUint(lastMessageID, 10, 64)
	if parseError != nil {
		return parseError
	}

	readMarkers.Data[channelID] = parsed

	flushWaitGroup.Done()

	return nil
}

// HasBeenRead checks whether the passed channel has an unread Message or not.
func HasBeenRead(channel *discordgo.Channel) bool {
	if channel.LastMessageID == "" {
		return true
	}

	data, present := readMarkers.Data[channel.ID]
	if !present {
		return false
	}

	parsed, parseError := strconv.ParseUint(channel.LastMessageID, 10, 64)
	if parseError != nil {
		return true
	}

	return data >= parsed
}

// Flush saves all local data to the harddrive.
func Flush() error {
	dataAsJSON, jsonError := json.MarshalIndent(readMarkers, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	return ioutil.WriteFile(readmarkersFilePath, dataAsJSON, 0666)
}
