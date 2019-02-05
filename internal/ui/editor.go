package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

var (
	// Temporary solution, so not every function has to handle the selection
	// character placement.
	multiSelectionCharWithSelectionToLeftPattern = regexp.MustCompile(selectionChar + "*" + regexp.QuoteMeta(selRegion) + selectionChar + "*" + regexp.QuoteMeta(endRegion))
)

const (
	selectionRune = '\u205F'
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

	inputCapture            func(event *tcell.EventKey) *tcell.EventKey
	mentionCharacterHandler func(event *tcell.EventKey) *tcell.EventKey
	heightRequestHandler    func(requestHeight int)
	requestedHeight         int
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
		left := []rune(editor.internalTextView.GetRegionText("left"))
		right := []rune(editor.internalTextView.GetRegionText("right"))
		selection := []rune(editor.internalTextView.GetRegionText("selection"))

		var newText string
		if event.Key() == tcell.KeyLeft &&
			(event.Modifiers() == tcell.ModShift || event.Modifiers() == tcell.ModNone) {
			expandSelection := (event.Modifiers() & tcell.ModShift) == tcell.ModShift
			if len(left) > 0 {
				newText = leftRegion + string(left[:len(left)-1]) + selRegion

				currentSelection := string(selection)
				if currentSelection == selectionChar {
					currentSelection = ""
				}

				if expandSelection {
					newText = newText + string(left[len(left)-1]) + currentSelection + rightRegion + string(right)
				} else {
					newText = newText + string(left[len(left)-1]) + rightRegion + currentSelection + string(right)
				}

				newText = newText + endRegion
				editor.internalTextView.SetText(newText)
			} else if len(selection) > 0 && !expandSelection {
				if len(right) > 0 {
					newText = selRegion + string(selection[0]) + rightRegion + string(selection[1:]) + string(right) + endRegion
				} else {
					newText = selRegion + string(selection[0]) + rightRegion + string(selection[1:]) + endRegion
				}
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyRight &&
			(event.Modifiers() == tcell.ModShift || event.Modifiers() == tcell.ModNone) {
			newText = leftRegion + string(left)
			expandSelection := (event.Modifiers() & tcell.ModShift) == tcell.ModShift
			if len(right) > 0 {
				if expandSelection {
					newText = newText + selRegion + string(selection) + string(right[0]) + rightRegion + string(right[1:])
				} else {
					newText = newText + string(selection) + selRegion + string(right[0]) + rightRegion + string(right[1:])
				}
			} else {
				endsWithSelectionChar := strings.HasSuffix(string(selection), selectionChar)
				if !endsWithSelectionChar {
					if expandSelection {
						newText = newText + selRegion + string(selection)
					} else if !expandSelection {
						newText = newText + string(selection) + selRegion
					}

					newText = newText + selectionChar
				} else {
					if expandSelection {
						newText = newText + selRegion + string(selection)
					} else {
						newText = newText + string(selection[:len(selection)-1]) + selRegion + selectionChar
					}
				}
			}

			newText = newText + endRegion
			editor.setAndFixText(newText)
		} else if event.Key() == tcell.KeyLeft &&
			(event.Modifiers()&(tcell.ModShift|tcell.ModCtrl)) == (tcell.ModShift|tcell.ModCtrl) {
			if len(left) > 0 {
				selectionFrom := 0
				for i := len(left) - 2; /*Skip space left to selection*/ i >= 0; i-- {
					if left[i] == ' ' {
						selectionFrom = i
						break
					}
				}

				if selectionFrom != 0 {
					newText = leftRegion + string(left[:selectionFrom+1]) + selRegion + string(left[selectionFrom+1:]) + string(string(selection)) + rightRegion + string(right) + endRegion
				} else {
					newText = selRegion + string(left) + string(string(selection)) + rightRegion + string(right) + endRegion
				}
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyRight &&
			(event.Modifiers()&(tcell.ModShift|tcell.ModCtrl)) == (tcell.ModShift|tcell.ModCtrl) {
			if len(right) > 0 {
				selectionFrom := len(right) - 1
				for i := 1; /*Skip space right to selection*/ i < len(right)-1; i++ {
					if right[i] == ' ' {
						selectionFrom = i
						break
					}
				}

				if selectionFrom != len(right)-1 {
					newText = leftRegion + string(left) + selRegion + string(string(selection)) + string(right[:selectionFrom]) + rightRegion + string(right[selectionFrom:]) + endRegion
				} else {
					newText = leftRegion + string(left) + selRegion + string(string(selection)) + string(right) + endRegion
				}
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyRight &&
			event.Modifiers() == tcell.ModCtrl {
			if len(right) > 0 {
				selectionAt := len(right) - 1
				for i := 1; /*Skip space right to selection*/ i < len(right)-1; i++ {
					if right[i] == ' ' {
						selectionAt = i
						break
					}
				}

				if selectionAt != len(right)-1 {
					newText = leftRegion + string(left) + string(string(selection)) + string(right[:selectionAt]) + selRegion + string(right[selectionAt]) + rightRegion + string(right[selectionAt+1:]) + endRegion
				} else {
					newText = leftRegion + string(left) + string(selection) + string(right) + selRegion + selectionChar + endRegion
				}
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyLeft &&
			event.Modifiers() == tcell.ModCtrl {
			if len(left) > 0 {
				selectionAt := 0
				for i := len(left) - 2; /*Skip space left to selection*/ i >= 0; i-- {
					if left[i] == ' ' {
						selectionAt = i
						break
					}
				}

				if selectionAt != 0 {
					newText = leftRegion + string(left[:selectionAt]) + selRegion + string(left[selectionAt]) + rightRegion + string(left[selectionAt+1:]) + string(string(selection)) + string(right) + endRegion
				} else {
					if len(left) > 1 {
						newText = selRegion + string(left[0]) + rightRegion + string(left[1:]) + string(selection) + string(right) + endRegion
					} else {
						newText = selRegion + string(left[0]) + rightRegion + string(selection) + string(right) + endRegion
					}
				}
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyCtrlA {
			if len(left) > 0 || len(right) > 0 {
				newText = selRegion + string(left) + string(selection) + string(right) + endRegion
				editor.setAndFixText(newText)
			}
		} else if event.Key() == tcell.KeyBackspace2 ||
			event.Key() == tcell.KeyBackspace {
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
		} else if event.Key() == tcell.KeyCtrlV {
			clipBoardContent, clipError := clipboard.ReadAll()
			if clipError == nil {
				newText = leftRegion + string(left) + selRegion + string(selection) + clipBoardContent + rightRegion + string(right) + endRegion
				editor.setAndFixText(newText)
			}
		} else {
			var character rune
			if event.Key() == tcell.KeyEnter {
				if (event.Modifiers() & tcell.ModAlt) == tcell.ModAlt {
					character = '\n'
				}
			} else {
				character = event.Rune()
			}

			if character == 0 {
				editor.inputCapture(event)
				return nil
			}

			if len(right) == 0 {
				editor.setAndFixText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"]", string(left), (string)(character), string(selectionChar)))
			} else {
				editor.setAndFixText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
					string(left), string(character), string(selection), string(right)))
			}

			if character == '@' {
				editor.mentionCharacterHandler(event)
			}
		}

		editor.triggerHeightRequestIfNeccessary()

		return nil
	})

	return &editor
}

func (editor *Editor) setAndFixText(text string) {
	newText := multiSelectionCharWithSelectionToLeftPattern.ReplaceAllString(text, selRegion+selectionChar+endRegion)
	editor.internalTextView.SetText(newText)
}

func (editor *Editor) triggerHeightRequestIfNeccessary() {
	splitLines := strings.Split(editor.GetText(), "\n")
	_, _, width, _ := editor.internalTextView.GetInnerRect()

	wrappedLines := 0
	for _, line := range splitLines {
		if len(line) > width {
			wrappedLines = wrappedLines + ((len(line) - (len(line) % width)) / width)
		}
	}

	newRequestedHeight := len(splitLines) + wrappedLines + 2 /*borders*/
	if newRequestedHeight != editor.requestedHeight {
		editor.requestedHeight = newRequestedHeight
		editor.heightRequestHandler(newRequestedHeight)
	}
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

// SetInputCapture sets the alternative input capture that will be used if the
// components default controls aren't being triggered.
func (editor *Editor) SetInputCapture(captureFunc func(event *tcell.EventKey) *tcell.EventKey) {
	editor.inputCapture = captureFunc
}

func (editor *Editor) SetMentionCharacterHandler(handlerFunc func(event *tcell.EventKey) *tcell.EventKey) {
	editor.mentionCharacterHandler = handlerFunc
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
