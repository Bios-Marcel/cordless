package tcellutil

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/ui/fdwscreen"
	"github.com/gdamore/tcell"
)

func NewScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("screen creation error: %v", err)
	}

	screen, err = fdwscreen.NewFdSwappingScreen(screen)
	if err != nil {
		return nil, fmt.Errorf("fd-wrapping screen creation error: %v", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("initialization error: %v", err)
	}

	return screen, nil
}
