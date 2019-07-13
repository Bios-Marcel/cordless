package tviewutil

import "strings"

// CalculateNeccessaryHeight calculates the necessary height in the ui given
// the text and the width of the component the text will appear in.
func CalculateNeccessaryHeight(lineWidth int, text string) int {
	splitLines := strings.Split(text, "\n")

	lineCount := 0
	for _, line := range splitLines {
		lineCount += CountNumberOfWrappedLines(lineWidth, line)
	}
	return lineCount + len(splitLines)
}

func CountNumberOfWrappedLines(lineWidth int, text string) int {
	wrappedLineCount := 0
	for len(text) >= lineWidth {
		wrappedLineCount++
		text = RemoveWrappedWords(lineWidth, text)
	}
	return wrappedLineCount
}

func RemoveWrappedWords(lineWidth int, text string) string {
	words := strings.Split(text, " ")
	numWordsToWrap := 0
	for len(strings.Join(words[:numWordsToWrap], " ")) < lineWidth {
		numWordsToWrap++
	}

	if numWordsToWrap == 1 {
		return RemoveWrappedChars(lineWidth, text)
	}
	return strings.Join(words[numWordsToWrap-1:], " ")
}
func RemoveWrappedChars(lineWidth int, text string) string {
	for len(text[:lineWidth]) < len(text) {
		lineWidth++
	}
	return text[lineWidth-1:]
}
