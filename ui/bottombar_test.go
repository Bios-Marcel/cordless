package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell"
)

func TestBottomBar(t *testing.T) {
	simScreen := tcell.NewSimulationScreen("UTF-8")

	simScreen.Init()
	simScreen.SetSize(10, 10)
	width, height := simScreen.Size()

	bottomBar := NewBottomBar()
	bottomBar.SetRect(0, 0, width, 1)
	bottomBar.Draw(simScreen)

	//If no items were added, we don't expect anything to be drawn.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			expectCell(' ', x, y, simScreen, t)
		}
	}

	bottomBar.AddItem(strings.Repeat("a", width))
	bottomBar.Draw(simScreen)

	//The first row should be filed with As as we have only one item with
	//the width of the screen.
	for x := 0; x < width; x++ {
		expectCell('a', x, 0, simScreen, t)
	}

	//We expect everything except for the first row to be empty
	for x := 0; x < width; x++ {
		for y := 1; y < height; y++ {
			expectCell(' ', x, y, simScreen, t)
		}
	}

	//Increase size as we add more items
	simScreen.SetSize(20, 10)
	width, height = simScreen.Size()
	bottomBar.SetRect(0, 0, width, 1)

	//Technically we need many more cells for this, which we don't have.
	//Testing this makes sure we don't crash.
	bottomBar.AddItem(strings.Repeat("b", 100))
	bottomBar.Draw(simScreen)

	//Nor do we intend to crash with a zero height.
	bottomBar.SetRect(0, 0, width, 0)
	bottomBar.Draw(simScreen)
}
