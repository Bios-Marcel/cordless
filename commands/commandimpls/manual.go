package commandimpls

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

const manualDocumentation = `[::b]NAME
	manual - view the documentation for a topic or command

[::b]SYNOPSIS
	[::b]manual <topic|command>

[::b]DESCRIPTION
	This command will show the manual of the supplied topic / command.

[::b]TOPICS
	Besides commands like [::b]friends[::-] or [::b]status[::-], you can also supply one of the following topics:
		- commands
		- chat-view
		- configuration
		- message-editor
		- navigation

[::b]EXAMPLES
	[gray]$ man user
	[white][::b]NAME
		user - manipulate and retrieve your ...
`

// Manual is the command that displays the application manual.
type Manual struct {
	window *ui.Window
}

// NewManualCommand constructs a new usable manual command for the user.
func NewManualCommand(window *ui.Window) *Manual {
	return &Manual{window}
}

func (manual *Manual) getPredefinedTopicPage(name string) string {
	switch name {
	case "chat-view", "chatview":
		return chatViewDocumentation
	case "commands":
		var commandList string
		for _, cmd := range manual.window.GetRegisteredCommands() {
			commandList += fmt.Sprintf("\t\t- %s\n", cmd.Name())
		}
		return fmt.Sprintf(commandsDocumentation, commandList)
	case "configuration", "config", "conf":
		return configurationDocumentation
	case "message-editor", "messageeditor":
		return messageEditorDocumentation
	case "navigation":
		return navigationDocumentation
	}

	return ""
}

// Execute runs the command piping its output into the supplied writer.
func (manual *Manual) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		manual.PrintHelp(writer)
	} else {
		//All parameters will be combined into words joined with dashes. Since that's how
		//command names are represented internally.
		inputDashes := strings.ToLower(strings.Join(parameters, "-"))

		//This handles the default pages, e.g. the ones that are unrelated to commands.
		page := manual.getPredefinedTopicPage(inputDashes)
		if page != "" {
			fmt.Fprintln(writer, page)
			return
		}

		//This will check whether the input matches either a commands name or one of it's aliases.
		for _, cmd := range manual.window.GetRegisteredCommands() {
			if cmd.Name() == inputDashes {
				cmd.PrintHelp(writer)
				return
			}

			for _, alias := range cmd.Aliases() {
				if alias == inputDashes {
					cmd.PrintHelp(writer)
					return
				}
			}
		}

		inputSpaces := strings.ToLower(strings.Join(parameters, " "))
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]No manual entry for '%s' found.\n", inputSpaces)

		//This code checks whether any of the command pages contains the given terms.
		var b = bytes.Buffer{}
		var matches []string
		for _, cmd := range manual.window.GetRegisteredCommands() {
			cmd.PrintHelp(&b)
			if strings.Contains(strings.ToLower(b.String()), inputSpaces) {
				matches = append(matches, cmd.Name())
			}
			b.Reset()
		}

		if len(matches) > 0 {
			fmt.Fprintln(writer, "The following pages contain the given term:")
			for _, match := range matches {
				fmt.Fprintf(writer, "\t%s\n", match)
			}
		}
	}
}

const chatViewDocumentation = `[::b]TOPIC
	chatview - component that displays the messages in a channel

[::b]DESCRIPTION
	The chatview is the component that displays the messages of the channel
	that you are currently looking at.

	The chatview has two modes. The navigation mode is active when the
	chatview does not have focus. While in navigation mode you can't select
	any message, but you can scroll through the messages using Ctrl+ArrowUp/ArrowDown
	or your mousewheel. If you focus the chatview, it enters the selection
	mode, allowing you to select single messages and interact with those.

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
	| Selection to bottom         | End        |
	--------------------------------------------

	Keep in mind, that those shortcuts might differ from your settings, as
	those are just the defaults.`

const commandsDocumentation = `[::b]TOPIC
	commands - commands allow you to execute certain actions within cordless

[::b]DESCRIPTION
	Commands can only be entered via the command-input component.
	Commands can't be called from outside the application or on startup.

	All commands follow a certain semantics pattern:
		COMMAND SUBCOMMAND --SETTING "Some setting value" MAIN_VALUE

	Not every command makes use of all of the possible combinations. Each
	command may have zero or more subcommands and zero or more settings.
	There may also be settings that do not require you passing a value.
	If a value contains spaces it needs to be quoted beforehand, otherwise
	the input will be separated at each given space. Some commands require
	some main value, which is basically the non-optional input for that
	command. That value doesn't require a setting-name to be prepended in
	front of it.

	After typing a command, it will be added to your history. The history
	doesn't persist between cordless sessions, it will be forgotten every
	time you close the application. The history can be travelled through
	by using the arrow up and down keys. An exception for historization
	are secret inputs like passwords, those aren't directly typed into the
	command-input. Instead cordless shows an extra dialog as soon as it
	requires you to input sensitive information like passwords.

	Since the command-input component uses the same underlying component
	as the message-input, you can use the same shortcuts for editing
	your input.

	Available commands:
%s
[::b]EXAMPLES
	[gray]$ user-set -n "Marcel Schramm" -a /home/pics/avatar.png
	[gray]$ user-set --name "Marcel Schramm" --avatar /home/pics/avatar.png
	
	[gray]$ status set online
	[gray]$ status get`
const configurationDocumentation = `[::b]TOPIC
	configuration - allows you to change settings and persist them between
	sessions

[::b]DESCRIPTION
	Currently all almost configuration is done via manually editing the
	configuration file. There are however some settings like the fix-layout
	setting and the chatheader setting that can be set via the commands
	feature.

	At some point there will be a user-interface for changing settings and
	commands will be removed.

	The configuration file can be found somewhere in the user directory.
	The precise location differs from platform to platform. Whenever you
	start cordless, it will display the location of your configuration file
	in the splashscreen.

	Typical location on Linux:   [::b]~/.config/cordless/config.json[::-].
	Typical location on Windows: [::b]~/AppData/Roaming/cordless/config.json[::-].
	Typical location on MacOS:   [::b]~/.cordless/config.json[::-].

[::b]SETTINGS
	The following settings are available in the configuration file:

	[::b]Token
		The token is used in order to authenticate you in the discord backend.
		This value is usually set through the user interface on startup.
		
		Type:    string
		Default: EMPTY

	[::b]Times
		Determines how message timestamps are rendered in the chatview.

		This setting has three different possible values:

		-------------------------------------------
		|         Name         |  Format  | Value |
		| -------------------- | -------- | ----- |
		| HourMinuteAndSeconds | HH:MM:SS | 0     |
		| HourAndMinute        | HH:MM    | 1     |
		| NoTime               | None     | 2     |
		-------------------------------------------

		Type:    int
		Default: NoTime (2)
		
	[::b]UserColors
		Determines how the color for a user is decided when rendering
		a message author or displaying a user somewhere else.

		This settings has four different possible values:
		  * "none"
		  * "single"
		  * "random"
		  * "role"

		Type:    string
		Default: "single"
		
	[::b]FocusChannelAfterGuildSelection
		Determines whether the focus automatically jumps to the channeltree
		after selecting a guild from the guildlist.
		
		Type:    boolean
		Default: true
	
	[::b]FocusMessageInputAfterChannelSelection
		Determines whether the focus automatically jumps to the message-input
		after selecting a channel from the channeltree.
		
		Type:    boolean
		Default: true
		
	[::b]ShowChatHeader
		Determines whether the name and topic of the currently loaded channel
		are shown at the top of the chatview.
		
		Type:    boolean
		Default: true
		
	[::b]ShowUserContainer
		Determines whether the user list is displayed to the right of the
		chatview while a guild channel or a group dm channel is loaded.
		This setting is usually not changed manually, but via the keyboard
		shortcut [::b]Alt+Shift+U[::-]. Note that this might differ from your
		configured shortcut for this action.
		
		Type:    boolean
		Default: true
		
	[::b]UseFixedLayout
		Determines whether the guild list and the channel tree use a fixed
		width or take horizontal space relative to the window size.
		
		Type:    boolean
		Default: false
		
	[::b]FixedSizeLeft
		Determines the width of the guild list and the channel tree.
		This setting only takes effect if [::b]UseFixedLayout[::-] is set to [::b]true[::-].
		
		Type:    int
		Default: 12
	
	[::b]FixedSizeRight
		Determines the width of the user list next to the chatview.
		This setting only takes effect if [::b]UseFixedLayout[::-] is set to [::b]true[::-].
		
		Type:    int
		Default: 12
		
	[::b]OnTypeInListBehaviour
		Determines whether typing in a list or tree-list will trigger a text
		search, do nothing or focus the message-input.
		
		This settings has three different possible values:
		
		-----------------------------------------
		|             Name              | Value |
		| ----------------------------- | ----- |
		| DoNothingOnTypeInList         | 0     |
		| SearchOnTypeInList            | 1     |
		| FocusMessageInputOnTypeInList | 2     |
		-----------------------------------------

		Type:    int
		Default: SearchOnTypeInList (1)
		
	[::b]MouseEnabled
		Determines whether the mouse is properly usable. If this settings is
		enabled, you can click buttons, scroll and change focus via clicking.
		Note that some stuff might still work with this setting disabled, but
		behave in a weird way.

		[::b]This setting will break terminal selection mode.
		
		Type:    boolean
		Default: true
		
	[::b]ShortenLinks
		Determines whether cordless runs it's own link-shortener in order to
		allow showing long links in tight spaces while keeping them clickable.
		Note that this setting will start a small internal in-memory http
		server.
		
		Type:    boolean
		Default: false

	[::b]ShortenWithExtension
		Determines whether the suffix is added to the shortened url. This
		setting only matters if [::b]ShortenLinks[::-] is set to [::b]true[::-]

		Type:    boolean
		Default: false
		
	[::b]ShortenerPort
		Determines which port the link-shortener uses in your system. This
		setting only matters if [::b]ShortenLinks[::-] is set to [::b]true[::-]
		
	[::b]DesktopNotifications
		Determines whether cordless will try to notify the host systems using
		the systems notification system. This setting might not work on all
		systems.
		
		Type:    boolean
		Default: true
		
	[::b]ShowPlaceholderForBlockedMessages
		Determines whether blocked messages are hidden or a placeholder is
		shown instead, so that you know that someone sent a message. This
		might help in avoiding confusion during conversations of blocked
		users with users that aren't blocked.
		
		Type:    boolean
		Default: true

	[::b]Accounts
		This settings holds an array of so called accounts, also referred to
		as profiles. Those allow you to let cordless know of multiple discord
		identities, allowing you to easily switch between them. This setting
		shouldn't be changed manually, but only via the [::b]account[::-]
		command.

	[::b]IndicateChannelAccessRestriction
		Decides whether a padlock emoji will be displayed next to the
		channelname in the channeltree if the channel isn't accessible
		to all users in a server.

	[::b]ShowBottomBar
		Decides whether the information bar at the bottom is displayed or not.
		It contains information about the currently logged in account and
		displays the shortcut to change keybindings. There might be more in
		here at some point.

	[::b]ShowNicknames
		Decides whether a users nickname is displayed throughout cordless.`

const messageEditorDocumentation = `[::b]TOPIC
	message-editor - the component that allows you to input text for a message.

[::b]DESCRIPTION
	The editor is a custom written widget and builds on top of the
	tview.TextView. It utilizes regions and highlighting in order to implement
	the complete selection behaviour. This widget isn't fully implemented yet
	and still has some flaws.

	By default, it offers the following shortcuts:

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

	It also offers the following functionalities:
		- Send emojis using ":emoji_code:"
		- Mention people using autocomplete by typing an "@" followed by part
		  of their name`

const navigationDocumentation = `[::b]TOPIC
	navigation - how to navigate around the application

[::b]DESCRIPTION
	Most of the controlling is currently done via the keyboard. However, the
	focus between components can be changed by using a mouse as well.

	By default the navigation is done via the following shortcuts:

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

	Some shortcuts can be changed via the shortcut dialog. The dialog can be
	opened via Ctrl+K.`

func (manual *Manual) Name() string {
	return "manual"
}

func (manual *Manual) Aliases() []string {
	return []string{"man", "help"}
}

// PrintHelp prints a static help page for this command
func (manual *Manual) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, manualDocumentation)
}
