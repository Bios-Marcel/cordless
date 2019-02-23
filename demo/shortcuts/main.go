package main

import (
	"fmt"
	"strings"

	"github.com/Bios-Marcel/tview"
	"github.com/gdamore/tcell"
)

const (
	actionCellIndex int = iota
	scopeCellIndex
	shortcutCellIndex
)

// Shortcut defines a shortcut within the application. The scope might for
// example be a widget or situation in which the user is.
type Shortcut struct {
	name  string
	scope string
	event *tcell.EventKey
}

// ShortcutTable is a component that displays shortcuts and allows changing
// them.
type ShortcutTable struct {
	table     *tview.Table
	shortcuts []*Shortcut
	selection int
}

// NewShortcutTable creates a new shortcut table that doesn't contain any data.
func NewShortcutTable() *ShortcutTable {
	table := tview.NewTable()
	shortcutsTable := &ShortcutTable{
		table:     table,
		selection: -1,
	}

	table.SetSelectable(true, false)
	table.SetBorder(true)

	//Header + emptyrow
	table.SetFixed(2, 3)

	table.SetCell(0, actionCellIndex, createHeaderCell("Action"))
	table.SetCell(0, scopeCellIndex, createHeaderCell("Scope"))
	table.SetCell(0, shortcutCellIndex, createHeaderCell("Shortcut"))

	table.SetInputCapture(shortcutsTable.handleInput)

	return shortcutsTable
}

func createHeaderCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetAlign(tview.AlignCenter).
		SetExpansion(1)
}

// SetShortcuts sets the shortcut data and changes the UI accordingly.
func (shortcutTable *ShortcutTable) SetShortcuts(shortcuts []*Shortcut) {
	shortcutTable.shortcuts = shortcuts

	row, _ := shortcutTable.table.GetFixed()

	//Using clear will remove the content of the fixed rows, therefore we
	// manually remove starting from the first non-fixed row.
	for index := row; index < shortcutTable.table.GetRowCount(); index++ {
		shortcutTable.table.RemoveRow(index)
	}

	for _, shortcut := range shortcuts {
		nameCell := tview.NewTableCell(shortcut.name)
		nameCell.SetExpansion(1)
		shortcutTable.table.SetCell(row, actionCellIndex, nameCell)

		scopeCell := tview.NewTableCell(shortcut.scope)
		scopeCell.SetExpansion(1)
		shortcutTable.table.SetCell(row, scopeCellIndex, scopeCell)

		eventCell := tview.NewTableCell(EventToString(shortcut.event))
		eventCell.SetExpansion(1)
		shortcutTable.table.SetCell(row, shortcutCellIndex, eventCell)

		row++
	}
}

// GetShortcuts returns the array containing the currently displayed shortcuts.
func (shortcutTable *ShortcutTable) GetShortcuts() []*Shortcut {
	return shortcutTable.shortcuts
}

func (shortcutTable *ShortcutTable) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if shortcutTable.selection == -1 && (event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown) {
		return event
	}

	firstNonFixedRow, _ := shortcutTable.table.GetFixed()
	selectedRow, _ := shortcutTable.table.GetSelection()
	//The first row of the table isn't the first row containing data
	dataIndex := selectedRow - firstNonFixedRow
	if shortcutTable.selection == -1 && selectedRow >= firstNonFixedRow {
		if event.Key() == tcell.KeyEnter {
			if selectedRow != -1 {
				shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText("Hit the desired keycombination")
				shortcutTable.selection = selectedRow
				return nil
			}
		} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText("")
			shortcutTable.shortcuts[dataIndex].event = nil
			return nil
		}
	} else if selectedRow >= firstNonFixedRow && shortcutTable.selection != -1 {
		shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText(EventToString(event))
		//Make a copy of the event?
		shortcutTable.shortcuts[dataIndex].event = event
		shortcutTable.selection = -1
		return nil
	}

	return event
}

func main() {

	app := tview.NewApplication()

	shortcutTable := NewShortcutTable()
	shortcutTable.SetShortcuts([]*Shortcut{
		&Shortcut{"dunno", "somewhere", tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone)},
		&Shortcut{"dunno2", "somewhere", tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone)},
		&Shortcut{"ooof", "ok", tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone)},
		&Shortcut{"whatever", "nowhere", tcell.NewEventKey(tcell.KeyCtrlK, 0, tcell.ModNone)},
	})

	app.SetRoot(shortcutTable.table, true).Run()
}

// EventsEqual compares the given events, respecting everything except for the
// When field.
func EventsEqual(eventOne, eventTwo *tcell.EventKey) bool {
	if (eventOne == nil && eventTwo != nil) || (eventOne != nil && eventTwo == nil) {
		return false
	}

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
