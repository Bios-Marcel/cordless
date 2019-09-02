package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"

	"github.com/Bios-Marcel/tview"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
)

var (
	// Temporary solution, so not every function has to handle the selection
	// character placement.
	multiSelectionCharWithSelectionToLeftPattern = regexp.MustCompile(selectionChar + "*" + regexp.QuoteMeta(selRegion) + selectionChar + "*" + regexp.QuoteMeta(endRegion))
)

const (
	selectionChar = string('\u205F')
	emptyText     = "[\"selection\"]\u205F[\"\"]"
	leftRegion    = "[\"left\"]"
	rightRegion   = "[\"right\"]"
	selRegion     = "[\"selection\"]"
	endRegion     = "[\"\"]"
)

// Editor is a simple component that wraps tview.TextView in order to gove the
// user minimal text edit functionality.
type Editor struct {
	internalTextView *tview.TextView

	inputCapture             func(event *tcell.EventKey) *tcell.EventKey
	mentionShowHandler       func(namePart string)
	mentionHideHandler       func()
	heightRequestHandler     func(requestHeight int)
	requestedHeight          int
	currentMentionBeginIndex int
	currentMentionEndIndex   int
}

func (e *Editor) ExpandSelectionToLeft(left, right, selection []rune) {
	if len(left) > 0 {
		newText := leftRegion + string(left[:len(left)-1]) + selRegion

		currentSelection := string(selection)
		if currentSelection == selectionChar {
			currentSelection = ""
		}

		newText = newText + string(left[len(left)-1]) + currentSelection + rightRegion + string(right) + endRegion
		e.internalTextView.SetText(newText)
	}
}

func (e *Editor) ExpandSelectionToRight(left, right, selection []rune) {
	newText := leftRegion + string(left)
	if len(right) > 0 {
		newText = newText + selRegion + string(selection) + string(right[0]) + rightRegion + string(right[1:])
	} else {
		endsWithSelectionChar := strings.HasSuffix(string(selection), selectionChar)
		newText = newText + selRegion + string(selection)
		if !endsWithSelectionChar {
			newText = newText + selectionChar
		}
	}

	newText = newText + endRegion
	e.setAndFixText(newText)
}

func (e *Editor) MoveCursorLeft(left, right, selection []rune) {
	var newText string
	if len(left) > 0 {
		newText = leftRegion + string(left[:len(left)-1]) + selRegion

		currentSelection := string(selection)
		if currentSelection == selectionChar {
			currentSelection = ""
		}

		newText = newText + string(left[len(left)-1]) + rightRegion + currentSelection + string(right) + endRegion

		e.internalTextView.SetText(newText)
	} else if len(selection) > 0 {
		if len(right) > 0 {
			newText = selRegion + string(selection[0]) + rightRegion + string(selection[1:]) + string(right) + endRegion
		} else {
			newText = selRegion + string(selection[0]) + rightRegion + string(selection[1:]) + endRegion
		}
		e.setAndFixText(newText)
	}
}

func (e *Editor) MoveCursorRight(left, right, selection []rune) {
	newText := leftRegion + string(left)
	if len(right) > 0 {
		newText = newText + string(selection) + selRegion + string(right[0]) + rightRegion + string(right[1:])
	} else {
		endsWithSelectionChar := strings.HasSuffix(string(selection), selectionChar)
		if !endsWithSelectionChar {
			newText = newText + string(selection) + selRegion + selectionChar
		} else {
			newText = newText + string(selection[:len(selection)-1]) + selRegion + selectionChar
		}
	}

	newText = newText + endRegion
	e.setAndFixText(newText)
}

func (e *Editor) MoveCursorToIndex(left, right, selection []rune, index int) {
	var newText string = string(left) + string(selection) + string(right)
	if index < 0 {
		index = 0
	} else if index >= len(newText) {
		index = len(newText) - 1
	}

	if index < len(left) {
		newText = leftRegion + string(left[:index]) + selRegion + string(left[index]) + rightRegion + string(left[index+1:]) + string(right) + endRegion
	} else {
		indexSelection := index - len(left)
		if indexSelection < len(selection) {
			newText = leftRegion + string(left) + string(left[:indexSelection]) + selRegion + string(selection[indexSelection]) + rightRegion + string(selection[indexSelection+1:]) + string(right) + endRegion
		} else {
			indexRight := index - len(left) - len(selection)
			if indexRight < len(right) {
				newText = leftRegion + string(left) + string(selection) + string(right[:indexRight]) + selRegion + string(right[indexRight]) + rightRegion + string(right[indexRight+1:]) + endRegion
			}
		}
	}

	e.setAndFixText(newText)
}

func (e *Editor) SelectWordLeft(left, right, selection []rune) {
	if len(left) > 0 {
		selectionFrom := 0
		for i := len(left) - 2; /*Skip space left to selection*/ i >= 0; i-- {
			if left[i] == ' ' || left[i] == '\n' {
				selectionFrom = i
				break
			}
		}

		var newText string
		if selectionFrom != 0 {
			newText = leftRegion + string(left[:selectionFrom+1]) + selRegion + string(left[selectionFrom+1:]) + string(string(selection)) + rightRegion + string(right) + endRegion
		} else {
			newText = selRegion + string(left) + string(string(selection)) + rightRegion + string(right) + endRegion
		}
		e.setAndFixText(newText)
	}
}

func (e *Editor) SelectWordRight(left, right, selection []rune) {
	if len(right) > 0 {
		selectionFrom := len(right) - 1
		for i := 1; /*Skip space right to selection*/ i < len(right)-1; i++ {
			if right[i] == ' ' || right[i] == '\n' {
				selectionFrom = i
				break
			}
		}

		var newText string
		if selectionFrom != len(right)-1 {
			newText = leftRegion + string(left) + selRegion + string(string(selection)) + string(right[:selectionFrom]) + rightRegion + string(right[selectionFrom:]) + endRegion
		} else {
			newText = leftRegion + string(left) + selRegion + string(string(selection)) + string(right) + endRegion
		}
		e.setAndFixText(newText)
	}
}

func (e *Editor) MoveCursorWordLeft(left, right, selection []rune) {
	if len(left) > 0 {
		selectionAt := 0
		for i := len(left) - 2; /*Skip space left to selection*/ i >= 0; i-- {
			if left[i] == ' ' || left[i] == '\n' {
				selectionAt = i
				break
			}
		}

		var newText string
		if selectionAt != 0 {
			newText = leftRegion + string(left[:selectionAt]) + selRegion + string(left[selectionAt]) + rightRegion + string(left[selectionAt+1:]) + string(string(selection)) + string(right) + endRegion
		} else {
			if len(left) > 1 {
				newText = selRegion + string(left[0]) + rightRegion + string(left[1:]) + string(selection) + string(right) + endRegion
			} else {
				newText = selRegion + string(left[0]) + rightRegion + string(selection) + string(right) + endRegion
			}
		}
		e.setAndFixText(newText)
	}
}

func (e *Editor) MoveCursorWordRight(left, right, selection []rune) {
	if len(right) > 0 {
		selectionAt := len(right) - 1
		for i := 1; /*Skip space right to selection*/ i < len(right)-1; i++ {
			if right[i] == ' ' || right[i] == '\n' {
				selectionAt = i
				break
			}
		}

		var newText string
		if selectionAt != len(right)-1 {
			newText = leftRegion + string(left) + string(string(selection)) + string(right[:selectionAt]) + selRegion + string(right[selectionAt]) + rightRegion + string(right[selectionAt+1:]) + endRegion
		} else {
			newText = leftRegion + string(left) + string(selection) + string(right) + selRegion + selectionChar + endRegion
		}
		e.setAndFixText(newText)
	}
}

func (e *Editor) SelectAll(left, right, selection []rune) {
	if len(left) > 0 || len(right) > 0 {
		e.setAndFixText(selRegion + string(left) + string(selection) + string(right) + endRegion)
	}
}

func (e *Editor) DeleteRight(left, right, selection []rune) {
	if len(selection) >= 1 && strings.HasSuffix(string(selection), selectionChar) {
		e.setAndFixText(leftRegion + string(left) + selRegion + selectionChar + endRegion)
	} else if string(selection) != selectionChar {
		var newText string
		newText = leftRegion + string(left) + selRegion
		if len(right) == 0 {
			newText = newText + selectionChar
		} else {
			newText = newText + string(right[0])
		}

		if len(right) > 1 {
			newText = newText + rightRegion + string(right[1:])
		}

		newText = newText + endRegion
		e.setAndFixText(newText)
	}
}

func (e *Editor) Paste(left, right, selection []rune, event *tcell.EventKey) {
	if e.inputCapture != nil {
		result := e.inputCapture(event)
		if result == nil {
			//Early exit, as even has been handled.
			return
		}
	}

	clipBoardContent, clipError := clipboard.ReadAll()
	if clipError == nil {
		var newText string
		if string(selection) == selectionChar {
			newText = leftRegion + string(left) + clipBoardContent + selRegion + string(selection)
		} else {
			newText = leftRegion + string(left) + clipBoardContent
			if len(selection) == 1 {
				newText = newText + selRegion + string(selection) + rightRegion + string(right)
			} else {
				newText = newText + selRegion
				if len(right) == 0 {
					newText = newText + selectionChar
				} else if len(right) == 0 {
					newText = newText + string(right[0])
				} else {
					newText = newText + string(right[0]) + rightRegion + string(right[1:])
				}
			}
		}
		e.setAndFixText(newText + endRegion)
		e.triggerHeightRequestIfNeccessary()
	}
}

func (e *Editor) InsertCharacter(left, right, selection []rune, character rune) {
	if len(right) == 0 {
		if len(selection) == 1 {
			if string(selection) == selectionChar {
				e.setAndFixText(fmt.Sprintf("[\"left\"]%s%c[\"\"][\"selection\"]%s[\"\"]", string(left), character, string(selectionChar)))
			} else {
				e.setAndFixText(fmt.Sprintf("[\"left\"]%s%c[\"\"][\"selection\"]%s[\"\"]", string(left), character, string(selection)))
			}
		} else {
			e.setAndFixText(fmt.Sprintf("[\"left\"]%s%c[\"\"][\"selection\"]%s[\"\"]", string(left), character, string(selectionChar)))
		}
	} else {
		if len(selection) == 1 {
			e.setAndFixText(fmt.Sprintf("[\"left\"]%s%c[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
				string(left), character, string(selection), string(right)))
		} else {
			e.setAndFixText(fmt.Sprintf("[\"left\"]%s%c[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
				string(left), character, string(right[0]), string(right[1:])))
		}
	}
}

// NewEditor Instanciates a ready to use text editor.
func NewEditor() *Editor {
	editor := Editor{
		internalTextView: tview.NewTextView(),
		requestedHeight:  3,
	}

	editor.internalTextView.SetWrap(true)
	editor.internalTextView.SetWordWrap(true)
	editor.internalTextView.SetBorder(true)
	editor.internalTextView.SetRegions(true)
	editor.internalTextView.SetScrollable(true)
	editor.internalTextView.SetText(emptyText)
	editor.internalTextView.Highlight("selection")

	editor.internalTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Since characters can have different widths, we can't directly
		// access the string, as it is basically handled like a byte array.
		left := []rune(editor.internalTextView.GetRegionText("left"))
		right := []rune(editor.internalTextView.GetRegionText("right"))
		selection := []rune(editor.internalTextView.GetRegionText("selection"))

		// TODO: This entire chunk could be cleaned up by assigning handlers to each event type,
		// e.g. event.trigger()
		if shortcuts.MoveCursorLeft.Equals(event) {
			editor.MoveCursorLeft(left, right, selection)
		} else if shortcuts.ExpandSelectionToLeft.Equals(event) {
			editor.ExpandSelectionToLeft(left, right, selection)
		} else if shortcuts.MoveCursorRight.Equals(event) {
			editor.MoveCursorRight(left, right, selection)
		} else if shortcuts.ExpandSelectionToRight.Equals(event) {
			editor.ExpandSelectionToRight(left, right, selection)
		} else if shortcuts.SelectWordLeft.Equals(event) {
			editor.SelectWordLeft(left, right, selection)
		} else if shortcuts.SelectWordRight.Equals(event) {
			editor.SelectWordRight(left, right, selection)
		} else if shortcuts.MoveCursorWordLeft.Equals(event) {
			editor.MoveCursorWordLeft(left, right, selection)
		} else if shortcuts.MoveCursorWordRight.Equals(event) {
			editor.MoveCursorWordRight(left, right, selection)
		} else if shortcuts.SelectAll.Equals(event) {
			editor.SelectAll(left, right, selection)
		} else if shortcuts.DeleteRight.Equals(event) {
			editor.DeleteRight(left, right, selection)
		} else if event.Key() == tcell.KeyBackspace2 ||
			event.Key() == tcell.KeyBackspace {
			// FIXME Legacy, has to be replaced when there is N-1 Keybind-Mapping.
			editor.Backspace(left, right, selection)
		} else if shortcuts.CopySelection.Equals(event) {
			clipboard.WriteAll(string(selection))
			//Returning nil, as copying won't do anything than filling the
			//clipboard buffer.
			return nil
		} else if shortcuts.PasteAtSelection.Equals(event) {
			editor.Paste(left, right, selection, event)
			return nil
		} else if shortcuts.InputNewLine.Equals(event) {
			editor.InsertCharacter(left, right, selection, '\n')
		} else if shortcuts.SendMessage.Equals(event) {
			return editor.inputCapture(event)
		} else if (editor.inputCapture == nil || editor.inputCapture(event) != nil) && event.Rune() != 0 {
			editor.InsertCharacter(left, right, selection, event.Rune())
		} else {
			return event
		}

		editor.UpdateMentionHandler()
		editor.triggerHeightRequestIfNeccessary()
		editor.internalTextView.ScrollToHighlight()
		return nil
	})
	return &editor
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
	text := editor.internalTextView.GetRegionText("left")
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
	newLeft := editor.internalTextView.GetRegionText("left")
	for i := len(newLeft) - 1; i >= 0; i-- {
		if newLeft[i] == '@' && (i == 0 || newLeft[i-1] == ' ' || newLeft[i-1] == '\n') {
			return i
		}
	}
	return -1
}

func (editor *Editor) Backspace(left, right, selection []rune) {
	var newText string

	if len(selection) == 1 && len(left) >= 1 {
		newText = leftRegion + string(left[:len(left)-1]) + selRegion + string(selection) + rightRegion + string(right) + endRegion
		editor.internalTextView.SetText(newText)
	} else if len(selection) > 1 {
		newText = leftRegion + string(left) + selRegion
		if len(right) > 0 {
			newText = newText + string(right[0]) + rightRegion + string(right[1:])
		} else {
			newText = newText + selectionChar
		}
		newText = newText + endRegion
		editor.setAndFixText(newText)
	}
}

func (editor *Editor) setAndFixText(text string) {
	newText := multiSelectionCharWithSelectionToLeftPattern.ReplaceAllString(text, selRegion+selectionChar+endRegion)
	editor.internalTextView.SetText(newText)
}

func (editor *Editor) countRows(text string) int {
	_, _, width, _ := editor.internalTextView.GetInnerRect()
	return tviewutil.CalculateNeccessaryHeight(width, text)
}

func (editor *Editor) triggerHeightRequestIfNeccessary() {
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
	if text == "" {
		editor.internalTextView.SetText(emptyText)
	} else {
		editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s[\"\"][\"selection\"]%s[\"\"]", text, string(selectionChar)))
	}

	editor.triggerHeightRequestIfNeccessary()
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
	left := editor.internalTextView.GetRegionText("left")
	right := editor.internalTextView.GetRegionText("right")
	selection := editor.internalTextView.GetRegionText("selection")

	if right == "" && selection == string(selectionChar) {
		return left
	}

	return left + selection + right
}

// GetPrimitive returnbs the internal component that can be added to a layout
func (editor *Editor) GetPrimitive() tview.Primitive {
	return editor.internalTextView
}
