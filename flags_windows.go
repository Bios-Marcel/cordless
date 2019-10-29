package main

import (
	"flag"
	"github.com/Bios-Marcel/cordless/config"
)

func init() {
	flag.BoolVar(&config.DisableUTF8, "disable-UTF8", true, "Replaces certain characters with question marks in order to avoid broken rendering")
}
