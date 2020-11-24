package shortcutdialog

import (
	"fmt"
	"strings"

	"github.com/Bios-Marcel/cordless/shortcuts"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/cordless/util/vim"
	tcell "github.com/gdamore/tcell/v2"
)

const (
	actionCellIndex int = iota
	scopeCellIndex
	shortcutCellIndex
	vimNormalCellIndex
	vimInsertCellIndex
	vimVisualCellIndex
)

// ShortcutTable is a component that displays shortcuts and allows changing
// them.
type ShortcutTable struct {
	table         *tview.Table
	shortcuts     []*shortcuts.Shortcut
	vimBindings   []*shortcuts.VimEvent
	selection     int
	focusNext     func()
	focusPrevious func()

	vimMode *int
}

// NewShortcutTable creates a new shortcut table that doesn't contain any data.
func NewShortcutTable(vimEnabled *int) *ShortcutTable {
	table := tview.NewTable()
	shortcutsTable := &ShortcutTable{
		table:     table,
		selection: -1,
		vimMode:   vimEnabled,
	}

	table.SetSelectable(true, false)
	table.SetBorder(true)

	//Header + emptyrow
	table.SetFixed(2, 3)

	table.SetCell(0, actionCellIndex, createHeaderCell("Action"))
	table.SetCell(0, scopeCellIndex, createHeaderCell("Scope"))
	table.SetCell(0, shortcutCellIndex, createHeaderCell("Shortcut"))
	if *vimEnabled != vim.Disabled {
		table.SetCell(0, vimNormalCellIndex, createHeaderCell("Vim: N"))
		table.SetCell(0, vimInsertCellIndex, createHeaderCell("Vim: I"))
		table.SetCell(0, vimVisualCellIndex, createHeaderCell("Vim: V"))
	}

	table.SetInputCapture(shortcutsTable.handleInput)

	return shortcutsTable
}

// GetPrimitive returns the primitive to be put onto a layout.
func (shortcutTable *ShortcutTable) GetPrimitive() tview.Primitive {
	return shortcutTable.table
}

func createHeaderCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetAlign(tview.AlignCenter).
		SetExpansion(1).
		SetMaxWidth(1)
}

// SetShortcuts sets the shortcut data and changes the UI accordingly.
func (shortcutTable *ShortcutTable) SetShortcuts(shortcuts []*shortcuts.Shortcut) {
	shortcutTable.shortcuts = shortcuts
	shortcutTable.selection = -1

	row, _ := shortcutTable.table.GetFixed()

	//Using clear will remove the content of the fixed rows, therefore we
	// manually remove starting from the first non-fixed row.
	for index := row; index < shortcutTable.table.GetRowCount(); index++ {
		shortcutTable.table.RemoveRow(index)
	}

	for _, shortcut := range shortcuts {
		nameCell := tview.NewTableCell(shortcut.Name).
			SetExpansion(1).
			SetMaxWidth(1)
		shortcutTable.table.SetCell(row, actionCellIndex, nameCell)

		scopeCell := tview.NewTableCell(shortcut.Scope.Name).
			SetExpansion(1).
			SetMaxWidth(1)
		shortcutTable.table.SetCell(row, scopeCellIndex, scopeCell)

		eventCell := tview.NewTableCell(EventToString(shortcut.Event)).
			SetExpansion(1).
			SetMaxWidth(1)
		shortcutTable.table.SetCell(row, shortcutCellIndex, eventCell)

		if *shortcutTable.vimMode != vim.Disabled {
			var normalKey string
			var insertKey string
			var visualKey string

			// HACK
			// Empty tcell EventKey corresponds to Ctrl-Space
			if shortcut.VimModifier.NormalEvent == nil {
				normalKey = EventToString(shortcut.Event)
			} else if shortcut.VimModifier.NormalEvent.Rune() == rune(tcell.KeyCtrlSpace) {
				normalKey = ""
			} else {
				normalKey = EventToString(shortcut.VimModifier.NormalEvent)
			}

			if shortcut.VimModifier.InsertEvent == nil {
				insertKey = EventToString(shortcut.Event)
			} else if shortcut.VimModifier.InsertEvent.Rune() == rune(tcell.KeyCtrlSpace) {
				insertKey = ""
			} else {
				insertKey = EventToString(shortcut.VimModifier.InsertEvent)
			}

			if shortcut.VimModifier.VisualEvent == nil {
				visualKey = EventToString(shortcut.Event)
			} else if shortcut.VimModifier.VisualEvent.Rune() == rune(tcell.KeyCtrlSpace) {
				visualKey = ""
			} else {
				visualKey = EventToString(shortcut.VimModifier.VisualEvent)
			}

			vimNormalCell := tview.NewTableCell(fmt.Sprintf("%s",
				normalKey)).
				SetExpansion(1).
				SetMaxWidth(1)
			shortcutTable.table.SetCell(row, vimNormalCellIndex, vimNormalCell)

			vimInsertCell := tview.NewTableCell(fmt.Sprintf("%s",
				insertKey)).
				SetExpansion(1).
				SetMaxWidth(1)
			shortcutTable.table.SetCell(row, vimInsertCellIndex, vimInsertCell)

			vimVisualCell := tview.NewTableCell(fmt.Sprintf("%s",
				visualKey)).
				SetExpansion(1).
				SetMaxWidth(1)
			shortcutTable.table.SetCell(row, vimVisualCellIndex, vimVisualCell)
		}

		row++
	}
}

// GetShortcuts returns the array containing the currently displayed shortcuts.
func (shortcutTable *ShortcutTable) GetShortcuts() []*shortcuts.Shortcut {
	return shortcutTable.shortcuts
}

// SetFocusNext decides which component will be focused when hitting tab
func (shortcutTable *ShortcutTable) SetFocusNext(function func()) {
	shortcutTable.focusNext = function
}

// SetFocusPrevious decides which component will be focused when hitting
// shift+tab
func (shortcutTable *ShortcutTable) SetFocusPrevious(function func()) {
	shortcutTable.focusPrevious = function
}

func (shortcutTable *ShortcutTable) handleInput(event *tcell.EventKey) *tcell.EventKey {
	if shortcutTable.selection == -1 {
		if event.Key() == tcell.KeyTab {
			if shortcutTable.focusNext != nil {
				shortcutTable.focusNext()
			}
		} else if event.Key() == tcell.KeyBacktab {
			if shortcutTable.focusPrevious != nil {
				shortcutTable.focusPrevious()
			}
		}
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			return event
		}
	}

	firstNonFixedRow, _ := shortcutTable.table.GetFixed()
	selectedRow, _ := shortcutTable.table.GetSelection()
	//The first row of the table isn't the first row containing data
	dataIndex := selectedRow - firstNonFixedRow
	if shortcutTable.selection == -1 && selectedRow >= firstNonFixedRow {
		if event.Key() == tcell.KeyEnter {
			if selectedRow != -1 {
				shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText("[blue][::ub]Hit the desired keycombination")
				shortcutTable.selection = selectedRow
			}
		} else if event.Key() == tcell.KeyRune && event.Rune() == 'r' && event.Modifiers() == tcell.ModNone {
			shortcut := shortcutTable.shortcuts[dataIndex]
			shortcut.Reset()
			shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText(EventToString(shortcut.Event))

			persistError := shortcuts.Persist()
			if persistError != nil {
				panic(persistError)
			}
		} else if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText("")
			shortcutTable.shortcuts[dataIndex].Event = nil

			persistError := shortcuts.Persist()
			if persistError != nil {
				panic(persistError)
			}
		}
	} else if selectedRow >= firstNonFixedRow && shortcutTable.selection != -1 {
		shortcutTable.table.GetCell(selectedRow, shortcutCellIndex).SetText(EventToString(event))

		shortcutTable.shortcuts[dataIndex].Event = tcell.NewEventKey(
			event.Key(),
			event.Rune(),
			event.Modifiers())

		persistError := shortcuts.Persist()
		if persistError != nil {
			panic(persistError)
		}

		shortcutTable.selection = -1
	}

	return nil
}

// EventToString renders a tcell.EventKey as a human readable string
func EventToString(event *tcell.EventKey) string {
	if event == nil {
		return ""
	}

	var m []string
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

	var (
		s  string
		ok bool
	)

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

// IsDefiningShortcut indicates whether the user is currently selecting a
// shortcut for any function.
func (shortcutTable *ShortcutTable) IsDefiningShortcut() bool {
	return shortcutTable.selection != -1
}
