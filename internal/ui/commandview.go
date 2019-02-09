package ui

import (
	"strings"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

// CommandView contains a simple textview for output and an input field for
// input. All commands are added to the history when confirmed via enter.
type CommandView struct {
	commandOutput *tview.TextView
	commandInput  *tview.InputField

	commandHistoryIndex int
	commandHistory      []string

	onExecuteCommand func(command string)
}

// NewCommandView creates a new struct containing the components necessary
// for a command view. It also contains the state for those components.
func NewCommandView(onExecuteCommand func(command string)) *CommandView {
	commandOutput := tview.NewTextView()
	commandOutput.SetBorder(true)
	commandOutput.SetDynamicColors(true)

	commandInput := tview.NewInputField()
	commandInput.SetBorder(true)

	cmdView := &CommandView{
		commandOutput: commandOutput,
		commandInput:  commandInput,

		commandHistoryIndex: -1,
		commandHistory:      make([]string, 0),

		onExecuteCommand: onExecuteCommand,
	}

	commandInput.SetInputCapture(cmdView.handleInput)

	return cmdView
}

func (cmdView *CommandView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEnter {
		cmdView.commandHistoryIndex = -1
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

	return event
}

// GetCommandInputWidget returns the component that can be added to the layout
// for the users command input.
func (cmdView *CommandView) GetCommandInputWidget() *tview.InputField {
	return cmdView.commandInput
}

// GetCommandOutputWidget is the component that can be added to the layout
// for the users command output.
func (cmdView *CommandView) GetCommandOutputWidget() *tview.TextView {
	return cmdView.commandOutput
}

// SetVisible sets the given visible state to both the input component and
// the output component
func (cmdView *CommandView) SetVisible(visible bool) {
	cmdView.commandInput.SetVisible(visible)
	cmdView.commandOutput.SetVisible(visible)
}
