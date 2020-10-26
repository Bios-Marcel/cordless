// +build chatview_revamp

package ui

import (
	"sync"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/Bios-Marcel/discordgo"
	tcell "github.com/gdamore/tcell/v2"
)

type revampedChatView struct {
	*sync.Mutex
	Box *tview.Box

	onMessageInteraction func(*discordgo.Message, *tcell.EventKey) *tcell.EventKey
}

func NewChatView(state *discordgo.State, ownUserID string) ChatView {
	chatView := &revampedChatView{
		Mutex: &sync.Mutex{},
		Box:   tview.NewBox(),
	}

	chatView.Box.
		SetIndicateOverflow(true).
		SetBorder(true).
		SetTitleColor(config.GetTheme().InverseTextColor)

	return chatView
}

// Draw draws this primitive onto the screen. Implementers can call the
// screen's ShowCursor() function but should only do so when they have focus.
// (They will need to keep track of this themselves.)
func (chatView *revampedChatView) Draw(screen tcell.Screen) bool {
	boxDrawn := chatView.Box.Draw(screen)
	if !boxDrawn {
		return false
	}

	//TODO Draw

	return true
}

func (chatView *revampedChatView) SetOnMessageAction(onMessageInteraction func(*discordgo.Message, *tcell.EventKey) *tcell.EventKey) {
	chatView.onMessageInteraction = onMessageInteraction
}

func (chatView *revampedChatView) SetText(_ string) {
	//TODO
}

func (chatView *revampedChatView) AddMessage(_ *discordgo.Message) {
	//TODO
}

func (chatView *revampedChatView) SetMessages(_ []*discordgo.Message) {
	//TODO
}

func (chatView *revampedChatView) UpdateMessage(_ *discordgo.Message) {
	//TODO
}

func (chatView *revampedChatView) GetData() []*discordgo.Message {
	//TODO
	return nil
}

func (chatView *revampedChatView) MessageCount() int {
	//TODO
	return 0
}

func (chatView *revampedChatView) DeleteMessage(_ *discordgo.Message) {
	//TODO
}

//FIXME []string is an API inconsistency
func (chatView *revampedChatView) DeleteMessages(_ []string) {
	//TODO
}

func (chatView *revampedChatView) ClearViewAndCache() {
	//TODO
}

func (chatView *revampedChatView) ClearSelection() {
	//TODO
}

func (chatView *revampedChatView) SetTitle(_ string) {
	//TODO
}

func (chatView *revampedChatView) Reprint() {
	//TODO
}

func (chatView *revampedChatView) SignalSelectionDeleted() {
	//TODO
}

func (chatView *revampedChatView) ScrollUp() {
	//TODO
}

func (chatView *revampedChatView) ScrollDown() {
	//TODO
}

func (chatView *revampedChatView) ScrollToStart() {
	//TODO
}

func (chatView *revampedChatView) ScrollToEnd() {
	//TODO
}

// Sets whether the primitive should be drawn onto the screen.
func (chatView *revampedChatView) SetVisible(visible bool) {
	chatView.Box.SetVisible(visible)
}

// Gets whether the primitive should be drawn onto the screen.
func (chatView *revampedChatView) IsVisible() bool {
	return chatView.Box.IsVisible()
}

// GetRect returns the current position of the primitive, x, y, width, and
// height.
func (chatView *revampedChatView) GetRect() (int, int, int, int) {
	return chatView.Box.GetRect()
}

// SetRect sets a new position of the primitive.
func (chatView *revampedChatView) SetRect(x int, y int, width int, height int) {
	chatView.Box.SetRect(x, y, width, height)
}

func (chatView *revampedChatView) SetParent(parent tview.Primitive) {
	chatView.Box.SetParent(parent)
}

func (chatView *revampedChatView) GetParent() tview.Primitive {
	return chatView.Box.GetParent()
}

// InputHandler returns a handler which receives key events when it has focus.
// It is called by the Application class.
//
// A value of nil may also be returned, in which case this primitive cannot
// receive focus and will not process any key events.
//
// The handler will receive the key event and a function that allows it to
// set the focus to a different primitive, so that future key events are sent
// to that primitive.
//
// The Application's Draw() function will be called automatically after the
// handler returns.
//
// The Box class provides functionality to intercept keyboard input. If you
// subclass from Box, it is recommended that you wrap your handler using
// Box.WrapInputHandler() so you inherit that functionality.
func (chatView *revampedChatView) InputHandler() tview.InputHandlerFunc {
	return chatView.Box.InputHandler()
}

// Focus is called by the application when the primitive receives focus.
// Implementers may call delegate() to pass the focus on to another primitive.
func (chatView *revampedChatView) Focus(delegate func(p tview.Primitive)) {
	chatView.Box.Focus(delegate)
}

// Blur is called by the application when the primitive loses focus.
func (chatView *revampedChatView) Blur() {
	chatView.Box.Blur()
}

// SetOnFocus sets the handler that gets called when Focus() gets called.
func (chatView *revampedChatView) SetOnFocus(handler func()) {
	chatView.Box.SetOnFocus(handler)
}

// SetOnBlur sets the handler that gets called when Blur() gets called.
func (chatView *revampedChatView) SetOnBlur(handler func()) {
	chatView.Box.SetOnBlur(handler)
}

// GetFocusable returns the item's Focusable.
func (chatView *revampedChatView) GetFocusable() tview.Focusable {
	return chatView.Box.GetFocusable()
}

// NextFocusableComponent decides which component should receive focus next.
// If nil is returned, the focus is retained.
func (chatView *revampedChatView) NextFocusableComponent(direction tview.FocusDirection) tview.Primitive {
	return chatView.Box.NextFocusableComponent(direction)
}

func (chatView *revampedChatView) SetInputCapture(handler func(*tcell.EventKey) *tcell.EventKey) {
	chatView.Box.SetInputCapture(handler)
}

func (chatView *revampedChatView) SetMouseHandler(handler func(*tcell.EventMouse) bool) {
	chatView.Box.SetMouseHandler(handler)
}

func (chatView *revampedChatView) SetNextFocusableComponents(direction tview.FocusDirection, components ...tview.Primitive) {
	chatView.Box.SetNextFocusableComponents(direction, components...)
}

func (chatView *revampedChatView) SetBorderSides(top bool, left bool, bottom bool, right bool) {
	chatView.Box.SetBorderSides(top, left, bottom, right)
}

func (chatView *revampedChatView) Dispose() {
	//TODO
}
