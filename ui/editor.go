package ui

import (
	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/femto"
	"github.com/Bios-Marcel/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
)

// Editor is a simple component that wraps tview.TextView in order to gove the
// user minimal text edit functionality.
type Editor struct {
	internalTextView *tview.TextView
	buffer           *femto.Buffer
	tempBuffer       *femto.Buffer

	inputCapture             func(event *tcell.EventKey) *tcell.EventKey
	mentionShowHandler       func(namePart string)
	mentionHideHandler       func()
	heightRequestHandler     func(requestHeight int)
	requestedHeight          int
	currentMentionBeginIndex int
	currentMentionEndIndex   int
}

func (editor *Editor) applyBuffer() {
	bufferToString := editor.buffer.String()
	selectionStart := editor.buffer.Cursor.CurSelection[0]
	selectionEnd := editor.buffer.Cursor.CurSelection[1]

	editor.tempBuffer.Replace(editor.tempBuffer.Start(), editor.tempBuffer.End(), bufferToString)
	editor.tempBuffer.Cursor.GotoLoc(editor.buffer.Cursor.Loc)
	editor.tempBuffer.Cursor.SetSelectionStart(selectionStart)
	editor.tempBuffer.Cursor.SetSelectionEnd(selectionEnd)

	if editor.buffer.Cursor.HasSelection() {
		editor.tempBuffer.Insert(selectionEnd, "[\"\"]")
		editor.tempBuffer.Insert(selectionStart, "[\"selection\"]")
	} else {
		cursorToRight := editor.tempBuffer.Cursor.Loc.Move(1, editor.tempBuffer)
		cursorText := editor.buffer.Substr(editor.tempBuffer.Cursor.Loc, cursorToRight)
		//Little workaround, since tviews textview actually trims whitespace on
		//newline, therefore selection would be gone
		if cursorText == "\n" || cursorText == "" {
			editor.tempBuffer.Insert(editor.tempBuffer.Cursor.Loc, "[\"selection\"] [\"\"]\x04")
		} else {
			editor.tempBuffer.Insert(cursorToRight, "[\"\"]")
			editor.tempBuffer.Insert(editor.tempBuffer.Cursor.Loc, "[\"selection\"]")
		}
	}

	editor.internalTextView.SetText(editor.tempBuffer.String())
}

func (editor *Editor) MoveCursorLeft() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[0])
	} else {
		editor.buffer.Cursor.Left()
	}

	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) MoveCursorRight() {
	if editor.buffer.Cursor.HasSelection() {
		editor.buffer.Cursor.GotoLoc(editor.buffer.Cursor.CurSelection[1])
	} else {
		editor.buffer.Cursor.Right()
	}

	editor.buffer.Cursor.ResetSelection()
	editor.applyBuffer()
}

func (editor *Editor) SelectionToLeft() {
	editor.selectLeft(false)
}

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

func (editor *Editor) SelectionToRight() {
	editor.selectRight(false)
}

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

func (editor *Editor) SelectAll() {
	start := editor.buffer.Start()
	editor.buffer.Cursor.SetSelectionStart(start)
	end := editor.buffer.End()
	editor.buffer.Cursor.SetSelectionEnd(end)
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
		// TODO: This entire chunk could be cleaned up by assigning handlers to each event type,
		// e.g. event.trigger()
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
		} else if shortcuts.MoveCursorWordLeft.Equals(event) {
			editor.MoveCursorWordLeft()
		} else if shortcuts.MoveCursorWordRight.Equals(event) {
			editor.MoveCursorWordRight()
		} else if shortcuts.SelectAll.Equals(event) {
			editor.SelectAll()
		} else if shortcuts.DeleteRight.Equals(event) {
			editor.DeleteRight()
		} else if event.Key() == tcell.KeyBackspace2 ||
			event.Key() == tcell.KeyBackspace {
			// FIXME Legacy, has to be replaced when there is N-1 Keybind-Mapping.
			editor.Backspace()
		} else if shortcuts.CopySelection.Equals(event) {
			clipboard.WriteAll(editor.buffer.Cursor.GetSelection())
			//Returning nil, as copying won't do anything than filling the
			//clipboard buffer.
			return nil
		} else if shortcuts.PasteAtSelection.Equals(event) {
			editor.Paste(event)
			return nil
		} else if shortcuts.InputNewLine.Equals(event) {
			editor.InsertCharacter('\n')
		} else if shortcuts.SendMessage.Equals(event) && editor.inputCapture != nil {
			return editor.inputCapture(event)
		} else if (editor.inputCapture == nil || editor.inputCapture(event) != nil) && event.Rune() != 0 {
			editor.InsertCharacter(event.Rune())
		} else {
			return event
		}

		editor.UpdateMentionHandler()
		editor.triggerHeightRequestIfNecessary()
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

func (editor *Editor) UpdateMentionHandler() {
	atSymbolIndex := editor.FindAtSymbolIndexInCurrentWord()
	if atSymbolIndex == -1 {
		editor.HideAndResetMentionHandler()
	} else {
		editor.ShowMentionHandler(atSymbolIndex)
	}
}

func (editor *Editor) ShowMentionHandler(atSymbolIndex int) {
	text := editor.GetTextLeftOfSelection()
	lookupKeyword := text[atSymbolIndex+1:]
	editor.currentMentionBeginIndex = atSymbolIndex + 1
	editor.currentMentionEndIndex = len(lookupKeyword) + atSymbolIndex
	if editor.mentionShowHandler != nil {
		editor.mentionShowHandler(lookupKeyword)
	}
}

func (editor *Editor) HideAndResetMentionHandler() {
	editor.currentMentionBeginIndex = 0
	editor.currentMentionEndIndex = 0
	if editor.mentionHideHandler != nil {
		editor.mentionHideHandler()
	}
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
	return tviewutil.CalculateNeccessaryHeight(width, text)
}

func (editor *Editor) triggerHeightRequestIfNecessary() {
	if editor.heightRequestHandler == nil {
		return
	}

	rowAmount := editor.countRows(editor.GetText())

	newRequestedHeight := rowAmount + 2 /*borders*/
	if newRequestedHeight != editor.requestedHeight {
		editor.requestedHeight = newRequestedHeight
		editor.heightRequestHandler(newRequestedHeight)
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
	editor.buffer.Replace(editor.buffer.Start(), editor.buffer.End(), text)
	editor.applyBuffer()
	editor.triggerHeightRequestIfNecessary()
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

// SetInputCapture sets the alternative input capture that will be used if the
// components default controls aren't being triggered.
func (editor *Editor) SetInputCapture(captureFunc func(event *tcell.EventKey) *tcell.EventKey) {
	editor.inputCapture = captureFunc
}

// SetMentionShowHandler sets the handler for when a mention is being requested
func (editor *Editor) SetMentionShowHandler(handlerFunc func(namePart string)) {
	editor.mentionShowHandler = handlerFunc
}

// SetMentionHideHandler sets the handler for when a mention is no longer being requested
func (editor *Editor) SetMentionHideHandler(handlerFunc func()) {
	editor.mentionHideHandler = handlerFunc
}

// GetCurrentMentionIndices gets the starting and ending indices of the input box text
// which are to be replaced
func (editor *Editor) GetCurrentMentionIndices() (int, int) {
	return editor.currentMentionBeginIndex, editor.currentMentionEndIndex
}

// GetText returns the text without color tags, region tags and so on.
func (editor *Editor) GetText() string {
	return editor.buffer.String()
}

// GetPrimitive returnbs the internal component that can be added to a layout
func (editor *Editor) GetPrimitive() tview.Primitive {
	return editor.internalTextView
}
