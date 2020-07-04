package tviewutil

import "strings"

// CalculateNecessaryHeight calculates the necessary height in the ui given
// the text and the width of the component the text will appear in.
func CalculateNecessaryHeight(width int, text string) int {
	splitLines := strings.Split(text, "\n")

	wrappedLines := 0
	for _, line := range splitLines {
		if len(line) >= width {
			wrappedLines = wrappedLines + ((len(line) - (len(line) % width)) / width)
		}
	}

	return len(splitLines) + wrappedLines

}
