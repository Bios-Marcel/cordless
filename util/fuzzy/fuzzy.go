package fuzzy

import (
	"sort"
	"strings"
	"unicode"

	"github.com/Bios-Marcel/discordgo"
)

type SearchResult struct {
	Key   string
	Value float64
}

func ScoreAndSortRoles(searchTerm string, searchables []*discordgo.Role) []*discordgo.Role {
	if len(searchTerm) == 0 {
		return searchables
	}

	searchableNameCache := make(map[*discordgo.Role]string)
	scoredItems := make(map[*discordgo.Role]float64, len(searchables))
	for _, searchable := range searchables {
		searchableValue := searchable.Name
		score := Score(searchTerm, searchableValue)
		if score > 0 {
			searchableNameCache[searchable] = searchableValue
			scoredItems[searchable] = score
		}
	}

	// Convert the results into an array.
	sortedItems := make([]*discordgo.Role, 0, len(scoredItems))
	for key := range scoredItems {
		sortedItems = append(sortedItems, key)
	}

	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(sortedItems, func(a, b int) bool {
		aItem := sortedItems[a]
		bItem := sortedItems[b]
		scoreA := scoredItems[aItem]
		scoreB := scoredItems[bItem]
		if scoreA == scoreB {
			return strings.Compare(strings.ToLower(searchableNameCache[aItem]), strings.ToLower(searchableNameCache[bItem])) < 0
		}
		return scoreA > scoreB
	})
	return sortedItems
}

func ScoreAndSortMembers(searchTerm string, searchables []*discordgo.Member) []*discordgo.Member {
	if len(searchTerm) == 0 {
		return searchables
	}

	searchableNameCache := make(map[*discordgo.Member]string)
	scoredItems := make(map[*discordgo.Member]float64, len(searchables))
	for _, searchable := range searchables {
		searchableValue := searchable.User.Username + "#" + searchable.User.Discriminator + " | " + searchable.Nick
		score := Score(searchTerm, searchableValue)
		if score > 0 {
			searchableNameCache[searchable] = searchableValue
			scoredItems[searchable] = score
		}
	}

	// Convert the results into an array.
	sortedItems := make([]*discordgo.Member, 0, len(scoredItems))
	for key := range scoredItems {
		sortedItems = append(sortedItems, key)
	}

	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(sortedItems, func(a, b int) bool {
		aItem := sortedItems[a]
		bItem := sortedItems[b]
		scoreA := scoredItems[aItem]
		scoreB := scoredItems[bItem]
		if scoreA == scoreB {
			return strings.Compare(strings.ToLower(searchableNameCache[aItem]), strings.ToLower(searchableNameCache[bItem])) < 0
		}
		return scoreA > scoreB
	})
	return sortedItems
}

func ScoreAndSortUsers(searchTerm string, searchables []*discordgo.User) []*discordgo.User {
	if len(searchTerm) == 0 {
		return searchables
	}

	searchableNameCache := make(map[*discordgo.User]string)
	scoredItems := make(map[*discordgo.User]float64, len(searchables))
	for _, searchable := range searchables {
		searchableValue := searchable.Username + "#" + searchable.Discriminator
		score := Score(searchTerm, searchableValue)
		if score > 0 {
			searchableNameCache[searchable] = searchableValue
			scoredItems[searchable] = score
		}
	}

	// Convert the results into an array.
	sortedItems := make([]*discordgo.User, 0, len(scoredItems))
	for key := range scoredItems {
		sortedItems = append(sortedItems, key)
	}

	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(sortedItems, func(a, b int) bool {
		aItem := sortedItems[a]
		bItem := sortedItems[b]
		scoreA := scoredItems[aItem]
		scoreB := scoredItems[bItem]
		if scoreA == scoreB {
			return strings.Compare(strings.ToLower(searchableNameCache[aItem]), strings.ToLower(searchableNameCache[bItem])) < 0
		}
		return scoreA > scoreB
	})
	return sortedItems
}

func ScoreAndSortEmoji(searchTerm string, unicodeEmoji []string, customEmoji []*discordgo.Emoji) []string {
	if len(searchTerm) == 0 {
		//For now, we just returned the normal ones in case no keyword was given
		return unicodeEmoji
	}

	scoredItems := make(map[string]float64, len(unicodeEmoji)+len(customEmoji))
	for _, searchable := range unicodeEmoji {
		score := Score(searchTerm, searchable)
		if score > 0 {
			scoredItems[searchable] = score
		}
	}
	for _, searchable := range customEmoji {
		score := Score(searchTerm, searchable.Name)
		if score > 0 {
			//Little hack to allow duplicate names
			scoredItems[" "+searchable.Name] = score
		}
	}

	// Convert the results into an array.
	sortedItems := make([]string, 0, len(scoredItems))
	for key := range scoredItems {
		sortedItems = append(sortedItems, key)
	}

	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(sortedItems, func(a, b int) bool {
		aItem := sortedItems[a]
		bItem := sortedItems[b]
		scoreA := scoredItems[aItem]
		scoreB := scoredItems[bItem]
		if scoreA == scoreB {
			return strings.Compare(strings.ToLower(aItem), strings.ToLower(bItem)) < 0
		}
		return scoreA > scoreB
	})
	return sortedItems
}

func ScoreAndSortChannels(searchTerm string, searchables []*discordgo.Channel) []*discordgo.Channel {
	if len(searchTerm) == 0 {
		return searchables
	}

	scoredItems := make(map[*discordgo.Channel]float64, len(searchables))
	for _, searchable := range searchables {
		score := Score(searchTerm, searchable.Name)
		if score > 0 {
			scoredItems[searchable] = score
		}
	}

	// Convert the results into an array.
	sortedItems := make([]*discordgo.Channel, 0, len(scoredItems))
	for key := range scoredItems {
		sortedItems = append(sortedItems, key)
	}

	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(sortedItems, func(a, b int) bool {
		aItem := sortedItems[a]
		bItem := sortedItems[b]
		scoreA := scoredItems[aItem]
		scoreB := scoredItems[bItem]
		if scoreA == scoreB {
			return strings.Compare(strings.ToLower(aItem.Name), strings.ToLower(bItem.Name)) < 0
		}
		return scoreA > scoreB
	})
	return sortedItems
}

func ScoreSearch(searchTerm string, searchItems []string) map[string]float64 {
	mymap := make(map[string]float64)
	for _, str := range searchItems {
		mymap[str] = Score(searchTerm, str)
	}
	return mymap
}

func SortSearchResults(results map[string]float64) []SearchResult {
	// Convert the results into an array.
	var arr []SearchResult
	for key, value := range results {
		if value >= 0.0 {
			arr = append(arr, SearchResult{key, value})
		}
	}
	// Sort results based on score. Equal scores are sorted alphabetically.
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Value == arr[j].Value {
			return strings.Compare(strings.ToLower(arr[i].Key), strings.ToLower(arr[j].Key)) < 0
		}
		return arr[i].Value > arr[j].Value
	})
	return arr
}

// Returns:
// -1 if the needle contains letters the haystack does not contain,
// or if the needle length exceeds the haystack length.
//
// 0 if no similarities were found
//
// > 0 based on similarities between the needle and haystack (increasing)
func Score(needle, haystack string) float64 {
	needleLength := len(needle)
	haystackLength := len(haystack)
	if needleLength > haystackLength {
		return -1
	}

	if needleLength == 0 {
		return 0
	}

	lowerNeedle := strings.ToLower(needle)
	lowerHaystack := strings.ToLower(haystack)
	score := 0.0
	for i, j := 0, 0; i < needleLength && j < haystackLength; i, j = i+1, j+1 {

		letter := lowerNeedle[i]
		letterIndex := strings.IndexByte(lowerHaystack[j:], letter) + j
		if (letterIndex - j) < 0 {
			return -1
		}

		if areLettersSameCase(needle[i], haystack[j]) {
			score += 0.5
		}

		if letterIndex == j {
			// Letter was consecutive
			score += 8
		} else {
			score += 1 - (0.1 * float64(letterIndex))
			// Move j up to the next found letter.
			j = letterIndex
		}

		if j == haystackLength-1 && i < needleLength-1 {
			return -1
		}
	}
	return score
}

func areLettersSameCase(letterA, letterB byte) bool {
	return unicode.IsUpper(rune(letterA)) && unicode.IsUpper(rune(letterB)) ||
		unicode.IsLower(rune(letterA)) && unicode.IsLower(rune(letterB))
}
