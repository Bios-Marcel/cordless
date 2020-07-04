## femto, an editor component for tview

`femto` is a text editor component for tview. The vast majority of the code is derived from
[the micro editor](github.com/zyedidia/micro).

**Note** The shape of the `femto` API is a work-in-progress, and should not be considered stable.

### Default keybindings
```
Up:             CursorUp
Down:           CursorDown
Right:          CursorRight
Left:           CursorLeft
ShiftUp:        SelectUp
ShiftDown:      SelectDown
ShiftLeft:      SelectLeft
ShiftRight:     SelectRight
AltLeft:        WordLeft
AltRight:       WordRight
AltUp:          MoveLinesUp
AltDown:        MoveLinesDown
AltShiftRight:  SelectWordRight
AltShiftLeft:   SelectWordLeft
CtrlLeft:       StartOfLine
CtrlRight:      EndOfLine
CtrlShiftLeft:  SelectToStartOfLine
ShiftHome:      SelectToStartOfLine
CtrlShiftRight: SelectToEndOfLine
ShiftEnd:       SelectToEndOfLine
CtrlUp:         CursorStart
CtrlDown:       CursorEnd
CtrlShiftUp:    SelectToStart
CtrlShiftDown:  SelectToEnd
Alt-{:          ParagraphPrevious
Alt-}:          ParagraphNext
Enter:          InsertNewline
CtrlH:          Backspace
Backspace:      Backspace
Alt-CtrlH:      DeleteWordLeft
Alt-Backspace:  DeleteWordLeft
Tab:            IndentSelection,InsertTab
Backtab:        OutdentSelection,OutdentLine
CtrlZ:          Undo
CtrlY:          Redo
CtrlC:          Copy
CtrlX:          Cut
CtrlK:          CutLine
CtrlD:          DuplicateLine
CtrlV:          Paste
CtrlA:          SelectAll
Home:           StartOfLine
End:            EndOfLine
CtrlHome:       CursorStart
CtrlEnd:        CursorEnd
PageUp:         CursorPageUp
PageDown:       CursorPageDown
CtrlR:          ToggleRuler
Delete:         Delete
Insert:         ToggleOverwriteMode
Alt-f:          WordRight
Alt-b:          WordLeft
Alt-a:          StartOfLine
Alt-e:          EndOfLine
Esc:            Escape
Alt-n:          SpawnMultiCursor
Alt-m:          SpawnMultiCursorSelect
Alt-p:          RemoveMultiCursor
Alt-c:          RemoveAllMultiCursors
Alt-x:          SkipMultiCursor
```

### Example Usage

The code below (also found in `cmd/femto/femto.go`) creates a `tview` application with a single full-screen editor
that operates on one file at a time. Ctrl-s saves any edits; Ctrl-q quits.

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gdamore/tcell"
	"femto"
	"femto/runtime"
	"github.com/Bios-Marcel/cordless/tview"
)

func saveBuffer(b *femto.Buffer, path string) error {
	return ioutil.WriteFile(path, []byte(b.String()), 0600)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: femto [filename]\n")
		os.Exit(1)
	}
	path := os.Args[1]

	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("could not read %v: %v", path, err)
	}

	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}

	app := tview.NewApplication()
	buffer := femto.NewBufferFromString(string(content), path)
	root := femto.NewView(buffer)
	root.SetRuntimeFiles(runtime.Files)
	root.SetColorscheme(colorscheme)
	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			saveBuffer(buffer, path)
			return nil
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		}
		return event
	})
	app.SetRoot(root, true)

	if err := app.Run(); err != nil {
		log.Fatalf("%v", err)
	}
}
```
