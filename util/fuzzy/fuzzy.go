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
	searchTerm = strings.ToLower(searchTerm)
	for _, str := range searchItems {
		mymap[str] = Score(searchTerm, strings.ToLower(str))
	}
	return mymap
}

func SortSearchResults(results map[string]float64) []SearchResult {
	var arr []SearchResult
	for key, value := range results {
		if value > 0.0 {
			arr = append(arr, SearchResult{key, value})
		}
	}

	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Value == arr[j].Value {
			return strings.Compare(arr[i].Key, arr[j].Key) > 0
		}
		return arr[i].Value > arr[j].Value
	})

	return arr
}

func Score(needle, haystack string) float64 {
	needleLength := len(needle)
	haystackLength := len(haystack)
	if needleLength > haystackLength || needleLength == 0 {
		return 0
	}

	score := 0.0
	needleIndex := 0
	for haystackIndex := 0; haystackIndex < haystackLength && needleIndex < needleLength; haystackIndex++ {

		letterScore, foundAtIndex := scoreLetter(needle[needleIndex], haystack, haystackIndex)
		if letterScore == 0.0 {
			return 0.0
		}

		if foundAtIndex == haystackIndex {
			// Letter was consecutive
			score += letterScore * 2
		} else {
			score += letterScore
			// Move haystackIndex up to the next found letter.
			haystackIndex = foundAtIndex
		}

		needleIndex++
	}
	return score
}

// Scores a letter from inside a string based on its distance from the start of the string.
// The index at which the letter was found will be returned.
// The score will be 0 and the index -1 if the character is not found.
func scoreLetter(c byte, haystack string, startIndex int) (float64, int) {
	haystackLength := len(haystack)
	for i := startIndex; i < len(haystack); i++ {
		if c == haystack[i] {
			var displacement float64 = float64(haystackLength - i)
			score := displacement / float64(haystackLength) * (1.0 / float64(haystackLength))
			return score, i
		}
	}
	// Letter not found.
	return 0, -1
}
