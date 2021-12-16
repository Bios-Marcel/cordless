// Package discordemojimap provides a Replace function in order to escape
// emoji sequences with their respective emojis.
package discordemojimap

import (
	"regexp"
	"strings"
)

var emojiCodeRegex = regexp.MustCompile("(?s):[a-zA-Z0-9_]+:")

// Replace all emoji sequences contained in the discord emoji map with their
// respective emojis.
//
// Examples for valid input:
//     Replace("Hello World :sun_with_face:")
// would result in
//     "Hello World 🌞"
func Replace(input string) string {
	if len(input) <= 2 {
		return input
	}

	replacedEmojis := emojiCodeRegex.ReplaceAllStringFunc(input, func(match string) string {
		emojified, contains := EmojiMap[strings.ToLower(match[1:len(match)-1])]
		if !contains {
			return match
		}

		return emojified
	})

	return replacedEmojis
}
