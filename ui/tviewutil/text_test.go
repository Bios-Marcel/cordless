package tviewutil

import "testing"

func TestCalculateNecessaryHeight(t *testing.T) {
	var necessaryHeight int
	necessaryHeight = CalculateNecessaryHeight(20, "abcdeabcdeabcde abcd \n")
	if necessaryHeight != 3 {
		t.Errorf("Result was %d, but should've been 3", necessaryHeight)
	}
}
