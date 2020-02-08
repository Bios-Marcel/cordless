package version

import (
	"context"
	"fmt"
	"github.com/google/go-github/v29/github"
)

var latestRemoteVersion *string

// IsLocalOutdated checks the latest release on github and compares the version
// numbers. Version numbers representing a later date than the local version
// will cause the call to return true. If an error occurs, we will return
// false. If dontRemindFor is passed, false will be returned if it equals the
// current remote version. This is used in order to avoid spamming the user
// with an update notification on every startup.
func IsLocalOutdated(dontRemindFor string) bool {
	if GetLatestRemoteVersion() == "" || GetLatestRemoteVersion() == dontRemindFor {
		return false
	}

	return isCurrentOlder(Version, GetLatestRemoteVersion())
}

func isCurrentOlder(current, other string) bool {
	var yearRemote, monthRemote, dayRemote int
	fmt.Sscanf(other, "%04d-%02d-%02d", &yearRemote, &monthRemote, &dayRemote)
	var yearLocal, monthLocal, dayLocal int
	fmt.Sscanf(current, "%04d-%02d-%02d", &yearLocal, &monthLocal, &dayLocal)
	return !(yearLocal >= yearRemote && monthLocal >= monthRemote && dayLocal >= dayRemote)
}

// GetLatestRemoteVersion queries GitHub for the latest Release-Tag and caches
// it. This value will never be updated during runtime of cordless.
func GetLatestRemoteVersion() string {
	if latestRemoteVersion != nil {
		return *latestRemoteVersion
	}

	repositoryRelease, _, lookupError := github.NewClient(nil).Repositories.GetLatestRelease(context.Background(), "Bios-Marcel", "cordless")
	if lookupError == nil {
		latestRemoteVersion = repositoryRelease.TagName
	} else {
		emptyVersion := ""
		latestRemoteVersion = &emptyVersion
	}

	return *latestRemoteVersion
}
