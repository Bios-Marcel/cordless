package ui

import (
	"strings"

	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

const noHistoryIndexSelected = -1

// CommandView contains a simple textview for output and an input field for
// input. All commands are added to the history when confirmed via enter.
type CommandView struct {
	commandOutput *tview.TextView
	commandInput  *Editor

	// commandHistoryIndex is the current index cycling through the history.
	// -1 means that no index is selected.
	commandHistoryIndex int
	// commandHistory is just a flat array of text that the user has sent.
	// The more recent ones are on the right side of the array.
	commandHistory []string

	onExecuteCommand func(command string)
}

// NewCommandView creates a new struct containing the components necessary
// for a command view. It also contains the state for those components.
func NewCommandView(app *tview.Application, onExecuteCommand func(command string)) *CommandView {
	commandOutput := tview.NewTextView()
	commandOutput.SetDynamicColors(true).
		SetWordWrap(true).
		SetWrap(true).
		SetText("[::b]### Welcome back. ###\n	If you need to know more, run the [::b]man[::-] command.\n").
		SetBorder(true).
		SetIndicateOverflow(true)

	commandInput := NewEditor(app)
	commandInput.internalTextView.
		SetWrap(false).
		SetWordWrap(false)

	cmdView := &CommandView{
		commandOutput: commandOutput,
		commandInput:  commandInput,

		commandHistoryIndex: noHistoryIndexSelected,
		commandHistory:      make([]string, 0),

		onExecuteCommand: onExecuteCommand,
	}

	commandInput.SetInputCapture(cmdView.handleInput)

	return cmdView
}

// SetInputCaptureForInput defines the input capture for the input component of
// the command view while prioritizing the predefined handler before passing the
// event to the externally specified handler.
func (cmdView *CommandView) SetInputCaptureForInput(handler func(event *tcell.EventKey) *tcell.EventKey) {
	cmdView.commandInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		passedEvent := cmdView.handleInput(event)

		if passedEvent == nil {
			return nil
		}

		return handler(passedEvent)
	})
}

// SetInputCaptureForOutput defines the input capture for the output component
// of the command view while prioritizing the predefined handler before passing
// the event to the externally specified handler.
func (cmdView *CommandView) SetInputCaptureForOutput(handler func(event *tcell.EventKey) *tcell.EventKey) {
	cmdView.commandOutput.SetInputCapture(handler)
}

func (cmdView *CommandView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Modifiers() == tcell.ModNone {
		if event.Key() == tcell.KeyPgUp {
			handler := cmdView.commandOutput.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone), nil)
			return nil
		}

		if event.Key() == tcell.KeyPgDn {
			handler := cmdView.commandOutput.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), nil)
			return nil
		}

		if event.Key() == tcell.KeyEnter {
			//We are resetting the index whenever hitting enter, no matter
			//whether the command itself was run successfully or not.
			//So that when you cycle up again, you'll be at the most recent command,
			//meaning the one that you just entered.
			cmdView.commandHistoryIndex = noHistoryIndexSelected

			command := cmdView.commandInput.GetText()
			if command == "" {
				return nil
			}

			cmdView.onExecuteCommand(strings.TrimSpace(command))
			cmdView.commandInput.SetText("")
			cmdView.commandHistory = append(cmdView.commandHistory, command)

			return nil
		}

		if event.Key() == tcell.KeyDown {
			if cmdView.commandHistoryIndex > len(cmdView.commandHistory)-1 {
				cmdView.commandHistoryIndex = 0
			} else {
				cmdView.commandHistoryIndex++
			}

			if cmdView.commandHistoryIndex > len(cmdView.commandHistory)-1 {
				return nil
			}

			cmdView.commandInput.SetText(cmdView.commandHistory[cmdView.commandHistoryIndex])
		}

		if event.Key() == tcell.KeyUp {
			if cmdView.commandHistoryIndex < 0 {
				cmdView.commandHistoryIndex = len(cmdView.commandHistory) - 1
			} else {
				cmdView.commandHistoryIndex--
			}

			if cmdView.commandHistoryIndex < 0 {
				return nil
			}

			cmdView.commandInput.SetText(cmdView.commandHistory[cmdView.commandHistoryIndex])
		}
	}

	if event.Modifiers() == tcell.ModCtrl {
		if event.Key() == tcell.KeyUp {
			handler := cmdView.commandOutput.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil)
			return nil
		}

		if event.Key() == tcell.KeyDown {
			handler := cmdView.commandOutput.InputHandler()
			handler(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
			return nil
		}
	}

	return event
}

// GetCommandInputWidget returns the component that can be added to the layout
// for the users command input.
func (cmdView *CommandView) GetCommandInputWidget() *tview.TextView {
	return cmdView.commandInput.internalTextView
}

// GetCommandOutputWidget is the component that can be added to the layout
// for the users command output.
func (cmdView *CommandView) GetCommandOutputWidget() *tview.TextView {
	return cmdView.commandOutput
}

// SetVisible sets the given visible state to both the input component and
// the output component
func (cmdView *CommandView) SetVisible(visible bool) {
	cmdView.commandInput.internalTextView.SetVisible(visible)
	cmdView.commandOutput.SetVisible(visible)
}

// Write lets us implement the io.Writer interface. Tab characters will be
// replaced with TabSize space characters. A "\n" or "\r\n" will be interpreted
// as a new line.
func (cmdView *CommandView) Write(p []byte) (n int, err error) {
	n, err = cmdView.commandOutput.Write(p)
	if err == nil {
		cmdView.commandOutput.ScrollToEnd()
	}

	return
}
