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
	globalScope        = addScope("Application wide", "global", nil)
	multilineTextInput = addScope("Multiline text input", "multiline_text_input", globalScope)
	chatview           = addScope("Chatview", "chatview", globalScope)

	QuoteSelectedMessage = addShortcut("quote_selected_message", "Quote selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	EditSelectedMessage = addShortcut("edit_selected_message", "Edit selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ReplySelectedMessage = addShortcut("reply_selected_message", "Reply to author selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	CopySelectedMessageLink = addShortcut("copy_selected_message_link", "Copy link to selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	CopySelectedMessage = addShortcut("copy_selected_message", "Copy content of selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ToggleSelectedMessageSpoilers = addShortcut("toggle_selected_message_spoilers", "Toggle spoilers in selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	DeleteSelectedMessage = addShortcut("toggle_selected_message_spoilers", "Toggle spoilers in selected message",
		chatview, tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone))

	ExpandSelectionToLeft = addShortcut("expand_selection_word_to_left", "Expand selection word to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModShift))
	ExpandSelectionToRight = addShortcut("expand_selection_word_to_right", "Expand selection word to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift))
	SelectAll = addShortcut("select_all", "Select all",
		multilineTextInput, tcell.NewEventKey(tcell.KeyCtrlA, rune(tcell.KeyCtrlA), tcell.ModCtrl))
	SelectWordLeft = addShortcut("select_word_to_left", "Select word to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl|tcell.ModShift))
	SelectWordRight = addShortcut("select_word_to_right", "Select word to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl|tcell.ModShift))

	MoveCursorLeft = addShortcut("move_cursor_to_left", "Move cursor to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
	MoveCursorRight = addShortcut("move_cursor_to_right", "Move cursor to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
	MoveCursorWordLeft = addShortcut("move_cursor_to_word_left", "Move cursor to word left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl))
	MoveCursorWordRight = addShortcut("move_cursor_to_word_right", "Move cursor to word right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl))

	// FIXME Gotta add this later, as there is Backspace and Backspace and those differ on linux.
	// DeleteLeft = addShortcut("delete_left","Delete left",multilineTextInput,tcell.NewEventKey(tcell.KeyBackspace2, rune(tcell.KeyBackspace2), tcell.ModNone))

	DeleteRight = addShortcut("delete_right", "Delete right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone))
	InputNewLine = addShortcut("add_new_line_character", "Add new line character",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModAlt))

	CopySelection = addShortcut("copy_selection", "Copy selected text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModAlt))
	PasteAtSelection = addShortcut("paste_at_selectiom", "Paste clipboard content",
		multilineTextInput, tcell.NewEventKey(tcell.KeyCtrlV, rune(tcell.KeyCtrlV), tcell.ModCtrl))

	SendMessage = addShortcut("send_message", "Sends the typed message",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone))

	ExitApplication = addShortcut("exit_application", "Exit application",
		globalScope, tcell.NewEventKey(tcell.KeyCtrlC, rune(tcell.KeyCtrlC), tcell.ModCtrl))

	FocusChannelContainer = addShortcut("focus_channel_container", "Focus channel container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt))
	FocusUserContainer = addShortcut("focus_user_container", "Focus user container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt))
	FocusGuildContainer = addShortcut("focus_guild_container", "Focus guild container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt))
	FocusPrivateChatPage = addShortcut("focus_private_chat_page", "Focus private chat page",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModAlt))
	FocusPreviousChannel = addShortcut("focus_previous_channel", "Focus previous channel",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModAlt))
	FocusMessageInput = addShortcut("focus_message_input", "Focus message input",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModAlt))
	FocusMessageContainer = addShortcut("focus_message_container", "Focus message container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt))
	FocusCommandInput = addShortcut("focus_command_input", "Focus command input",
		globalScope, tcell.NewEventKey(tcell.KeyCtrlI, rune(tcell.KeyCtrlI), tcell.ModNone))
	FocusCommandOutput = addShortcut("focus_command_output", "Focus command output",
		globalScope, tcell.NewEventKey(tcell.KeyCtrlO, rune(tcell.KeyCtrlO), tcell.ModCtrl))

	ToggleUserContainer = addShortcut("toggle_user_container", "Toggle user container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'U', tcell.ModAlt))
	ToggleCommandView = addShortcut("toggle_command_view", "Toggle command view",
		globalScope, tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt))

	scopes    []*Scope
	Shortcuts []*Shortcut
)

func addScope(name, identifier string, parent *Scope) *Scope {
	scope := &Scope{
		Parent:     parent,
		Identifier: identifier,
		Name:       name,
	}

	scopes = append(scopes, scope)

	return scope
}

func addShortcut(name, identifier string, scope *Scope, event *tcell.EventKey) *Shortcut {
	shortcut := &Shortcut{
		Identifier:   identifier,
		Name:         name,
		scope:        scope,
		Event:        event,
		defaultEvent: event,
	}

	Shortcuts = append(Shortcuts, shortcut)

	return shortcut
}

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
