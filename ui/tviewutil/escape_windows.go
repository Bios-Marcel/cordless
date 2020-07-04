package tviewutil

import (
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/mattn/go-runewidth"
)

// Escape delegates to tview escape, optionally doing additional escaping.
func Escape(text string) string {
	if config.DisableUTF8 {
		runes := []rune(text)
		for index, r := range runes {
			if r > 65536 || runewidth.RuneWidth(r) > 1 {
				runes[index] = '?'
			}
		}

		return tview.Escape(string(runes))
	}

	return tview.Escape(text)
}
