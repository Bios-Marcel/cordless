package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

// Theme is a wrapper around the tview.Theme. This wrapper can be extended with
// additional theming properties and the underlying tview.Theme can still be
// applied to tview.Styles
type Theme struct {
	*tview.Theme
}

var (
	theme = createDefaultTheme()
)

func createDefaultTheme() *Theme {
	return &Theme{&tview.Theme{
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
	}}
}

func GetThemeFile() (string, error) {
	configDir, configError := GetConfigDirectory()

	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDir, "theme.json"), nil
}

func LoadTheme() (*Theme, error) {
	themeFilePath, themeError := GetThemeFile()
	if themeError != nil {
		return nil, themeError
	}

	themeFile, openError := os.Open(themeFilePath)

	if os.IsNotExist(openError) {
		return theme, nil
	}

	if openError != nil {
		return nil, openError
	}

	defer themeFile.Close()
	decoder := json.NewDecoder(themeFile)
	themeLoadError := decoder.Decode(&theme)

	//io.EOF would mean empty, therefore we use defaults.
	if themeLoadError != nil && themeLoadError != io.EOF {
		return nil, themeLoadError
	}

	return theme, nil
}
