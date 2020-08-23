package text

import "unicode"

// FindEmojiIndices will find all parts of a string (rune array) that
// could potentially be emoji sequences and will return them backwards.
// they'll be returned backwards in order to allow easily manipulating the
// data without invalidating the following indexes accidentally.
// Example:
//     Hello :world:, what a :nice: day.
// would result in
//     []int{22,27,6,12}
func FindEmojiIndices(runes []rune) []int {
	var sequencesBackwards []int
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == ':' {
			for j := i - 1; j >= 0; j-- {
				char := runes[j]
				if char == ':' {
					//Handle this ':' in the next iteration of the outer loop otherwise
					if j != i-1 {
						sequencesBackwards = append(sequencesBackwards, j, i)
						i = j
						break
					} else {
						break
					}
				} else if unicode.IsSpace(char) {
					i = j
					break
				}
			}
		}
	}

	return sequencesBackwards
}
