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
		}}
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
