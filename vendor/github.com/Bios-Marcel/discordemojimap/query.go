package discordemojimap

import "strings"

// ContainsEmoji returns true if that emoji is mapped to one or more key.
func ContainsEmoji(emoji string) bool {
	for _, emojiInMap := range EmojiMap {
		if emojiInMap == emoji {
			return true
		}
	}

	return false
}

// ContainsCode returns true if that emojicode is mapped to an emoji.
func ContainsCode(emojiCode string) bool {
	_, contains := EmojiMap[emojiCode]
	return contains
}

// GetEmojiCodes contains all codes for an emoji in an array. If no code could
// be found, then the resulting array will be empty.
func GetEmojiCodes(emoji string) []string {
	codes := make([]string, 0)
	for code, emojiInMap := range EmojiMap {
		if emojiInMap == emoji {
			codes = append(codes, code)
		}
	}

	return codes
}

// GetEmoji returns the matching emoji or an empty string in case no match was
// found for the given code.
func GetEmoji(emojiCode string) string {
	emoji, _ := EmojiMap[emojiCode]
	return emoji
}

// GetEntriesStartingWith returns all key-value pairs where the key(code)
// is prefixed with the given string. If no matches were found, the map is
// empty.
func GetEntriesStartingWith(startsWith string) map[string]string {
	matches := make(map[string]string)
	if len(startsWith) == 0 {
		return matches
	}

	searchTerm := strings.TrimSuffix(startsWith, ":")

	for emojiCode, emoji := range EmojiMap {
		if strings.HasPrefix(emojiCode, searchTerm) {
			matches[emojiCode] = emoji
		}
	}

	return matches
}
