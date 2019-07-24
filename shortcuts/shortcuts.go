package shortcuts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/gdamore/tcell"
)

var (
	globalScope = &Scope{
		Parent:     nil,
		Identifier: "global",
		Name:       "Application wide",
	}

	multilineTextInput = &Scope{
		Parent:     globalScope,
		Identifier: "multiline_text_input",
		Name:       "Multiline text input",
	}

	ExpandSelectionToLeft = &Shortcut{
		Name:         "Expand selection word to left",
		Identifier:   "expand_selection_word_to_left",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModShift),
		defaultEvent: tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModShift),
	}
	ExpandSelectionToRight = &Shortcut{
		Name:         "Expand selection word to right",
		Identifier:   "expand_selection_word_to_right",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift),
		defaultEvent: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift),
	}
	SelectAll = &Shortcut{
		Name:         "Select all",
		Identifier:   "select_all",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyCtrlA, rune(tcell.KeyCtrlA), tcell.ModCtrl),
		defaultEvent: tcell.NewEventKey(tcell.KeyCtrlA, rune(tcell.KeyCtrlA), tcell.ModCtrl),
	}
	SelectWordLeft = &Shortcut{
		Name:         "Select word to left",
		Identifier:   "select_word_to_left",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl|tcell.ModShift),
		defaultEvent: tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl|tcell.ModShift),
	}

	SelectWordRight = &Shortcut{
		Name:         "Select word to right",
		Identifier:   "select_word_to_right",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl|tcell.ModShift),
		defaultEvent: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl|tcell.ModShift),
	}

	MoveCursorLeft = &Shortcut{
		Name:         "Move cursor to left",
		Identifier:   "move_cursor_to_left",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone),
	}
	MoveCursorRight = &Shortcut{
		Name:         "Move cursor to right",
		Identifier:   "move_cursor_to_right",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
	}
	MoveCursorWordLeft = &Shortcut{
		Name:         "Move cursor to word left",
		Identifier:   "move_cursor_to_word_left",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl),
		defaultEvent: tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl),
	}
	MoveCursorWordRight = &Shortcut{
		Name:         "Move cursor to word right",
		Identifier:   "move_cursor_to_word_right",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl),
		defaultEvent: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl),
	}

	// FIXME Gotta add this later, as there is Backspace and Backspace and those differ on linux.
	/*DeleteLeft = &Shortcut{
		Name:         "Delete left",
		Identifier:   "delete_left",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyBackspace2, rune(tcell.KeyBackspace2), tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyBackspace2, rune(tcell.KeyBackspace2), tcell.ModNone),
	}*/

	DeleteRight = &Shortcut{
		Name:         "Delete right",
		Identifier:   "delete_right",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone),
	}

	InputNewLine = &Shortcut{
		Name:         "Add new line character",
		Identifier:   "add_new_line_character",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModAlt),
	}

	ExitApplication = &Shortcut{
		Name:         "Exit application",
		Identifier:   "exit_application",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyCtrlC, rune(tcell.KeyCtrlC), tcell.ModCtrl),
		defaultEvent: tcell.NewEventKey(tcell.KeyCtrlC, rune(tcell.KeyCtrlC), tcell.ModCtrl),
	}

	FocusChannelContainer = &Shortcut{
		Name:         "Focus channel container",
		Identifier:   "focus_channel_container",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt),
	}

	FocusUserContainer = &Shortcut{
		Name:         "Focus user container",
		Identifier:   "focus_user_container",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt),
	}

	ToggleUserContainer = &Shortcut{
		Name:         "Toggle user container",
		Identifier:   "toggle_user_container",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 'U', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 'U', tcell.ModAlt),
	}

	FocusGuildContainer = &Shortcut{
		Name:         "Focus guild container",
		Identifier:   "focus_guild_container",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt),
	}

	FocusPrivateChatPage = &Shortcut{
		Name:         "Focus private chat page",
		Identifier:   "focus_private_chat_page",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModAlt),
	}

	FocusMessageInput = &Shortcut{
		Name:         "Focus message input",
		Identifier:   "focus_message_input",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModAlt),
	}

	FocusMessageContainer = &Shortcut{
		Name:         "Focus message container",
		Identifier:   "focus_message_container",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt),
	}

	FocusCommandInput = &Shortcut{
		Name:         "Focus command input",
		Identifier:   "focus_command_input",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyCtrlI, rune(tcell.KeyCtrlI), tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyCtrlI, rune(tcell.KeyCtrlI), tcell.ModNone),
	}

	FocusCommandOutput = &Shortcut{
		Name:         "Focus command output",
		Identifier:   "focus_command_output",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyCtrlO, rune(tcell.KeyCtrlO), tcell.ModCtrl),
		defaultEvent: tcell.NewEventKey(tcell.KeyCtrlO, rune(tcell.KeyCtrlO), tcell.ModCtrl),
	}

	ToggleCommandView = &Shortcut{
		Name:         "Toggle command view",
		Identifier:   "toggle_command_view",
		scope:        globalScope,
		Event:        tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt),
		defaultEvent: tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt),
	}

	SendMessage = &Shortcut{
		Name:         "Sends the typed message",
		Identifier:   "send_message",
		scope:        multilineTextInput,
		Event:        tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone),
		defaultEvent: tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone),
	}

	scopes = []*Scope{
		globalScope,
		multilineTextInput,
	}

	Shortcuts = []*Shortcut{
		ExitApplication,
		FocusChannelContainer,
		FocusUserContainer,
		ToggleUserContainer,
		FocusGuildContainer,
		FocusPrivateChatPage,
		FocusMessageInput,
		FocusMessageContainer,
		FocusCommandInput,
		FocusCommandOutput,
		ToggleCommandView,
		InputNewLine,
		MoveCursorLeft,
		MoveCursorRight,
		MoveCursorWordLeft,
		MoveCursorWordRight,
		ExpandSelectionToLeft,
		ExpandSelectionToRight,
		SelectAll,
		SelectWordLeft,
		SelectWordRight,
		DeleteRight,
		SendMessage,
	}
)

// Shortcut defines a shortcut within the application. The scope might for
// example be a widget or situation in which the user is.
type Shortcut struct {
	// Identifier will be used for persistence and should never change
	Identifier string

	// Name will be shown on the UI
	Name string

	// The scope will be omitted, as this needed be persisted anyway.
	scope *Scope

	// Event is the shortcut expressed as it's resulting tcell Event.
	Event *tcell.EventKey

	//This shortcuts default, in order to be able to reset it.
	defaultEvent *tcell.EventKey
}

// Equals compares the given EventKey with the Shortcuts Event.
func (s *Shortcut) Equals(event *tcell.EventKey) bool {
	return eventsEqual(s.Event, event)
}

// Scope is what describes a shortcuts scope within the application. Usually
// a scope can only have a specific shortcut once and a children scope will
// overwrite that shortcut, since that lower scope has the upper hand.
type Scope struct {
	// Parent is this scopes upper Scope, which may be null, in case this is a
	// root scope.
	Parent *Scope

	// Identifier will be used for persistence and should never change
	Identifier string

	// Name will be shown on the UI
	Name string
}

// ShortcutDataRepresentation represents a shortcut on the users harddrive.
// This prevents redudancy of scopes.
type ShortcutDataRepresentation struct {
	Identifier      string
	ScopeIdentifier string
	EventKey        tcell.Key
	EventRune       rune
	EventMod        tcell.ModMask
}

// MarshalJSON marshals a Shortcut into a ShortcutDataRepresentation. This
// happens in order to prevent saving the scopes multiple times, therefore
// the scopes will be hardcoded.
func (shortcut *Shortcut) MarshalJSON() ([]byte, error) {
	if shortcut.Event == nil {
		return json.MarshalIndent(&ShortcutDataRepresentation{
			Identifier:      shortcut.Identifier,
			ScopeIdentifier: shortcut.scope.Identifier,
			EventKey:        -1,
			EventRune:       -1,
			EventMod:        -1,
		}, "", "    ")
	}

	return json.MarshalIndent(&ShortcutDataRepresentation{
		Identifier:      shortcut.Identifier,
		ScopeIdentifier: shortcut.scope.Identifier,
		EventKey:        shortcut.Event.Key(),
		EventRune:       shortcut.Event.Rune(),
		EventMod:        shortcut.Event.Modifiers(),
	}, "", "    ")
}

// UnmarshalJSON unmarshals ShortcutDataRepresentation JSON into a Shortcut.
// This happens in order to prevent saving the scopes multiple times,
// therefore the scopes will be hardcoded.
func (shortcut *Shortcut) UnmarshalJSON(data []byte) error {
	var temp ShortcutDataRepresentation
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	if temp.EventKey == -1 && temp.EventMod == -1 && temp.EventRune == -1 {
		shortcut.Event = nil
	} else {
		shortcut.Event = tcell.NewEventKey(temp.EventKey, temp.EventRune, temp.EventMod)
	}
	shortcut.Identifier = temp.Identifier

	for _, scope := range scopes {
		if scope.Identifier == temp.ScopeIdentifier {
			shortcut.scope = scope
			return nil
		}
	}

	return fmt.Errorf(fmt.Sprintf("error finding scope '%s'", temp.ScopeIdentifier))
}

// Reset resets this shortcuts event value to the default one.
func (shortcut *Shortcut) Reset() {
	shortcut.Event = shortcut.defaultEvent
}

func getShortcutsPath() (string, error) {
	configDirecotry, configError := config.GetConfigDirectory()
	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDirecotry, "shortcuts.json"), nil
}

// Load loads the shortcuts and copies the events to the correct shortcuts
// that reside inside of the memory.
func Load() error {
	shortcutsPath, pathError := getShortcutsPath()
	if pathError != nil {
		return pathError
	}

	shortcutsFile, openError := os.Open(shortcutsPath)

	if os.IsNotExist(openError) {
		return nil
	}

	if openError != nil {
		return openError
	}

	defer shortcutsFile.Close()
	decoder := json.NewDecoder(shortcutsFile)
	tempShortcuts := make([]*Shortcut, 0)
	shortcutsLoadError := decoder.Decode(&tempShortcuts)

	//io.EOF would mean empty, therefore we use defaults.
	if shortcutsLoadError != nil && shortcutsLoadError != io.EOF {
		return nil
	}

	if shortcutsLoadError != nil {
		return shortcutsLoadError
	}

OUTER_LOOP:
	for _, shortcut := range tempShortcuts {
		for _, otherShortcut := range Shortcuts {
			if otherShortcut.Identifier == shortcut.Identifier &&
				otherShortcut.scope.Identifier == shortcut.scope.Identifier {
				otherShortcut.Event = shortcut.Event
				continue OUTER_LOOP
			}
		}
	}

	return nil
}

// Persist saves the currently shortcuts that are currently being held in
// memory.
func Persist() error {
	filePath, pathError := getShortcutsPath()
	if pathError != nil {
		return pathError
	}

	shortcutsAsJSON, jsonError := json.MarshalIndent(&Shortcuts, "", "    ")
	if jsonError != nil {
		return jsonError
	}

	writeError := ioutil.WriteFile(filePath, shortcutsAsJSON, 0666)
	if writeError != nil {
		return writeError
	}

	return nil
}
