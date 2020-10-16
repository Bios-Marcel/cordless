package ui

import (
	"unicode"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/femto"
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

// Editor is a simple component that wraps tview.TextView in order to give
// the user minimal text edit functionality.
type Editor struct {
	internalTextView *tview.TextView
	buffer           *femto.Buffer
	tempBuffer       *femto.Buffer

	inputCapture                    func(event *tcell.EventKey) *tcell.EventKey
	autocompleteValuesUpdateHandler func(values []*AutocompleteValue)
	autocompleters                  []*Autocomplete

	heightRequestHandler func(requestHeight int)
	requestedHeight      int
	autocompleteFrom     *femto.Loc
}

func (editor *Editor) applyBuffer() {
	editor.applyBufferWithoutAutocompletionCheck()
	if config.Current.Autocomplete {
		editor.checkForAutocompletion()
	}
}

func (editor *Editor) applyBufferWithoutAutocompletionCheck() {
	selectionStart := editor.buffer.Cursor.CurSelection[0]
	selectionEnd := editor.buffer.Cursor.CurSelection[1]

	//Copy relevant buffer-state over to temporary buffer
	editor.tempBuffer.Replace(editor.tempBuffer.Start(), editor.tempBuffer.End(), tviewutil.Escape(editor.buffer.String()))
	editor.tempBuffer.Cursor.GotoLoc(editor.buffer.Cursor.Loc)
	editor.tempBuffer.Cursor.SetSelectionStart(selectionStart)
	editor.tempBuffer.Cursor.SetSelectionEnd(selectionEnd)

	//The \x04 is a non-printable character, this is used in order to prevent
	//weird behaviour of tview in combinations that have selection and newlines
	//or whitespace.
	if editor.buffer.Cursor.HasSelection() {
		editor.tempBuffer.Insert(selectionEnd, "[\"\"]\x04")
		editor.tempBuffer.Insert(selectionStart, "[\"selection\"]\x04")
	} else {
		if editor.tempBuffer.Cursor.RuneUnder(editor.tempBuffer.Cursor.Loc.X) == '\n' {
			editor.tempBuffer.Insert(editor.tempBuffer.Cursor.Loc, "[\"selection\"]\x04 [\"\"]\x04")
		} else {
			editor.tempBuffer.Insert(editor.tempBuffer.Cursor.Loc.Move(1, editor.tempBuffer), "[\"\"]\x04")
			editor.tempBuffer.Insert(editor.tempBuffer.Cursor.Loc, "[\"selection\"]\x04")
		}
	}

	editor.internalTextView.SetText(editor.tempBuffer.String())
}

func (editor *Editor) checkForAutocompletion() {
	if editor.autocompleteValuesUpdateHandler != nil {
		cursorLoc := editor.buffer.Cursor.Loc
		var spaceFound bool
		for {
			cursorLoc.X = cursorLoc.X - 1
			if cursorLoc.X < 0 {
				break
			}

			runeAtCursor := editor.buffer.RuneAt(cursorLoc)
			if runeAtCursor == ' ' {
				spaceFound = true
			}

			for _, value := range editor.autocompleters {
				if value.firstRune == runeAtCursor {
					if spaceFound && !value.allowSpaces {
						break
					}

					cursorLocCopy := cursorLoc
					cursorLocCopy.X--

					if cursorLocCopy.X >= 0 && !unicode.IsSpace(editor.buffer.RuneAt(cursorLocCopy)) {
						break
					}

					editor.autocompleteFrom = &cursorLoc
					//We don't want the autocomplete character to be part of the search value
					cursorLocCopy = cursorLoc
					cursorLocCopy.X++
					editor.autocompleteValuesUpdateHandler(value.valueSupplier(
						editor.buffer.Substr(cursorLocCopy, editor.buffer.Cursor.Loc)))
					return
				}

			}
		}

		editor.autocompleteFrom = nil
		editor.autocompleteValuesUpdateHandler(nil)
	}
}

// MoveCursorLeft moves the cursor left by one cell.
func (editor *Editor) MoveCursorLeft() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[0])
		editor.buffer.Cursor.ResetSelection()
	} else {
		editor.buffer.Cursor.Left()
	}

	editor.applyBuffer()
}

// MoveCursorRight moves the cursor right by one cell.
func (editor *Editor) MoveCursorRight() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[1])
		editor.buffer.Cursor.ResetSelection()
	} else {
		editor.buffer.Cursor.Right()
	}

	editor.applyBuffer()
}

// SelectionToLeft extends the selection one cell to the left.
func (editor *Editor) SelectionToLeft() {
	editor.selectLeft(false)
}

// SelectWordLeft extends the selection one word to the left. A word may
// span multiple words. A word however can be one cell and mustn't be a word
// in terms of human language definition.
func (editor *Editor) SelectWordLeft() {
	editor.selectLeft(true)
}

func (editor *Editor) selectLeft(word bool) {
	oldCursor := editor.buffer.Cursor.Loc
	selectionStart := editor.buffer.Cursor.CurSelection[0]
	selectionEnd := editor.buffer.Cursor.CurSelection[1]

	if word {
		editor.buffer.Cursor.WordLeft()
	} else {
		editor.buffer.Cursor.Left()
	}
	newCursor := editor.buffer.Cursor.Loc
	if !editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.SetSelectionStart(newCursor)
		editor.buffer.Cursor.SetSelectionEnd(oldCursor)
	} else if oldCursor.GreaterEqual(selectionStart) {
		if newCursor.GreaterEqual(selectionStart) {
			editor.buffer.Cursor.SetSelectionStart(selectionStart)
			editor.buffer.Cursor.SetSelectionEnd(newCursor)
		} else {
			editor.buffer.Cursor.SetSelectionStart(newCursor)
			editor.buffer.Cursor.SetSelectionEnd(selectionEnd)
		}
	} else {
		editor.buffer.Cursor.SetSelectionStart(newCursor)
		editor.buffer.Cursor.SetSelectionEnd(selectionEnd)
	}

	editor.applyBuffer()
}

// SelectionToRight extends the selection one cell to the right.
func (editor *Editor) SelectionToRight() {
	editor.selectRight(false)
}

// SelectWordRight extends the selection one word to the right. A word may
// span multiple words. A word however can be one cell and mustn't be a word
// in terms of human language definition.
func (editor *Editor) SelectWordRight() {
	editor.selectRight(true)
}

func (editor *Editor) selectRight(word bool) {
	oldCursor := editor.buffer.Cursor.Loc
	selectionStart := editor.buffer.Cursor.CurSelection[0]
	selectionEnd := editor.buffer.Cursor.CurSelection[1]

	if word {
		editor.buffer.Cursor.WordRight()
	} else {
		editor.buffer.Cursor.Right()
	}
	newCursor := editor.buffer.Cursor.Loc
	if !editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.SetSelectionStart(oldCursor)
		editor.buffer.Cursor.SetSelectionEnd(newCursor)
	} else if newCursor.LessThan(selectionEnd) {
		editor.buffer.Cursor.SetSelectionStart(newCursor)
		editor.buffer.Cursor.SetSelectionEnd(selectionEnd)
	} else {
		editor.buffer.Cursor.SetSelectionStart(selectionStart)
		editor.buffer.Cursor.SetSelectionEnd(newCursor)
	}

	editor.applyBuffer()
}

// SelectAll selects all text (cells) currently filled. If no text is
// available, nothing will change.
func (editor *Editor) SelectAll() {
	start := editor.buffer.Start()
	editor.buffer.Cursor.SetSelectionStart(start)
	end := editor.buffer.End()
	editor.buffer.Cursor.SetSelectionEnd(end)
	editor.buffer.Cursor.GotoLoc(end)
	editor.applyBuffer()
}

// SelectToStartOfLine will select all text to the left til the next newline
// is found. Lines doesn't mean "editor line" in this context, as the editor
// doesn't currently support vertical navigation.
func (editor *Editor) SelectToStartOfLine() {
	oldCursor := editor.buffer.Cursor.Loc
	editor.buffer.Cursor.StartOfText()
	newCursor := editor.buffer.Cursor.Loc
	if !oldCursor.GreaterThan(newCursor) {
		editor.buffer.Cursor.Start()
		newCursor = editor.buffer.Cursor.Loc
	}
	editor.buffer.Cursor.SetSelectionStart(newCursor)
	editor.buffer.Cursor.SetSelectionEnd(oldCursor)
	editor.applyBuffer()
}

// SelectToEndOfLine will select all text to the right til the next newline
// is found. Lines doesn't mean "editor line" in this context, as the editor
// doesn't currently support vertical navigation.
func (editor *Editor) SelectToEndOfLine() {
	oldCursor := editor.buffer.Cursor.Loc
	editor.buffer.Cursor.End()
	editor.buffer.Cursor.SetSelectionStart(oldCursor)
	editor.buffer.Cursor.SetSelectionEnd(editor.buffer.Cursor.Loc)
	editor.applyBuffer()
}

// SelectToStartOfText will select all text to the start of the editor.
// Meaning the top-left most cell.
func (editor *Editor) SelectToStartOfText() {
	oldCursor := editor.buffer.Cursor.Loc
	textStart := editor.buffer.Start()
	editor.buffer.Cursor.GotoLoc(textStart)
	editor.buffer.Cursor.SetSelectionStart(textStart)
	editor.buffer.Cursor.SetSelectionEnd(oldCursor)
	editor.applyBuffer()
}

// SelectToEndOfText will select all text to the end of the editor.
// Meaning the bottom-right most cell.
func (editor *Editor) SelectToEndOfText() {
	oldCursor := editor.buffer.Cursor.Loc
	textEnd := editor.buffer.End()
	editor.buffer.Cursor.GotoLoc(textEnd)
	editor.buffer.Cursor.SetSelectionStart(oldCursor)
	editor.buffer.Cursor.SetSelectionEnd(textEnd)
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorWordLeft() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[0])
	}
	editor.buffer.Cursor.WordLeft()
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorWordRight() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[1])
	}
	editor.buffer.Cursor.WordRight()
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorStartOfLine() {
	oldCursor := editor.buffer.Cursor.Loc
	editor.buffer.Cursor.StartOfText()
	if !oldCursor.GreaterThan(editor.buffer.Cursor.Loc) {
		editor.buffer.Cursor.Start()
	}
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorEndOfLine() {
	editor.buffer.Cursor.End()
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorStartOfText() {
	editor.buffer.Cursor.GotoLoc(editor.buffer.Start())
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorEndOfText() {
	editor.buffer.Cursor.GotoLoc(editor.buffer.End())
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) Backspace() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Remove(editor.buffer.Cursor.CurSelection[0], editor.buffer.Cursor.CurSelection[1])
		editor.applyBuffer()
	} else if editor.buffer.Cursor.Loc.X > 0 || editor.buffer.Cursor.Loc.Y > 0 {
		editor.buffer.Remove(editor.buffer.Cursor.Loc.Move(-1, editor.buffer), editor.buffer.Cursor.Loc)
		editor.applyBuffer()
	}
}

func (editor *Editor) DeleteWordLeft() {
	if editor.buffer.Cursor.HasSelection() {
		editor.Backspace()
	} else {
		oldLocation := editor.buffer.Cursor.Loc
		editor.buffer.Cursor.WordLeft()
		newLocation := editor.buffer.Cursor.Loc

		if oldLocation.X != newLocation.X || oldLocation.Y != newLocation.Y {
			editor.buffer.Cursor.SetSelectionStart(newLocation)
			editor.buffer.Cursor.SetSelectionEnd(oldLocation)
			editor.Backspace()
		}
	}
}

func (editor *Editor) DeleteRight() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Remove(editor.buffer.Cursor.CurSelection[0], editor.buffer.Cursor.CurSelection[1])
	} else {
		editor.buffer.Remove(editor.buffer.Cursor.Loc, editor.buffer.Cursor.Loc.Move(1, editor.buffer))
	}

	editor.applyBuffer()
}

func (editor *Editor) Paste(event *tcell.EventKey) {
	if editor.inputCapture != nil {
		result := editor.inputCapture(event)
		if result == nil {
			//Early exit, as even has been handled.
			return
		}
	}

	clipBoardContent, clipError := clipboard.ReadAll()
	if clipError == nil {
		if editor.buffer.Cursor.HasSelection() {
			editor.buffer.Replace(editor.buffer.Cursor.CurSelection[0], editor.buffer.Cursor.CurSelection[1], clipBoardContent)
		} else {
			editor.buffer.Insert(editor.buffer.Cursor.Loc, clipBoardContent)
		}
		editor.applyBuffer()
	}
}

func (editor *Editor) InsertCharacter(character rune) {
	selectionEnd := editor.buffer.Cursor.CurSelection[1]
	selectionStart := editor.buffer.Cursor.CurSelection[0]
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Replace(selectionStart, selectionEnd, string(character))
	} else {
		editor.buffer.Insert(editor.buffer.Cursor.Loc, string(character))
	}
	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

// NewEditor instantiates a ready to use text editor.
func NewEditor() *Editor {
	editor := Editor{
		internalTextView: tview.NewTextView(),
		requestedHeight:  3,
		buffer:           femto.NewBufferFromString("", ""),
		tempBuffer:       femto.NewBufferFromString("", ""),
	}

	editor.internalTextView.SetWrap(true)
	editor.internalTextView.SetWordWrap(true)
	editor.internalTextView.SetBorder(true)
	editor.internalTextView.SetRegions(true)
	editor.internalTextView.SetScrollable(true)
	editor.internalTextView.Highlight("selection")
	editor.applyBuffer()

	editor.buffer.Cursor.SetSelectionStart(editor.buffer.Start())
	editor.buffer.Cursor.SetSelectionEnd(editor.buffer.End())

	editor.internalTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		inputCapture := editor.inputCapture
		if inputCapture != nil {
			event = inputCapture(event)
			if event == nil {
				return nil
			}
		}

		if shortcuts.MoveCursorLeft.Equals(event) {
			editor.MoveCursorLeft()
		} else if shortcuts.ExpandSelectionToLeft.Equals(event) {
			editor.SelectionToLeft()
		} else if shortcuts.MoveCursorRight.Equals(event) {
			editor.MoveCursorRight()
		} else if shortcuts.ExpandSelectionToRight.Equals(event) {
			editor.SelectionToRight()
		} else if shortcuts.SelectWordLeft.Equals(event) {
			editor.SelectWordLeft()
		} else if shortcuts.SelectWordRight.Equals(event) {
			editor.SelectWordRight()
		} else if shortcuts.SelectToStartOfLine.Equals(event) {
			editor.SelectToStartOfLine()
		} else if shortcuts.SelectToEndOfLine.Equals(event) {
			editor.SelectToEndOfLine()
		} else if shortcuts.SelectToStartOfText.Equals(event) {
			editor.SelectToStartOfText()
		} else if shortcuts.SelectToEndOfText.Equals(event) {
			editor.SelectToEndOfText()
		} else if shortcuts.SelectAll.Equals(event) {
			editor.SelectAll()
		} else if shortcuts.MoveCursorWordLeft.Equals(event) {
			editor.MoveCursorWordLeft()
		} else if shortcuts.MoveCursorWordRight.Equals(event) {
			editor.MoveCursorWordRight()
		} else if shortcuts.MoveCursorStartOfLine.Equals(event) {
			editor.MoveCursorStartOfLine()
		} else if shortcuts.MoveCursorEndOfLine.Equals(event) {
			editor.MoveCursorEndOfLine()
		} else if shortcuts.MoveCursorStartOfText.Equals(event) {
			editor.MoveCursorStartOfText()
		} else if shortcuts.MoveCursorEndOfText.Equals(event) {
			editor.MoveCursorEndOfText()
		} else if shortcuts.DeleteRight.Equals(event) {
			editor.DeleteRight()
		} else if event.Key() == tcell.KeyBackspace2 ||
			event.Key() == tcell.KeyBackspace {
			// FIXME Legacy, has to be replaced when there is N-1 Keybind-Mapping.
			editor.Backspace()
		} else if shortcuts.DeleteWordLeft.Equals(event) {
			editor.DeleteWordLeft()
		} else if shortcuts.CopySelection.Equals(event) {
			clipboard.WriteAll(editor.buffer.Cursor.GetSelection())
			//Returning nil, as copying won't do anything than filling the
			//clipboard buffer.
			return nil
		} else if shortcuts.PasteAtSelection.Equals(event) {
			editor.Paste(event)
			editor.TriggerHeightRequestIfNecessary()
			return nil
		} else if shortcuts.InputNewLine.Equals(event) {
			editor.InsertCharacter('\n')
		} else if event.Rune() != 0 {
			editor.InsertCharacter(event.Rune())
		}

		editor.TriggerHeightRequestIfNecessary()
		editor.internalTextView.ScrollToHighlight()
		return nil
	})
	return &editor
}

func (editor *Editor) GetTextLeftOfSelection() string {
	var to femto.Loc
	if editor.buffer.Cursor.HasSelection() {
		to = editor.buffer.Cursor.CurSelection[1]
	} else {
		to = editor.buffer.End()
	}
	return editor.buffer.Substr(editor.buffer.Start(), to)
}

func (editor *Editor) FindAtSymbolIndexInCurrentWord() int {
	newLeft := editor.GetTextLeftOfSelection()
	for i := len(newLeft) - 1; i >= 0; i-- {
		if newLeft[i] == '@' && (i == 0 || newLeft[i-1] == ' ' || newLeft[i-1] == '\n') {
			return i
		}
	}
	return -1
}

func (editor *Editor) countRows(text string) int {
	_, _, width, _ := editor.internalTextView.GetInnerRect()
	return tviewutil.CalculateNecessaryHeight(width, text)
}

// TriggerHeightRequestIfNecessary informs the parent that more or less height
// is requierd for rendering than currently in use, unless the height is already
// optimal.
func (editor *Editor) TriggerHeightRequestIfNecessary() {
	if editor.heightRequestHandler == nil {
		return
	}

	rowAmount := editor.countRows(editor.GetText())

	newRequestedHeight := rowAmount
	if editor.internalTextView.IsBorderTop() {
		newRequestedHeight++
	}
	if editor.internalTextView.IsBorderBottom() {
		newRequestedHeight++
	}

	if newRequestedHeight != editor.requestedHeight {
		editor.requestedHeight = newRequestedHeight
		editor.heightRequestHandler(newRequestedHeight)
	}
}

type AutocompleteValue struct {
	RenderValue string
	InsertValue string
}

type Autocomplete struct {
	firstRune     rune
	allowSpaces   bool
	valueSupplier func(string) []*AutocompleteValue
}

func (editor *Editor) RegisterAutocomplete(firstRune rune, allowSpaces bool, valueSupplier func(string) []*AutocompleteValue) {
	editor.autocompleters = append(editor.autocompleters, &Autocomplete{
		firstRune:     firstRune,
		allowSpaces:   allowSpaces,
		valueSupplier: valueSupplier,
	})
}

func (editor *Editor) Autocomplete(value string) {
	if editor.autocompleteFrom != nil {
		editor.buffer.Replace(*editor.autocompleteFrom, editor.buffer.Cursor.Loc, value+" ")
		editor.autocompleteFrom = nil
		//Not necessary, since you probably don't want to autocomplete any
		//further after you've chosen a value.
		editor.applyBufferWithoutAutocompletionCheck()
	}
}

// GetRequestedHeight returns the currently requested size.
func (editor *Editor) GetRequestedHeight() int {
	return editor.requestedHeight
}

// SetOnHeightChangeRequest handles the cases where the component thinks it needs
// more space or would be fine with less.
func (editor *Editor) SetOnHeightChangeRequest(handler func(requestHeight int)) {
	editor.heightRequestHandler = handler
}

// SetBackgroundColor sets the background color of the internal TextView
func (editor *Editor) SetBackgroundColor(color tcell.Color) {
	editor.internalTextView.SetBackgroundColor(color)
}

// SetText sets the texts of the internal TextView, but also sets the selection
// and necessary groups for the navigation behaviour.
func (editor *Editor) SetText(text string) {
	if text == "" {
		editor.buffer.Remove(editor.buffer.Start(), editor.buffer.End())
	} else {
		editor.buffer.Replace(editor.buffer.Start(), editor.buffer.End(), text)
	}
	editor.buffer.Cursor.ResetSelection()
	editor.buffer.Cursor.GotoLoc(editor.buffer.End())
	editor.applyBuffer()
	editor.TriggerHeightRequestIfNecessary()
}

// SetBorderFocusColor delegates to the underlying components
// SetBorderFocusColor method.
func (editor *Editor) SetBorderFocusColor(color tcell.Color) {
	editor.internalTextView.SetBorderFocusColor(color)
}

// SetBorderColor delegates to the underlying components SetBorderColor
// method.
func (editor *Editor) SetBorderColor(color tcell.Color) {
	editor.internalTextView.SetBorderColor(color)
}

// SetBorderAttributes delegates to the underlying components SetBorderAttributes
// method.
func (editor *Editor) SetBorderAttributes(attr tcell.AttrMask) {
	editor.internalTextView.SetBorderAttributes(attr)
}

// SetBorderFocusAttributes delegates to the underlying components SetBorderFocusAttributes
// method.
func (editor *Editor) SetBorderFocusAttributes(attr tcell.AttrMask) {
	editor.internalTextView.SetBorderFocusAttributes(attr)
}

// SetInputCapture sets the alternative input capture that will be used if the
// components default controls aren't being triggered.
func (editor *Editor) SetInputCapture(captureFunc func(event *tcell.EventKey) *tcell.EventKey) {
	editor.inputCapture = captureFunc
}

func (editor *Editor) SetAutocompleteValuesUpdateHandler(handlerFunc func(values []*AutocompleteValue)) {
	editor.autocompleteValuesUpdateHandler = handlerFunc
}

// GetText returns the text without color tags, region tags and so on.
func (editor *Editor) GetText() string {
	return editor.buffer.String()
}

// GetPrimitive returns the internal component that can be added to a layout
func (editor *Editor) GetPrimitive() tview.Primitive {
	return editor.internalTextView
}
