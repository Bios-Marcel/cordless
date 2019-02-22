package main

import (
	"fmt"
	"strings"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

var shortcuts map[string]*tcell.EventKey

func main() {
	shortcuts = map[string]*tcell.EventKey{
		"exit":  tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone),
		"dunno": tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone),
	}

	table := tview.NewTable()
	table.SetSelectable(true, false)

	table.SetFixed(2, 2)
	actionCell := tview.NewTableCell("Action")
	actionCell.SetSelectable(false)
	actionCell.SetAlign(tview.AlignCenter)
	actionCell.SetExpansion(1)

	shortcutCell := tview.NewTableCell("Shortcut")
	shortcutCell.SetSelectable(false)
	shortcutCell.SetAlign(tview.AlignCenter)
	shortcutCell.SetExpansion(1)

	table.SetCell(0, 0, actionCell)
	table.SetCell(0, 1, shortcutCell)
	table.SetBorder(true)

	emptyCellOne := tview.NewTableCell("")
	emptyCellOne.SetSelectable(false)
	table.SetCell(1, 0, emptyCellOne)
	emptyCellTwo := tview.NewTableCell("")
	emptyCellTwo.SetSelectable(false)
	table.SetCell(2, 0, emptyCellTwo)

	rows := make([]string, len(shortcuts))

	app := tview.NewApplication()

	selection := -1
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			return event
		}

		selectedRow, _ := table.GetSelection()
		if selection == -1 && selectedRow >= 2 {
			if event.Key() == tcell.KeyEnter {
				if selectedRow != -1 {
					table.SetCellSimple(selectedRow, 1, "Please hit your desired keystroke")
					selection = selectedRow
				}
			} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
				table.SetCellSimple(selectedRow, 1, "")
				shortcuts[rows[selectedRow-2]] = nil
			}
		} else if selectedRow >= 2 && selection != -1 {
			name := rows[selectedRow-2]
			_, available := shortcuts[name]
			if available {
				table.SetCellSimple(selectedRow, 1, EventToString(event))
				shortcuts[name] = event
				selection = -1
			}
		}

		return event
	})

	row := 2

	for name, event := range shortcuts {
		nameCell := tview.NewTableCell(name)
		nameCell.SetExpansion(1)
		table.SetCell(row, 0, nameCell)

		eventCell := tview.NewTableCell(EventToString(event))
		eventCell.SetExpansion(1)
		table.SetCell(row, 1, eventCell)

		rows[row-2] = name
		row++
	}

	app.SetRoot(table, true).Run()
}

// EventsEqual compares the given events, respecting everything except for the
// When field.
func EventsEqual(eventOne, eventTwo *tcell.EventKey) bool {
	return eventOne.Rune() == eventTwo.Rune() &&
		eventOne.Modifiers() == eventTwo.Modifiers() &&
		eventOne.Key() == eventTwo.Key()
}

// EventToString renders a tcell.EventKey as a human readable string
func EventToString(event *tcell.EventKey) string {
	s := ""
	m := []string{}
	if event.Modifiers()&tcell.ModCtrl != 0 {
		m = append(m, "Ctrl")
	}
	if event.Modifiers()&tcell.ModShift != 0 {
		m = append(m, "Shift")
	}
	if event.Modifiers()&tcell.ModAlt != 0 {
		m = append(m, "Alt")
	}
	if event.Modifiers()&tcell.ModMeta != 0 {
		m = append(m, "Meta")
	}

	ok := false
	if s, ok = tcell.KeyNames[event.Key()]; !ok {
		if event.Key() == tcell.KeyRune {
			if event.Rune() >= 'A' && event.Rune() <= 'Z' {
				s = "Shift+" + string(event.Rune())
			} else {
				s = strings.ToUpper(string(event.Rune()))
			}
		} else {
			s = fmt.Sprintf("Key[%d,%d]", event.Key(), int(event.Rune()))
		}
	}
	if len(m) != 0 {
		if event.Modifiers()&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
		}
		return fmt.Sprintf("%s+%s", strings.Join(m, "+"), s)
	}

	return tview.Escape(s)
}
