package ui

import (
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
)

// PrompSecretSingleLineInput shows a fullscreen input dialog that masks the
// user input. The returned value will either be empty or what the user
// has entered.
func PrompSecretSingleLineInput(app *tview.Application, title, message string) string {
	return promptSingleLineInput(app, '*', title, message)
}

// PromptSingleLineInput shows a fullscreen input dialog.
// The returned value will either be empty or what the user has entered.
func PromptSingleLineInput(app *tview.Application,
	activePrimitiveChanged func(primitive tview.Primitive), title, message string) string {
	return promptSingleLineInput(app, 0, title, message)
}

func promptSingleLineInput(app *tview.Application, maskCharacter rune, title, message string) string {

	waitChannel := make(chan struct{})
	var output string
	var previousFocus tview.Primitive
	previousRoot := app.GetRoot()
	app.QueueUpdateDraw(func() {
		previousFocus = app.GetFocus()
		inputField := tview.NewInputField()
		inputField.SetMaskCharacter(maskCharacter)
		inputField.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				output = inputField.GetText()
				waitChannel <- struct{}{}
			} else if key == tcell.KeyEscape {
				waitChannel <- struct{}{}
			}
		})
		inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if shortcuts.PasteAtSelection.Equals(event) {
				content, clipError := clipboard.ReadAll()
				if clipError == nil {
					//FIXME It'd be cool to insert at cursor somehow.
					inputField.SetText(content)
				}
				return nil
			}

			return event
		})
		frame := tview.NewFrame(inputField)
		frame.SetTitle(title)
		frame.SetBorder(true)
		frame.AddText(message, true, tview.AlignLeft, tcell.ColorDefault)
		app.SetRoot(frame, true)
	})
	<-waitChannel
	app.QueueUpdateDraw(func() {
		app.SetRoot(previousRoot, true)
		app.SetFocus(previousFocus)
		waitChannel <- struct{}{}
	})
	<-waitChannel
	return output
}
