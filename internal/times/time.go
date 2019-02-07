package times

import (
	"fmt"
	"time"

	"github.com/Bios-Marcel/cordless/internal/config"
)

func TimeToString(time *time.Time) string {
	localTime := time.Local()
	if config.GetConfig().Times == config.HourMinuteAndSeconds {
		return fmt.Sprintf("%02d:%02d:%02d", localTime.Hour(), localTime.Minute(), localTime.Second())
	} else if config.GetConfig().Times == config.HourAndMinute {
		return fmt.Sprintf("%02d:%02d", localTime.Hour(), localTime.Minute())
	}

	return ""
}
