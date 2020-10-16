// Demo code which illustrates how to implement your own primitive.
package main

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

// RadioButtons implements a simple primitive for radio button selections.
type RadioButtons struct {
	*tview.Box
	options       []string
	currentOption int
}

// NewRadioButtons returns a new radio button primitive.
func NewRadioButtons(options []string) *RadioButtons {
	return &RadioButtons{
		Box:     tview.NewBox(),
		options: options,
	}
}

// Draw draws this primitive onto the screen.
func (r *RadioButtons) Draw(screen tcell.Screen) bool {
	res := r.Box.Draw(screen)
	if !res {
		return false
	}

	x, y, width, height := r.GetInnerRect()

	for index, option := range r.options {
		if index >= height {
			break
		}
		radioButton := "\u25ef" // Unchecked.
		if index == r.currentOption {
			radioButton = "\u25c9" // Checked.
		}
		line := fmt.Sprintf(`%s[white]  %s`, radioButton, option)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}

	return true
}

// InputHandler returns the handler for this primitive.
func (r *RadioButtons) InputHandler() tview.InputHandlerFunc {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			r.currentOption--
			if r.currentOption < 0 {
				r.currentOption = 0
			}
			return nil
		case tcell.KeyDown:
			r.currentOption++
			if r.currentOption >= len(r.options) {
				r.currentOption = len(r.options) - 1
			}
			return nil
		}

		return event
	})
}

func main() {
	radioButtons := NewRadioButtons([]string{"Lions", "Elephants", "Giraffes"})
	radioButtons.SetBorder(true).
		SetTitle("Radio Button Demo").
		SetRect(0, 0, 30, 5)
	if err := tview.NewApplication().SetRoot(radioButtons, false).Run(); err != nil {
		panic(err)
	}
}
