package fuzzy

import (
	"sort"
	"strings"
)

type SearchResult struct {
	Key   string
	Value float64
}

func ScoreSearch(searchTerm string, searchItems []string) map[string]float64 {
	mymap := make(map[string]float64)
	// Ignore case if there are no capital letters in the search term.
	caseInsensitive := searchTerm == strings.ToLower(searchTerm)
	if caseInsensitive {
		searchTerm = strings.ToLower(searchTerm)
	}
	for _, str := range searchItems {
		if caseInsensitive {
			mymap[str] = Score(searchTerm, strings.ToLower(str))
		} else {
			mymap[str] = Score(searchTerm, str)
		}
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

	score := 0.0
	needleIndex := 0
	for haystackIndex := 0; needleIndex < needleLength; haystackIndex++ {

		letterIndex := findLetterIndex(needle[needleIndex], haystack, haystackIndex)
		if letterIndex < 0 {
			return -1
		}

		if letterIndex == haystackIndex {
			// Letter was consecutive
			score += 8
		} else {
			score += 1 - (0.1 * float64(letterIndex))
			// Move haystackIndex up to the next found letter.
			haystackIndex = letterIndex
		}

		needleIndex++
	}
	return score
}

// Scores a letter from inside a string based on its distance from the start of the string.
// The index at which the letter was found will be returned.
// The score and index will be -1 if the character is not found.
func findLetterIndex(c byte, haystack string, startIndex int) int {
	for i := startIndex; i < len(haystack); i++ {
		if c == haystack[i] {
			return i
		}
	}
	// Letter not found.
	return -1
}
