//+build go1.12

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Bios-Marcel/cordless/app"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/logging"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/ui/shortcutdialog"
	"github.com/Bios-Marcel/cordless/version"
	"github.com/Bios-Marcel/cordless/windowman"
)

func main() {
	showVersion := flag.Bool("version", false, "Show the version instead of starting cordless")
	showShortcutsDialog := flag.Bool("shortcut-dialog", false, "Shows the shortcuts dialog instead of launching cordless")
	setConfigDirectory := flag.String("config-dir", "", "Sets the configuration directory")
	setScriptDirectory := flag.String("script-dir", "", "Sets the script directory")
	setConfigFilePath := flag.String("config-file", "", "Sets exact path of the configuration file")
	accountToUse := flag.String("account", "", "Defines which account cordless tries to load")
	logPath := flag.String("log", "", "Defines what file we log to")
	uiPrototype := flag.Bool("ui-prototype", false, "Start with the prototype UI")
	flag.Parse()

	if logPath != nil && *logPath != "" {
		logFile, openError := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if openError != nil {
			panic(openError)
		}
		defer logFile.Close()
		logging.SetDefaultOutput(logFile)
	}

	if setConfigDirectory != nil {
		config.SetConfigDirectory(*setConfigDirectory)
	}
	if setScriptDirectory != nil {
		config.SetScriptDirectory(*setScriptDirectory)
	}
	if setConfigFilePath != nil {
		config.SetConfigFile(*setConfigFilePath)
	}

	themeLoadingError := config.LoadTheme()
	if themeLoadingError != nil {
		panic(themeLoadingError)
	}
	tview.Styles = *config.GetTheme().Theme

	if showShortcutsDialog != nil && *showShortcutsDialog {
		wm := windowman.GetWindowManager()
		shortcutWindow := shortcutdialog.NewShortcutWindow()
		wmError := wm.Run(shortcutWindow)
		if wmError != nil {
			panic(wmError)
		}
	} else if showVersion != nil && *showVersion {
		fmt.Printf("You are running cordless version %s\nKeep in mind that this version might not be correct for manually built versions, as those can contain additional commits.\n", version.Version)
	} else if uiPrototype != nil && *uiPrototype {
		fmt.Println("You've started up using the prototype UI. There's nothing to see here yet.")
	} else {
		if accountToUse != nil && *accountToUse != "" {
			app.RunWithAccount(*accountToUse)
		} else {
			app.Run()
		}
	}
}
