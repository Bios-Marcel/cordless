package version

import (
	"context"
	"fmt"

	"github.com/google/go-github/v29/github"
)

var latestRemoteVersion string

// CheckForUpdate checks whether the Version-string saved at version.Version
// is different to the name of the latest tagged GitHub release. If it is, and
// the release name isn't equal to the `donRemindFor`-parameter, we return
// true, which stands for "update available". This allows the user to manually
// say "Don't remind me again for this version". The actual return value is
// supplied via a channel that's closed after one value has been read.
func CheckForUpdate(dontRemindFor string) chan bool {
	//Note, this isn't buffered, so that we can safely close the channel
	//when it has been read by the caller.
	updateAvailableChannel := make(chan bool)
	go func() {
		remoteVersion := GetLatestRemoteVersion()
		if remoteVersion == "" || remoteVersion == dontRemindFor {
			//Error retrieving version or user wishes to ignore update.
			updateAvailableChannel <- false
		} else {
			updateAvailableChannel <- isLocalOlderThanRemote(Version, remoteVersion)
		}
		close(updateAvailableChannel)
	}()
	return updateAvailableChannel
}

func isLocalOlderThanRemote(local, remote string) bool {
	yearRemote, monthRemote, dayRemote := parseTag(remote)
	yearLocal, monthLocal, dayLocal := parseTag(local)
	return !(yearLocal >= yearRemote && monthLocal >= monthRemote && dayLocal >= dayRemote)
}

func parseTag(tag string) (year int, month int, day int) {
	fmt.Sscanf(tag, "%04d-%02d-%02d", &year, &month, &day)
	return
}

// GetLatestRemoteVersion queries GitHub for the latest Release-Tag and caches
// it. This value will never be updated during runtime.
func GetLatestRemoteVersion() string {
	if latestRemoteVersion != "" {
		return latestRemoteVersion
	}

	repositoryRelease, _, lookupError := github.NewClient(nil).Repositories.GetLatestRelease(context.Background(), "Bios-Marcel", "cordless")
	if lookupError != nil || repositoryRelease == nil {
		return ""
	}

	latestRemoteVersion = *repositoryRelease.TagName
	return latestRemoteVersion
}
