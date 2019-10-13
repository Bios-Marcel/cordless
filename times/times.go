package times

import (
	"fmt"
	"time"

	"github.com/Bios-Marcel/cordless/config"
)

// TimeToLocalString formats a time to a string depending on the users settings.
// The time will first be converted into a local time.
func TimeToLocalString(time *time.Time) string {
	localTime := time.Local()
	return TimeToString(&localTime)
}

// TimeToString formats a time to a string depending on the users settings.
func TimeToString(time *time.Time) string {
	if config.GetConfig().Times == config.NoTime {
		return ""
	}
	if config.GetConfig().Times == config.HourMinuteAndSeconds {
		return fmt.Sprintf("%02d:%02d:%02d", time.Hour(), time.Minute(), time.Second())
	} else if config.GetConfig().Times == config.HourAndMinute {
		return fmt.Sprintf("%02d:%02d", time.Hour(), time.Minute())
	}
	return ""
}

// AreDatesTheSameDay returns true if the passed times represent the same day
// of the year.
func AreDatesTheSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}
