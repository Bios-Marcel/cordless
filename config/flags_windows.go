package config

import (
	"os"
	"strings"
)

// DisableUTF8 set to true will cause cordless to replace characters with a
// codepoint higher than 65536 or a runewidth of more than one character.
var DisableUTF8 bool

func init() {
	wtSessionValue, avail := os.LookupEnv("WT_SESSION")
	DisableUTF8 = !avail || strings.TrimSpace(wtSessionValue) == ""
}
