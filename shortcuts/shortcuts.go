package shortcuts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	tcell "github.com/gdamore/tcell/v2"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/vim"
	//"github.com/Bios-Marcel/cordless/util/vim"
)

var (
	globalScope        = addScope("global", "Application wide", nil)
	multilineTextInput = addScope("multiline_text_input", "Multiline text input", globalScope)
	chatview           = addScope("chatview", "Chatview", globalScope)
	guildlist          = addScope("guildlist", "Guildlist", globalScope)
	channeltree        = addScope("channeltree", "Channeltree", globalScope)

	// Normal mode will always prevail in any scope. To exit insert or visual mode press ESC.
	//
	// A nil VimEvent means the original non-vim key will be used.
	// A NullVimEvent will ignore that mapping inside the selected mode.
	QuoteSelectedMessage = addShortcut("quote_selected_message", "Quote selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	//				Normal		Insert		  Visual
	)

	EditSelectedMessage = addShortcut("edit_selected_message", "Edit selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),NullVimEvent),
	)

	DownloadMessageFiles = addShortcut("download_message_files", "Download all files in selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),
		// FIXME solve delete messages keybinding conflict in visual mode
		addVimEvent(NullVimEvent,NullVimEvent,NullVimEvent),
	)

	ReplySelectedMessage = addShortcut("reply_selected_message", "Reply to author selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone),NullVimEvent),
	)

	NewDirectMessage = addShortcut("new_direct_message", "Create a new direct message channel with this user",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyEnter,rune(tcell.KeyEnter),tcell.ModNone),NullVimEvent),
	)

	CopySelectedMessageLink = addShortcut("copy_selected_message_link", "Copy link to selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	)

	CopySelectedMessage = addShortcut("copy_selected_message", "Copy content of selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone),NullVimEvent),
	)

	ToggleSelectedMessageSpoilers = addShortcut("toggle_selected_message_spoilers", "Toggle spoilers in selected message",
		chatview, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	)

	DeleteSelectedMessage = addShortcut("delete_selected_message", "Delete the selected message",
		chatview, tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone),NullVimEvent),
	)

	ViewSelectedMessageImages = addShortcut("view_selected_message_images", "View selected message's attached files",
		chatview, tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),
		addVimEvent(NullVimEvent, nil, NullVimEvent),
	)

	ChatViewSelectionUp = addShortcut("selection_up", "Move selection up by one",
		chatview, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),NullVimEvent),
	)

	ChatViewSelectionDown = addShortcut("selection_down", "Move selection down by one",
		chatview, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),NullVimEvent),
	)

	ChatViewSelectionTop = addShortcut("selection_top", "Move selection to the upmost message",
		chatview, tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),NullVimEvent),
	)

	ChatViewSelectionBottom = addShortcut("selection_bottom", "Move selection to the downmost message",
		chatview, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),NullVimEvent),
	)

	// START OF INPUT
	ExpandSelectionToLeft = addShortcut("expand_selection_word_to_left", "Expand selection word to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone)),
	)

	ExpandSelectionToRight = addShortcut("expand_selection_word_to_right", "Expand selection word to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone)),
	)

	SelectAll = addShortcut("select_all", "Select all",
		multilineTextInput, tcell.NewEventKey(tcell.KeyCtrlA, rune(tcell.KeyCtrlA), tcell.ModCtrl),
		addVimEvent(NullVimEvent, NullVimEvent, tcell.NewEventKey(tcell.KeyRune, 'V', tcell.ModNone)),
	)

	SelectWordLeft = addShortcut("select_word_to_left", "Select word to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl|tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone)),
	)

	SelectWordRight = addShortcut("select_word_to_right", "Select word to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl|tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone)),
	)

	// TODO create a way to store keys making multi key commands like yw, di or y$
	SelectToStartOfLine = addShortcut("select_to_start_of_line", "Select to start of line",
		multilineTextInput, tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone)),
	)

	SelectToEndOfLine = addShortcut("select_to_end_of_line", "Select to end of line",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone)),
	)

	SelectToStartOfText = addShortcut("select_to_start_of_text", "Select to start of text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModCtrl|tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone)),
	)

	SelectToEndOfText = addShortcut("select_to_end_of_text", "Select to end of text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModCtrl|tcell.ModShift),
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone)),
	)

	MoveCursorLeft = addShortcut("move_cursor_to_left", "Move cursor to left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	)

	MoveCursorRight = addShortcut("move_cursor_to_right", "Move cursor to right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	)

	MoveCursorWordLeft = addShortcut("move_cursor_to_word_left", "Move cursor to word left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	MoveCursorWordRight = addShortcut("move_cursor_to_word_right", "Move cursor to word right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	MoveCursorStartOfLine = addShortcut("move_cursor_to_start_of_line", "Move cursor to start of line",
		multilineTextInput, tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, '0', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	MoveCursorEndOfLine = addShortcut("move_cursor_to_end_of_line", "Move cursor to end of line",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, '$', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	MoveCursorStartOfText = addShortcut("move_cursor_to_start_of_text", "Move cursor to start of text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	MoveCursorEndOfText = addShortcut("move_cursor_to_end_of_text", "Move cursor to end of text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone),NullVimEvent,NullVimEvent),
	)



	DeleteLeft  = addDeleteLeftShortcut()




	DeleteRight = addShortcut("delete_right", "Delete right",
		multilineTextInput, tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone),addVimEvent(NullVimEvent,nil,NullVimEvent))
	DeleteWordLeft = addShortcut("delete_word_left", "Delete word left",
		multilineTextInput, tcell.NewEventKey(tcell.KeyCtrlW, rune(tcell.KeyCtrlW), tcell.ModCtrl),addVimEvent(NullVimEvent,nil,NullVimEvent))
	InputNewLine = addShortcut("add_new_line_character", "Add new line character",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModAlt),addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),NullVimEvent,NullVimEvent))


	CopySelection = addShortcut("copy_selection", "Copy selected text",
		multilineTextInput, tcell.NewEventKey(tcell.KeyRune, 'C', tcell.ModAlt),
		addVimEvent(NullVimEvent,NullVimEvent, tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone)),
	)
	//Don't fix the typo, it'll break stuff ;)   ---- (glups)
	PasteAtSelection = addShortcut("paste_at_selectiom", "Paste clipboard content",
		multilineTextInput, tcell.NewEventKey(tcell.KeyCtrlV, rune(tcell.KeyCtrlV), tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'p',tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	SendMessage = addShortcut("send_message", "Sends the typed message",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone),
		addVimEvent(NullVimEvent,nil,NullVimEvent),
	)

	AddNewLineInCodeBlock = addShortcut("add_new_line_in_code_block", "Adds a new line inside a code block",
		multilineTextInput, tcell.NewEventKey(tcell.KeyEnter, rune(tcell.KeyEnter), tcell.ModNone),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	ExitApplication = addShortcut("exit_application", "Exit application",
		globalScope, tcell.NewEventKey(tcell.KeyCtrlC, rune(tcell.KeyCtrlC), tcell.ModCtrl),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusUp = addShortcut("focus_up", "Focus the next widget above",
		globalScope, tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusDown = addShortcut("focus_down", "Focus the next widget below",
		globalScope, tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusLeft = addShortcut("focus_left", "Focus the next widget to the left",
		globalScope, tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusRight = addShortcut("focus_right", "Focus the next widget to the right",
		globalScope, tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusChannelContainer = addShortcut("focus_channel_container", "Focus channel container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusUserContainer = addShortcut("focus_user_container", "Focus user container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusGuildContainer = addShortcut("focus_guild_container", "Focus guild container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusPrivateChatPage = addShortcut("focus_private_chat_page", "Focus private chat page",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'P', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	SwitchToPreviousChannel = addShortcut("switch_to_previous_channel", "Switch to previous channel",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModAlt),
		// FIXME unbound
		addVimEvent(NullVimEvent,NullVimEvent,NullVimEvent),
	)

	FocusMessageInput = addShortcut("focus_message_input", "Focus message input",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModAlt),
		// Toggle insert mode
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusMessageContainer = addShortcut("focus_message_container", "Focus message container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	FocusCommandInput = addShortcut("focus_command_input", "Focus command input",
		// FIXME unbound
		globalScope, nil, addVimEvent(NullVimEvent,NullVimEvent,NullVimEvent))
	FocusCommandOutput = addShortcut("focus_command_output", "Focus command output",
		// FIXME unbound
		globalScope, nil, addVimEvent(NullVimEvent,NullVimEvent,NullVimEvent))

	ToggleUserContainer = addShortcut("toggle_user_container", "Toggle user container",
		globalScope, tcell.NewEventKey(tcell.KeyRune, 'U', tcell.ModAlt),
		addVimEvent(nil,NullVimEvent,NullVimEvent),
	)

	ToggleCommandView = addShortcut("toggle_command_view", "Toggle command view",
		globalScope, tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt),
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, ':', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	ToggleBareChat = addShortcut("toggle_bare_chat", "Toggle bare chat",
		globalScope, tcell.NewEventKey(tcell.KeyCtrlB, rune(tcell.KeyCtrlB), tcell.ModCtrl),
		// FIXME unknown binding
		addVimEvent(NullVimEvent,NullVimEvent,NullVimEvent),
	)

	GuildListMarkRead = addShortcut("guild_mark_read", "Mark server as read",
		guildlist, tcell.NewEventKey(tcell.KeyCtrlR, rune(tcell.KeyCtrlR), tcell.ModCtrl),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone), tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone)),
	)

	ChannelTreeMarkRead = addShortcut("channel_mark_read", "Mark channel as read",
		channeltree, tcell.NewEventKey(tcell.KeyCtrlR, rune(tcell.KeyCtrlR), tcell.ModCtrl),
		addVimEvent(NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone), tcell.NewEventKey(tcell.KeyRune, 'm', tcell.ModNone)),
	)

	VimInsertMode = addShortcut("vim_insert_mode", "Change to Vim insert mode",
		globalScope, nil,
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	VimVisualMode = addShortcut("vim_visual_mode", "Change to Vim visual mode",
		globalScope, nil,
		addVimEvent(tcell.NewEventKey(tcell.KeyRune, 'v', tcell.ModNone),NullVimEvent,NullVimEvent),
	)

	VimNormalMode = addShortcut("vim_normal_mode", "Return to vim normal mode",
		globalScope, nil,
		// FIXME escape key not working in my machine. Using hyphen instead temporarily.
		addVimEvent(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)),
	)

	VimSimKeyUp = addShortcut("vim_sim_up", "Simulate an arrow key press in vim mode.",
		globalScope, nil,
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone)))

	VimSimKeyDown = addShortcut("vim_sim_down", "Simulate an arrow key press in vim mode.",
		globalScope, nil,
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone)))

	VimSimKeyLeft = addShortcut("vim_sim_left", "Simulate an arrow key press in vim mode.",
		globalScope, nil,
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone)))

	VimSimKeyRight = addShortcut("vim_sim_right", "Simulate an arrow key press in vim mode.",
		globalScope, nil,
		addVimEvent(NullVimEvent,NullVimEvent,tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone)))

	scopes    []*Scope
	Shortcuts []*Shortcut
	// VimKeyStore is needed to fix normal mode and visual mode accidentally typing inside the text box.
	VimKeyStore  map[string]string
	//				Key     Mode
)

// Vim event holds possible modifiers for keys. They can be nil. In that case, they will default
// to the original binding. If the event is associated to NullVimEvent, any events while in that mode
// will be ignored.
type VimEvent struct {
	NormalEvent *tcell.EventKey
	InsertEvent *tcell.EventKey
	VisualEvent *tcell.EventKey
}

// NullVimEvent is the null event for the current vim mode. Any events that have this will be ignored.
var NullVimEvent *tcell.EventKey = new(tcell.EventKey)

func addVimEvent(events... *tcell.EventKey) *VimEvent {
	vimE := VimEvent{NormalEvent: events[0], InsertEvent: events[1], VisualEvent: events[2]}
	return &vimE
}

// func createMask(runes rune...) int {
// 	mask := 0
// 	for _, r := range runes {
// 		mask = mask&tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
// 	}
// 	return mask
// }

func addScope(identifier, name string, parent *Scope) *Scope {
	scope := &Scope{
		Parent:     parent,
		Identifier: identifier,
		Name:       name,
	}

	scopes = append(scopes, scope)

	return scope
}

func addShortcut(identifier, name string, scope *Scope, event *tcell.EventKey, vimEvent *VimEvent) *Shortcut {
	shortcut := &Shortcut{
		Identifier:   identifier,
		Name:         name,
		Scope:        scope,
		Event:        event,
		defaultEvent: event,
		VimModifier:  vimEvent,
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

	// The Scope will be omitted, as this needed be persisted anyway.
	Scope *Scope

	// Event is the shortcut expressed as it's resulting tcell Event.
	Event *tcell.EventKey

	//This shortcuts default, in order to be able to reset it.
	defaultEvent *tcell.EventKey

	// Every shortcut receives a pointer to vimMode, therefore not
	// wasting memory
	VimStatus	*vim.Vim

	// VimModifier is the shortcut that will be used inside vim mode.
	VimModifier *VimEvent
}

// Equals compares the given EventKey with the Shortcuts Event.
// If any vim mode is enabled, it will replace the default event.
func (shortcut *Shortcut) Equals(event *tcell.EventKey) bool {
	if shortcut.VimStatus.CurrentMode == vim.NormalMode {
		selectedEvent := shortcut.VimModifier.NormalEvent
		if selectedEvent == nil {
			selectedEvent = shortcut.Event
		}
		return EventsEqual(selectedEvent, event)
	} else if shortcut.VimStatus.CurrentMode == vim.InsertMode {
		selectedEvent := shortcut.VimModifier.InsertEvent
		if selectedEvent == nil {
			selectedEvent = shortcut.Event
		}
		return EventsEqual(selectedEvent, event)
	} else if shortcut.VimStatus.CurrentMode == vim.VisualMode {
		selectedEvent := shortcut.VimModifier.VisualEvent
		if selectedEvent == nil {
			selectedEvent = shortcut.Event
		}
		return EventsEqual(selectedEvent, event)
	} else {
		return EventsEqual(shortcut.Event, event)
	}
}

// EventsEqual compares the given events, respecting everything except for the
// When field.
// Special event NullVimEvent will always return nil, because it is empty.
func EventsEqual(eventOne, eventTwo *tcell.EventKey) bool {
	if (eventOne == nil && eventTwo != nil) || (eventOne != nil && eventTwo == nil) {
		return false
	}

	return eventOne.Rune() == eventTwo.Rune() &&
		eventOne.Modifiers() == eventTwo.Modifiers() &&
		eventOne.Key() == eventTwo.Key()
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

// ShortcutDataRepresentation represents a shortcut configured by the user.
// This prevents redundancy of scopes.
type ShortcutDataRepresentation struct {
	Identifier      string
	ScopeIdentifier string
	EventKey        tcell.Key
	EventMod        tcell.ModMask
	EventRune       rune
}

// MarshalJSON marshals a Shortcut into a ShortcutDataRepresentation. This
// happens in order to prevent saving the scopes multiple times, therefore
// the scopes will be hardcoded.
func (shortcut *Shortcut) MarshalJSON() ([]byte, error) {
	if shortcut.Event == nil {
		return json.MarshalIndent(&ShortcutDataRepresentation{
			Identifier:      shortcut.Identifier,
			ScopeIdentifier: shortcut.Scope.Identifier,
			EventKey:        -1,
			EventRune:       -1,
			EventMod:        -1,
		}, "", "    ")
	}

	return json.MarshalIndent(&ShortcutDataRepresentation{
		Identifier:      shortcut.Identifier,
		ScopeIdentifier: shortcut.Scope.Identifier,
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
			shortcut.Scope = scope
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
	configDirectory, configError := config.GetConfigDirectory()
	if configError != nil {
		return "", configError
	}

	return filepath.Join(configDirectory, "shortcuts.json"), nil
}

// Load loads the shortcuts and copies the events to the correct shortcuts
// that reside inside of the memory.
func Load(vimMode *vim.Vim) error {
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
				otherShortcut.Scope.Identifier == shortcut.Scope.Identifier {
				otherShortcut.Event = shortcut.Event
				otherShortcut.VimStatus = vimMode
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

func DirectionalFocusHandling(event *tcell.EventKey, app *tview.Application) *tcell.EventKey {
	focused := app.GetFocus()

	if FocusUp.Equals(event) {
		tviewutil.FocusNextIfPossible(tview.Up, app, focused)
	} else if FocusDown.Equals(event) {
		tviewutil.FocusNextIfPossible(tview.Down, app, focused)
	} else if FocusLeft.Equals(event) {
		tviewutil.FocusNextIfPossible(tview.Left, app, focused)
	} else if FocusRight.Equals(event) {
		tviewutil.FocusNextIfPossible(tview.Right, app, focused)
	} else {
		return event
	}
	return nil
}
