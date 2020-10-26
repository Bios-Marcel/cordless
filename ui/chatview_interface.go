package ui

import (
	"sync"

	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"
	tcell "github.com/gdamore/tcell/v2"
)

type ChatView interface {
	tview.Primitive
	sync.Locker

	SetOnMessageAction(func(*discordgo.Message, *tcell.EventKey) *tcell.EventKey)
	SetText([]*TextBlock)
	AddMessage(*discordgo.Message)
	SetMessages([]*discordgo.Message)
	UpdateMessage(*discordgo.Message)
	GetData() []*discordgo.Message
	MessageCount() int
	DeleteMessage(*discordgo.Message)
	//FIXME []string is an API inconsistency
	DeleteMessages([]string)
	ClearViewAndCache()
	ClearSelection()
	SetTitle(string)
	Reprint()
	SignalSelectionDeleted()
	ScrollUp()
	ScrollDown()
	ScrollToStart()
	ScrollToEnd()
	SetInputCapture(func(*tcell.EventKey) *tcell.EventKey)
	SetMouseHandler(func(*tcell.EventMouse) bool)
	SetNextFocusableComponents(tview.FocusDirection, ...tview.Primitive)
	SetBorderSides(top, left, bottom, right bool)
	Dispose()
}

type TextBlock struct {
	content  string
	style tcell.Style
}
