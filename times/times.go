package times

import (
	"fmt"
	"time"

	"github.com/Bios-Marcel/cordless/config"
)

// TimeToString formats a time to a string depending on the users settings.
// The time will first be converted into a local time.
func TimeToString(time *time.Time) string {
	if config.GetConfig().Times == config.NoTime {
		return ""
	}

	localTime := time.Local()
	if config.GetConfig().Times == config.HourMinuteAndSeconds {
		return fmt.Sprintf("%02d:%02d:%02d", localTime.Hour(), localTime.Minute(), localTime.Second())
	} else if config.GetConfig().Times == config.HourAndMinute {
		return fmt.Sprintf("%02d:%02d", localTime.Hour(), localTime.Minute())
	}

	return ""
}
