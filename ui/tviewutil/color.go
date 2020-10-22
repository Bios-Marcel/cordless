package tviewutil

import (
	"fmt"

	tcell "github.com/gdamore/tcell/v2"
)

var (
	colorCache = make(map[tcell.Color]string)
)

// ColorToHex converts the tcell.Color to it's hexadecimal presentation
// and returns it as a string prepended with a # character.
func ColorToHex(color tcell.Color) string {
	value, found := colorCache[color]
	if found {
		return value
	}

	newValue := fmt.Sprintf("#%06x", color.Hex())
	colorCache[color] = newValue

	return newValue
}
