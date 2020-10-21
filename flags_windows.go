package main

import (
	"flag"
	"os"
	"strings"

	"github.com/Bios-Marcel/cordless/config"
)

func init() {
	wtSessionValue, avail := os.LookupEnv("WT_SESSION")
	disableUTF8Default = !avail || strings.TrimSpace(wtSessionValue) == ""
	flag.BoolVar(&config.DisableUTF8, "disable-UTF8", disableUTF8Default, "Replaces certain characters with question marks in order to avoid broken rendering")
}
