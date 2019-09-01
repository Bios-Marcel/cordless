package tviewutil

import (
	"fmt"

	"github.com/gdamore/tcell"
)

func ColorToHex(color tcell.Color) string {
	return fmt.Sprintf("#%06x", color.Hex())
}
