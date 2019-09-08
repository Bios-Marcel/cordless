package fuzzy

import (
	"sort"
	"strings"
	"unicode"
)

type SearchResult struct {
	Key   string
	Value float64
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
	// Sort results based on score. Equal scores are sorted alphbetically.
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
