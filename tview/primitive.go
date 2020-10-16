package tview

import "github.com/gdamore/tcell"

// Primitive is the top-most interface for all graphical primitives.
type Primitive interface {
	// Draw draws this primitive onto the screen. Implementers can call the
	// screen's ShowCursor() function but should only do so when they have focus.
	// (They will need to keep track of this themselves.)
	Draw(screen tcell.Screen) bool

	// Sets whether the primitive should be drawn onto the screen.
	SetVisible(bool)

	// Gets whether the primitive should be drawn onto the screen.
	IsVisible() bool

	// GetRect returns the current position of the primitive, x, y, width, and
	// height.
	GetRect() (int, int, int, int)

	// SetRect sets a new position of the primitive.
	SetRect(x, y, width, height int)

	SetParent(Primitive)

	GetParent() Primitive

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
	InputHandler() InputHandlerFunc

	// Focus is called by the application when the primitive receives focus.
	// Implementers may call delegate() to pass the focus on to another primitive.
	Focus(delegate func(p Primitive))

	// Blur is called by the application when the primitive loses focus.
	Blur()

	// SetOnFocus sets the handler that gets called when Focus() gets called.
	SetOnFocus(handler func())

	// SetOnBlur sets the handler that gets called when Blur() gets called.
	SetOnBlur(handler func())

	// GetFocusable returns the item's Focusable.
	GetFocusable() Focusable

	// NextFocusableComponent decides which component should receive focus next.
	// If nil is returned, the focus is retained.
	NextFocusableComponent(FocusDirection) Primitive
}

type InputHandlerFunc func(*tcell.EventKey, func(p Primitive)) *tcell.EventKey

// FocusDirection decides in what direction the focus should travel relative
// to the currently focused component.
type FocusDirection int

const (
	Up FocusDirection = iota
	Down
	Left
	Right
)
