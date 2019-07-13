package tviewutil

import "testing"

func TestCalculateNeccessaryHeight(t *testing.T) {
	TestHeight := func(lineWidth int, text string, expectedHeight int) {
		result := CalculateNeccessaryHeight(lineWidth, text)
		if result != expectedHeight {
			t.Errorf("Result was %d, but should've been %d", result, expectedHeight)
		}
	}

	TestHeight(20, "abcdeabcdeabcde abcd \n", 3)
	TestHeight(14, "1234567890123456", 2)
	TestHeight(10, "1\n23456 8901", 3)
}
