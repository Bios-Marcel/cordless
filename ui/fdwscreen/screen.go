package fdwscreen

import (
	"github.com/gdamore/tcell"
)

// FdSwappingScreen wraps tcell.Screen with a fd-swapping mechanism.
type FdSwappingScreen struct {
	tcell.Screen
	swapper *fdSwapper
}

var _ tcell.Screen = (*FdSwappingScreen)(nil)

// Init initializes the screen for use.
func (s *FdSwappingScreen) Init() error {
	if s.swapper != nil {
		s.swapper.InitSwap()
	}
	return s.Screen.Init()
}

// Fini finalizes the screen also releasing resources.
func (s *FdSwappingScreen) Fini() {
	s.Screen.Fini()
	if s.swapper != nil {
		s.swapper.FiniSwap()
	}
}

// NewFdSwappingScreen wraps tcell.Screen with a fd-swapping mechanism.
func NewFdSwappingScreen(s tcell.Screen) (*FdSwappingScreen, error) {
	swapper, err := newFdSwapper()
	if err != nil {
		return nil, err
	}
	return &FdSwappingScreen{
		Screen:  s,
		swapper: swapper,
	}, nil
}
