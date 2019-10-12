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

// CompareMessageDates returns false if the dates do not match and the t2 time
func CompareMessageDates(t1, t2 time.Time) bool {
	if t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay() {
		return true
	}
	return false
}
