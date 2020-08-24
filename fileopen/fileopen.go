package fileopen

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/util/files"
	"github.com/skratchdot/open-golang/open"
)

var cacheCleanerLock = &sync.Mutex{}

// LaunchCacheCleaner clears all files in the given folder structure that are
// older than the given timeframe. Root-Paths are ignored for safety reasons.
func LaunchCacheCleaner(targetFolder string, olderThan time.Duration) {
	//We try to avoid deleting someones whole hard-drive-content.
	//Is there a better way to do this?
	if targetFolder == "" || targetFolder == "/" || (len(targetFolder) == 3 && strings.HasSuffix(targetFolder, ":/")) {
		return
	}

	go func() {
		cacheCleanerLock.Lock()
		defer cacheCleanerLock.Unlock()
		now := time.Now().UnixNano()
		filepath.Walk(targetFolder, func(path string, f os.FileInfo, err error) error {
			if now-f.ModTime().UnixNano() >= olderThan.Nanoseconds() {
				removeError := os.Remove(path)
				if removeError != nil {
					log.Printf("Couldn't remove file %s from cache.\n", removeError)
				}
			}

			return nil
		})
	}()
}

// OpenFile attempts downloading and opening a file from the given link.
// Files are either cached locally or saved permanently. In both cases
// cordless attempts loading the previously downloaded version of the
// file. Files are saved using the ID of the attachment in order to avoid
// false positives when doing cache-matching.
func OpenFile(targetFolder, fileID, downloadURL string) error {
	extension := strings.TrimPrefix(filepath.Ext(downloadURL), ".")
	targetFile := filepath.Join(targetFolder, fileID+"."+extension)
	downloadError := files.DownloadFileOrAccessCache(targetFile, downloadURL)
	if downloadError != nil {
		return downloadError
	}

	handler, handlerSet := config.Current.FileOpenHandlers[extension]
	if handlerSet {
		handlerTrimmed := strings.TrimSpace(handler)
		//Empty means to not open files with the given extension.
		if handlerTrimmed == "" {
			log.Printf("skip opening link %s, as the extension %s has been disabled.\n", downloadURL, extension)
			return nil
		}

		return openWithHandler(handlerTrimmed, targetFile)
	}

	defaultHandler, isDefaultHandlerSet := config.Current.FileOpenHandlers["*"]
	if isDefaultHandlerSet {
		defaultHandlerTrimmed := strings.TrimSpace(defaultHandler)
		if defaultHandlerTrimmed != "" {
			return openWithHandler(defaultHandlerTrimmed, targetFile)
		}
	}

	return open.Run(targetFile)
}

func openWithHandler(handler, targetFile string) error {
	commandParts := commands.ParseCommand(strings.ReplaceAll(handler, "{$file}", targetFile))
	command := exec.Command(commandParts[0], commandParts[1:]...)
	return command.Start()
}
