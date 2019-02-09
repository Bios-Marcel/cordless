package commandimpls

import (
	"fmt"
	"io"
)

const manualDocumentation = `[orange]# manual[white]

Please supply a topic that you want to know more about.
	[-]manual <topic>[white]

Available topics:
	* chat-view
	* commands
	* configuration
	* message-editor
	* navigation
`

// Manual is the command that displays the application manual.
type Manual struct{}

// NewManualCommand constructs a new usable manual command for the user.
func NewManualCommand() *Manual {
	return &Manual{}
}

// Execute runs the command piping its output into the supplied writer.
func (manual *Manual) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 1 {
		switch parameters[0] {
		case "chat-view":
			fmt.Fprintln(writer, chatViewDocumentation)

		case "commands":
			fmt.Fprintln(writer, commandsDocumentation)

		case "configuration":
			fmt.Fprintln(writer, configurationDocumentation)

		case "message-editor":
			fmt.Fprintln(writer, messageEditorDocumentation)

		case "navigation":
			fmt.Fprintln(writer, navigationDocumentation)

		}
	} else {
		manual.PrintHelp(writer)
	}
}

const chatViewDocumentation = `[orange]# Chatview[white]

The Chatview is the component that displays the messages of the channel that you are currently looking at.

The Chatview has two modes, the navigation mode is activate when the chatview does not have focus. While in navigation mode you can't select any message, you can only scroll through the view. If you focus the chatview, the it enters the selection mode and you can select single messages and interact with them.

When in selection mode, those shortcuts are active:

--------------------------------------------
|            Action           |  Shortcut  |
| --------------------------- | ---------- |
| Edit message                | e          |
| Delete message              | Delete     |
| Copy content                | c          |
| Copy link to message        | l          |
| Reply with mention          | r          |
| Quote message               | q          |
| Hide / show spoiler content | s          |
| Selection up                | ArrowUp    |
| Selection down              | ArrowDown  |
| Selection to top            | Home       |
| Selectino to bottom         | End        |
--------------------------------------------
`
const commandsDocumentation = `[orange]# Commands[white]`
const configurationDocumentation = `[orange]# Configuration[white]`
const messageEditorDocumentation = `[orange]# Message editor[white]`
const navigationDocumentation = `[orange]# Navigation[white]`

// Name represents this commands indentifier.
func (manual *Manual) Name() string {
	return "manual"
}

// PrintHelp prints a static help page for this command
func (manual *Manual) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, manualDocumentation)
}
