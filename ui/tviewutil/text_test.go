package tviewutil

import "testing"

func TestCalculateNeccessaryHeight(t *testing.T) {
	var neccessaryHeight int
	neccessaryHeight = CalculateNeccessaryHeight(20, "abcdeabcdeabcde abcd \n")
	if neccessaryHeight != 3 {
		t.Errorf("Result was %d, but should've been 3", neccessaryHeight)
	}
}
