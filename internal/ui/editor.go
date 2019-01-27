package ui

import (
	"fmt"
	"strings"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

const (
	spaceChar = '\u205F'
	emptyText = "[\"selection\"]\u205F[\"\"]"
)

//Editor is a simple component that wraps tview.TextView in order to gove the
//user minimal text edit functionallity.
type Editor struct {
	internalTextView *tview.TextView

	inputCapture         func(event *tcell.EventKey) *tcell.EventKey
	heightRequestHandler func(requestHeight int)
	requestedHeight      int
}

//NewEditor Instanciates a ready to use text editor.
func NewEditor() *Editor {
	editor := Editor{
		internalTextView: tview.NewTextView(),
		requestedHeight:  3,
	}

	editor.internalTextView.SetWrap(true)
	editor.internalTextView.SetWordWrap(true)
	editor.internalTextView.SetBorder(true)
	editor.internalTextView.SetRegions(true)
	editor.internalTextView.SetText(emptyText)
	editor.internalTextView.Highlight("selection")

	editor.internalTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		left := []rune(editor.internalTextView.GetRegionText("left"))
		right := []rune(editor.internalTextView.GetRegionText("right"))
		selection := []rune(editor.internalTextView.GetRegionText("selection"))

		if event.Key() == tcell.KeyLeft {
			lengthMinusOne := len(left) - 1
			if lengthMinusOne > -1 {
				editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s[\"\"][\"selection\"]%s[\"\"][\"right\"]%s%s[\"\"]",
					string(left[:len(left)-1]), string(left[len(left)-1:]), string(selection), string(right)))
				editor.internalTextView.Highlight("selection")
			}
			return nil
		} else if event.Key() == tcell.KeyRight {
			if len(right) == 0 {
				if selection[0] == spaceChar {
					editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s[\"\"][\"selection\"]%s[\"\"]",
						string(left), string(selection)))
				} else {
					editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"]",
						string(left), string(selection), string(spaceChar)))
				}
			} else {
				editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
					string(left), string(selection), string(right[:1]), string(right[1:])))
			}

			editor.internalTextView.Highlight("selection")
			return nil
		} else if event.Key() == tcell.KeyBackspace2 {
			if (len(left) - 1) > -1 {
				editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
					string(left[:len(left)-1]), string(selection), string(right)))
				editor.internalTextView.Highlight("selection")
				editor.triggerHeightRequestIfNeccessary()
			}
			return nil
		}

		var character rune
		//TODO Find a way to listen to Shift + Enter, tcell or tview seem to ignore it.
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
			editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"]", string(left), (string)(character), string(spaceChar)))
		} else {
			editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s%s[\"\"][\"selection\"]%s[\"\"][\"right\"]%s[\"\"]",
				string(left), string(character), string(selection), string(right)))
		}
		editor.internalTextView.Highlight("selection")

		editor.triggerHeightRequestIfNeccessary()

		return nil
	})

	return &editor
}

func (editor *Editor) triggerHeightRequestIfNeccessary() {
	//+3 because of borders and the fact that there is always a line
	newRequestedHeight := strings.Count(editor.GetText(), "\n") + 3
	if newRequestedHeight != editor.requestedHeight {
		editor.requestedHeight = newRequestedHeight
		editor.heightRequestHandler(newRequestedHeight)
	}
}

//SetOnHeightChangeRequest handles the cases where the component thinks it needs
//more space or would be fine with less.
func (editor *Editor) SetOnHeightChangeRequest(handler func(requestHeight int)) {
	editor.heightRequestHandler = handler
}

//SetBackgroundColor sets the background color of the internal TextView
func (editor *Editor) SetBackgroundColor(color tcell.Color) {
	editor.internalTextView.SetBackgroundColor(color)
}

//SetText sets the texts of the internal TextView, but also sets the selection
//and necessary groups for the navigation behaviour.
func (editor *Editor) SetText(text string) {
	if text == "" {
		editor.internalTextView.SetText(emptyText)
	} else {
		editor.internalTextView.SetText(fmt.Sprintf("[\"left\"]%s[\"\"][\"selection\"]%s[\"\"]", text, string(spaceChar)))
	}

	editor.triggerHeightRequestIfNeccessary()
}

//SetInputCapture sets the alternative input capture that will be used if the
//components default controls aren't being triggered.
func (editor *Editor) SetInputCapture(captureFunc func(event *tcell.EventKey) *tcell.EventKey) {
	editor.inputCapture = captureFunc
}

//GetText returns the text without color tags, region tags and so on.
func (editor *Editor) GetText() string {
	left := editor.internalTextView.GetRegionText("left")
	right := editor.internalTextView.GetRegionText("right")
	selection := editor.internalTextView.GetRegionText("selection")

	if right == "" && selection == string(spaceChar) {
		return left
	}

	return left + selection + right
}

//GetPrimitive returnbs the internal component that can be added to a layout
func (editor *Editor) GetPrimitive() tview.Primitive {
	return editor.internalTextView
}
