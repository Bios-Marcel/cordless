package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"

	"github.com/Bios-Marcel/cordless/config"
)

func main() {
	theme := &config.Theme{
		Theme: &tview.Theme{
			PrimitiveBackgroundColor:    tcell.NewRGBColor(70, 70, 70),
			ContrastBackgroundColor:     tcell.NewRGBColor(104, 142, 196),
			MoreContrastBackgroundColor: tcell.NewRGBColor(79, 79, 79),
			BorderColor:                 tcell.NewRGBColor(213, 220, 229),
			BorderFocusColor:            tcell.NewRGBColor(104, 142, 196),
			TitleColor:                  tcell.ColorWhite,
			GraphicsColor:               tcell.ColorWhite,
			PrimaryTextColor:            tcell.ColorWhite,
			SecondaryTextColor:          tcell.ColorWhite,
			TertiaryTextColor:           tcell.ColorWhite,
			InverseTextColor:            tcell.NewRGBColor(104, 142, 196),
			ContrastSecondaryTextColor:  tcell.NewRGBColor(104, 142, 196),
		},
		BlockedUserColor: tcell.ColorGray,
		InfoMessageColor: tcell.ColorGray,
		BotColor:         tcell.NewRGBColor(0x94, 0x96, 0xfc),
		MessageTimeColor: tcell.ColorGray,
		LinkColor:        tcell.ColorDarkCyan,
		DefaultUserColor: tcell.NewRGBColor(0x44, 0xe5, 0x44),
		AttentionColor:   tcell.ColorOrange,
		ErrorColor:       tcell.ColorRed,
		RandomUserColors: []tcell.Color{
			tcell.NewRGBColor(0xd8, 0x50, 0x4e),
			tcell.NewRGBColor(0xd8, 0x7e, 0x4e),
			tcell.NewRGBColor(0xd8, 0xa5, 0x4e),
			tcell.NewRGBColor(0xd8, 0xc6, 0x4e),
			tcell.NewRGBColor(0xb8, 0xd8, 0x4e),
			tcell.NewRGBColor(0x91, 0xd8, 0x4e),
			tcell.NewRGBColor(0x67, 0xd8, 0x4e),
			tcell.NewRGBColor(0x4e, 0xd8, 0x7c),
			tcell.NewRGBColor(0x4e, 0xd8, 0xaa),
			tcell.NewRGBColor(0x4e, 0xd8, 0xcf),
			tcell.NewRGBColor(0x4e, 0xb6, 0xd8),
			tcell.NewRGBColor(0x4e, 0x57, 0xd8),
			tcell.NewRGBColor(0x75, 0x4e, 0xd8),
			tcell.NewRGBColor(0xa3, 0x4e, 0xd8),
			tcell.NewRGBColor(0xcf, 0x4e, 0xd8),
			tcell.NewRGBColor(0xd8, 0x4e, 0x9c),
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	encoder.Encode(theme)
}

// Usage: fromHex("#FF0000")
func fromHex(hexString string) tcell.Color {
	trimmed := strings.TrimPrefix(strings.TrimSpace(hexString), "#")
	var r, g, b int32
	fmt.Sscanf(trimmed, "%02x%02x%02x", &r, &g, &b)
	return tcell.NewRGBColor(r, g, b)
}
