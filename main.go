//+build go1.12

package main

import (
	"flag"
	"fmt"
	"github.com/Bios-Marcel/cordless/app"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/version"
)

func main() {
	showVersion := flag.Bool("version", false, "Show the version instead of starting cordless")
	showShortcutsDialog := flag.Bool("shortcut-dialog", false, "Shows the shortcuts dialog instead of launching cordless")
	setConfigDirectory := flag.String("config-dir", "", "Set cordless' application directory")
	setConfigFile := flag.String("config-file", "", "Sets the exact location of the configuration file")
	flag.Parse()

	if showShortcutsDialog != nil && *showShortcutsDialog {
		shortcuts.RunShortcutsDialogStandalone()
	} else if showVersion != nil && *showVersion {
		fmt.Printf("You are running cordless version %s\nKeep in mind that this version might not be correct for manually built versions, as those can contain additional commits.\n", version.Version)
	} else {
		if setConfigDirectory != nil {
			config.SetConfigDirectory(*setConfigDirectory)
		}
		if setConfigFile != nil {
			config.SetConfigFile(*setConfigFile)
		}
		app.Run()
	}
}
