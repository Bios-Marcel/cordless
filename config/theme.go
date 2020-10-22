package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

// Theme is a wrapper around the tview.Theme. This wrapper can be extended with
// additional theming properties and the underlying tview.Theme can still be
// applied to tview.Styles
type Theme struct {
	*tview.Theme
	BlockedUserColor tcell.Color
	InfoMessageColor tcell.Color
	BotColor         tcell.Color
	MessageTimeColor tcell.Color
	DefaultUserColor tcell.Color
	LinkColor        tcell.Color
	AttentionColor   tcell.Color
	ErrorColor       tcell.Color
	RandomUserColors []tcell.Color
}

var (
	theme = createDefaultTheme()
)

// GetTheme returns a pointer to the currently loaded Theme.
func GetTheme() *Theme {
	return theme
}

func createDefaultTheme() *Theme {
	return &Theme{
		Theme: &tview.Theme{
			PrimitiveBackgroundColor:    tcell.ColorBlack,
			ContrastBackgroundColor:     tcell.ColorBlue,
			MoreContrastBackgroundColor: tcell.ColorGreen,
			BorderColor:                 tcell.ColorWhite,
			BorderFocusColor:            tcell.ColorBlue,
			TitleColor:                  tcell.ColorWhite,
			GraphicsColor:               tcell.ColorWhite,
			PrimaryTextColor:            tcell.ColorWhite,
			SecondaryTextColor:          tcell.ColorYellow,
			TertiaryTextColor:           tcell.ColorGreen,
			InverseTextColor:            tcell.ColorBlue,
			ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
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
}

// GetThemeFile returns the path to the theme file.
func GetThemeFile() (string, error) {
	configDir, configError := GetConfigDirectory()

	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDir, "theme.json"), nil
}

// LoadTheme reads the theme from the users configuration folder and stored it
// in the local state. It can be retrieved via GetTheme.
func LoadTheme() error {
	themeFilePath, themeError := GetThemeFile()
	if themeError != nil {
		return themeError
	}

	themeFile, openError := os.Open(themeFilePath)

	if os.IsNotExist(openError) {
		return nil
	}

	if openError != nil {
		return openError
	}

	defer themeFile.Close()
	decoder := json.NewDecoder(themeFile)
	themeLoadError := decoder.Decode(&theme)

	//io.EOF would mean empty, therefore we use defaults.
	if themeLoadError != nil && themeLoadError != io.EOF {
		return themeLoadError
	}

	return nil
}
