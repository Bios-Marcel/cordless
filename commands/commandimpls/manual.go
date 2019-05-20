package commandimpls

import (
	"fmt"
	"io"
	"strings"
)

const manualDocumentation = `[orange][::u]# manual[white]

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
		switch strings.ToLower(parameters[0]) {
		case "chat-view", "chatview":
			fmt.Fprintln(writer, chatViewDocumentation)
		case "commands":
			fmt.Fprintln(writer, commandsDocumentation)
		case "configuration":
			fmt.Fprintln(writer, configurationDocumentation)
		case "message-editor", "messageeditor":
			fmt.Fprintln(writer, messageEditorDocumentation)
		case "navigation":
			fmt.Fprintln(writer, navigationDocumentation)
		default:
			manual.PrintHelp(writer)
		}
	} else {
		manual.PrintHelp(writer)
	}
}

const chatViewDocumentation = `[orange][::u]# Chatview[white]

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
const commandsDocumentation = `[orange][::u]# Commands[white]

All commands can only be entered via the command-input component. Commands can't be called from outside the application or on startup.

All commands follow a certain symantics pattern:
	COMMAND SUBCOMMAND --SETTING "Some setting value" MAIN_VALUE
	
Not every command makes use of all of the possible combinations. Each command may have zero or more subcommands and zero or more settings. There may also be settings that do not require you passing a value. If a value contains spaces it needs to be quoted beforehand, otherwise the input will be seperated at each given space. Some commands require some main value, which is basically the non-optional input for that command. That value doesn't require a setting-name to be prepended in front of it.

After typing a command, it will be added to your history. The history doesn't persist between cordless sessions, it will be forgotten every time you close the application. The history can be travelled through by using the arrow up and down keys. An exception for historization are secret inputs like passwords, those aren't directly typed into the command-input. Instead cordless shows an extra dialog that as soon as you are required to input sensitive information like passwords.

Since the command-input component uses the same underlying component as the message-input, you can use the same shortcuts for editing your input.
`
const configurationDocumentation = `[orange][::u]# Configuration[white]`
const messageEditorDocumentation = `[orange][::u]# Message editor[white]

The editor is a custom written widget and builds on top of the tview.TextView.
It utilizes regions and highlighting in order to implement the complete selection behaviour
Currently it is not a finished widget. 

## Shortcuts

----------------------------------------------------
|           Action           |       Shortcut      |
| -------------------------- | ------------------- |
| Delete left                | Backspace           |
| Delete Right               | Delete              |
| Delete Selection           | Backspace or Delete |
| Jump to beginning          | Ctrl+A -> Left      |
| Jump to end                | Ctrl+A -> Right     |
| Jump one word to the left  | Ctrl+Left           |
| Jump one word to the right | Ctrl+Right          |
| Select all                 | Ctrl+A              |
| Select word to left        | Ctrl+Shift+Left     |
| Select word to right       | Ctrl+Shift+Right    |
| Scroll chatview up         | Ctrl+Up             |
| Scroll chatview down       | Ctrl+Down           |
| Paste Image / text         | Ctrl+V              |
| Insert new line            | Alt+Enter           |
| Send message               | Enter               |
----------------------------------------------------

## Other features

* Send emojis using ":emoji_code:"
* Mention people using autocomplete by typing an "@" followed by part of their name
`

const navigationDocumentation = `[orange][::u]# Navigation[white]

Most of the controlling is currently done via the keyboard. However, the focus
between components can be changed by using a mouse as well.

[::u]## Shortcuts

--------------------------------------------------------------------
|          Action         | Shortcut |            Scope            |
| ----------------------- | -------- | ----------------------------|
| Close application       | Ctrl-C   | Everywhere                  |
| Focus user container    | Alt+U    | Guild channel / group chat  |
| Focus private chat page | Alt+P    | Everywhere                  |
| Focus guild container   | Alt+S    | Everywhere                  |
| Focus channel container | Alt+C    | Everywhere                  |
| Focus message input     | Alt+M    | Everywhere                  |
| Focus message container | Alt+T    | Everywhere                  |
| Toggle command view     | Alt+Dot  | Everywhere                  |
| Focus command output    | Ctrl+O   | Everywhere                  |
| Focus command input     | Ctrl+I   | Everywhere                  |
| Edit last message       | ArrowUp  | In empty message input      |
| Leave message edit mode | Esc      | When editing message        |
--------------------------------------------------------------------

Some shortcuts can be changed via the shortcut dialog. The dialog can be opened via Alt+Shift+S.
`

// Name represents this commands indentifier.
func (manual *Manual) Name() string {
	return "manual"
}

// PrintHelp prints a static help page for this command
func (manual *Manual) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, manualDocumentation)
}
