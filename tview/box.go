package tview

import (
	"github.com/gdamore/tcell"
)

// Box implements Primitive with a background and optional elements such as a
// border and a title. Most subclasses keep their content contained in the box
// but don't necessarily have to.
//
// Note that all classes which subclass from Box will also have access to its
// functions.
//
// See https://github.com/Bios-Marcel/cordless/tview/wiki/Box for an example.
type Box struct {
	// The position of the rect.
	x, y, width, height int

	// The inner rect reserved for the box's content.
	innerX, innerY, innerWidth, innerHeight int

	// Border padding.
	paddingTop, paddingBottom, paddingLeft, paddingRight int

	// Border Size
	borderTop, borderBottom, borderLeft, borderRight bool

	// The box's background color.
	backgroundColor tcell.Color

	// Reverse video.
	reverse bool

	// Whether or not a border is drawn, reducing the box's space for content by
	// two in width and height.
	border bool

	// The color of the border.
	borderColor tcell.Color

	// The color of the border when the box has focus.
	borderFocusColor tcell.Color

	// The style attributes of the border.
	borderAttributes tcell.AttrMask

	// The style attributes of the border when the box has focus.
	borderFocusAttributes tcell.AttrMask

	// If set to true, the text view will show down and up arrows if there is
	// content out of sight. While box doesn't implement scrolling, this is
	// an abstraction for other components
	indicateOverflow bool

	// The title. Only visible if there is a border, too.
	title string

	// The color of the title.
	titleColor tcell.Color

	// The alignment of the title.
	titleAlign int

	// Provides a way to find out if this box has focus. We always go through
	// this interface because it may be overridden by implementing classes.
	focus Focusable

	// Whether or not this box has focus.
	hasFocus bool

	visible bool

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the primitive's default input handler (nil if
	// nothing should be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	mouseHandler func(event *tcell.EventMouse) bool

	// An optional function which is called before the box is drawn.
	draw func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)

	// Handler that gets called when this component receives focus.
	onFocus func()

	// Handler that gets called when this component loses focus.
	onBlur func()

	nextFocusableComponents map[FocusDirection][]Primitive
	parent                  Primitive
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	b := &Box{
		width:                   15,
		height:                  10,
		innerX:                  -1, // Mark as uninitialized.
		backgroundColor:         Styles.PrimitiveBackgroundColor,
		borderColor:             Styles.BorderColor,
		borderFocusColor:        Styles.BorderFocusColor,
		borderFocusAttributes:   tcell.AttrNone,
		titleColor:              Styles.TitleColor,
		titleAlign:              AlignCenter,
		borderTop:               true,
		borderBottom:            true,
		borderLeft:              true,
		borderRight:             true,
		visible:                 true,
		nextFocusableComponents: make(map[FocusDirection][]Primitive),
	}

	if IsVtxxx {
		b.borderFocusAttributes = tcell.AttrBold
	}

	b.focus = b
	return b
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
	return b
}

// SetVisible sets whether the Box should be drawn onto the screen.
func (b *Box) SetVisible(visible bool) {
	b.visible = visible
}

// IsVisible gets whether the Box should be drawn onto the screen.
func (b *Box) IsVisible() bool {
	return b.visible
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) GetRect() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// SetOnFocus sets the handler that gets called when Focus() gets called.
func (b *Box) SetOnFocus(handler func()) {
	b.onFocus = handler
}

// SetOnBlur sets the handler that gets called when Blur() gets called.
func (b *Box) SetOnBlur(handler func()) {
	b.onBlur = handler
}

// GetInnerRect returns the position of the inner rectangle (x, y, width,
// height), without the border and without any padding. Width and height values
// will clamp to 0 and thus never be negative.
func (b *Box) GetInnerRect() (int, int, int, int) {
	if b.innerX >= 0 {
		return b.innerX, b.innerY, b.innerWidth, b.innerHeight
	}
	x, y, width, height := b.GetRect()
	if b.border {
		x += boolToInt(b.borderLeft)
		y += boolToInt(b.borderTop)
		width -= boolToInt(b.borderLeft) + boolToInt(b.borderRight)
		height -= boolToInt(b.borderTop) + boolToInt(b.borderBottom)
	}
	x, y, width, height = x+b.paddingLeft,
		y+b.paddingTop,
		width-b.paddingLeft-b.paddingRight,
		height-b.paddingTop-b.paddingBottom
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return x, y, width, height
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

// SetRect sets a new position of the primitive. Note that this has no effect
// if this primitive is part of a layout (e.g. Flex, Grid) or if it was added
// like this:
//
//   application.SetRoot(b, true)
func (b *Box) SetRect(x, y, width, height int) {
	b.x = x
	b.y = y
	b.width = width
	b.height = height
	b.innerX = -1 // Mark inner rect as uninitialized.
}

// SetDrawFunc sets a callback function which is invoked after the box primitive
// has been drawn. This allows you to add a more individual style to the box
// (and all primitives which extend it).
//
// The function is provided with the box's dimensions (set via SetRect()). It
// must return the box's inner dimensions (x, y, width, height) which will be
// returned by GetInnerRect(), used by descendent primitives to draw their own
// content.
func (b *Box) SetDrawFunc(handler func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)) *Box {
	b.draw = handler
	return b
}

// GetDrawFunc returns the callback function which was installed with
// SetDrawFunc() or nil if no such function has been installed.
func (b *Box) GetDrawFunc() func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	return b.draw
}

// WrapInputHandler wraps an input handler (see InputHandler()) with the
// functionality to capture input (see SetInputCapture()) before passing it
// on to the provided (default) input handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapInputHandler(inputHandler InputHandlerFunc) InputHandlerFunc {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) *tcell.EventKey {
		if b.inputCapture != nil {
			event = b.inputCapture(event)
		}
		if event != nil && inputHandler != nil {
			event = inputHandler(event, setFocus)
		}

		return event
	}
}

// InputHandler returns nil.
func (b *Box) InputHandler() InputHandlerFunc {
	return b.WrapInputHandler(nil)
}

// SetInputCapture installs a function which captures key events before they are
// forwarded to the primitive's default key event handler. This function can
// then choose to forward that key event (or a different one) to the default
// handler by returning it. If nil is returned, the default handler will not
// be called.
//
// Providing a nil handler will remove a previously existing handler.
//
// Note that this function will not have an effect on primitives composed of
// other primitives, such as Form, Flex, or Grid. Key events are only captured
// by the primitives that have focus (e.g. InputField) and only one primitive
// can have focus at a time. Composing primitives such as Form pass the focus on
// to their contained primitives and thus never receive any key events
// themselves. Therefore, they cannot intercept key events.
func (b *Box) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Box {
	b.inputCapture = capture
	return b
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (b *Box) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return b.inputCapture
}

// SetMouseHandler sets the mouse event handler.
func (b *Box) SetMouseHandler(handler func(event *tcell.EventMouse) bool) {
	b.mouseHandler = handler
}

// MouseHandler returns the mouse event handler or nil if none is present.
func (b *Box) MouseHandler() func(event *tcell.EventMouse) bool {
	return b.mouseHandler
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	b.backgroundColor = color
	return b
}

// SetReverse turns on or off the reverse video attribute.
func (b *Box) SetReverse(on bool) *Box {
	b.reverse = on
	return b
}

// SetBorder sets the flag indicating whether or not the box should have a
// border.
func (b *Box) SetBorder(show bool) *Box {
	b.border = show
	return b
}

// SetBorderColor sets the box's border color.
func (b *Box) SetBorderColor(color tcell.Color) *Box {
	b.borderColor = color
	return b
}

// SetBorderFocusColor sets the box's border color when focused.
func (b *Box) SetBorderFocusColor(color tcell.Color) *Box {
	b.borderFocusColor = color
	return b
}

// SetBorderSides decides which sides of the border should be shown in case the
// border has been activated.
func (b *Box) SetBorderSides(top, left, bottom, right bool) *Box {
	b.borderTop = top
	b.borderLeft = left
	b.borderBottom = bottom
	b.borderRight = right

	return b
}

// IsBorder indicates whether a border is rendered at all.
func (b *Box) IsBorder() bool {
	return b.border
}

// IsBorderTop indicates whether a border is rendered on the top side of
// this primitive.
func (b *Box) IsBorderTop() bool {
	return b.border && b.borderTop
}

// IsBorderBottom indicates whether a border is rendered on the bottom side
// of this primitive.
func (b *Box) IsBorderBottom() bool {
	return b.border && b.borderBottom
}

// IsBorderRight indicates whether a border is rendered on the right side of
// this primitive.
func (b *Box) IsBorderRight() bool {
	return b.border && b.borderRight
}

// IsBorderLeft indicates whether a border is rendered on the left side of
// this primitive.
func (b *Box) IsBorderLeft() bool {
	return b.border && b.borderLeft
}

// SetBorderAttributes sets the border's style attributes. You can combine
// different attributes using bitmask operations:
//
//   box.SetBorderAttributes(tcell.AttrUnderline | tcell.AttrBold)
func (b *Box) SetBorderAttributes(attr tcell.AttrMask) *Box {
	b.borderAttributes = attr
	return b
}

// SetBorderFocusAttributes sets the border's style attributes when focused. You can combine
// different attributes using bitmask operations:
//
//   box.SetBorderFocusAttributes(tcell.AttrUnderline | tcell.AttrBold)
func (b *Box) SetBorderFocusAttributes(attr tcell.AttrMask) *Box {
	b.borderFocusAttributes = attr
	return b
}

// SetTitle sets the box's title.
func (b *Box) SetTitle(title string) *Box {
	b.title = title
	return b
}

// SetTitleColor sets the box's title color.
func (b *Box) SetTitleColor(color tcell.Color) *Box {
	b.titleColor = color
	return b
}

// SetTitleAlign sets the alignment of the title, one of AlignLeft, AlignCenter,
// or AlignRight.
func (b *Box) SetTitleAlign(align int) *Box {
	b.titleAlign = align
	return b
}

// Draw draws this primitive onto the screen.
func (b *Box) Draw(screen tcell.Screen) bool {
	// Don't draw anything if there is no space.
	if b.width <= 0 || b.height <= 0 || !b.visible {
		return false
	}

	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(b.backgroundColor).Reverse(b.reverse)
	for y := b.y; y < b.y+b.height; y++ {
		for x := b.x; x < b.x+b.width; x++ {
			screen.SetContent(x, y, ' ', nil, background)
		}
	}

	// Draw border.
	if b.border && b.width >= 2 && b.height >= 1 {
		var border tcell.Style
		if b.hasFocus {
			if b.borderFocusAttributes != 0 {
				border = background.Foreground(b.borderFocusColor) | tcell.Style(b.borderFocusAttributes)
			} else {
				border = background.Foreground(b.borderFocusColor) | tcell.Style(b.borderAttributes)
			}
		} else {
			border = background.Foreground(b.borderColor) | tcell.Style(b.borderAttributes)
		}
		var vertical, horizontal, topLeft, topRight, bottomLeft, bottomRight rune

		horizontal = Borders.Horizontal
		vertical = Borders.Vertical
		topLeft = Borders.TopLeft
		topRight = Borders.TopRight
		bottomLeft = Borders.BottomLeft
		bottomRight = Borders.BottomRight

		//Special case in order to render only the title-line of something properly.
		if b.borderTop {
			for x := b.x + 1; x < b.x+b.width-1; x++ {
				screen.SetContent(x, b.y, horizontal, nil, border)
			}

			if b.borderLeft {
				screen.SetContent(b.x, b.y, topLeft, nil, border)
			} else {
				screen.SetContent(b.x, b.y, horizontal, nil, border)
			}

			if b.borderRight {
				screen.SetContent(b.x+b.width-1, b.y, topRight, nil, border)
			} else {
				screen.SetContent(b.x+b.width-1, b.y, horizontal, nil, border)
			}
		}

		//Special case in order to render only the title-line of something properly.
		if b.height > 1 {
			if b.borderBottom {
				for x := b.x + 1; x < b.x+b.width-1; x++ {
					screen.SetContent(x, b.y+b.height-1, horizontal, nil, border)
				}

				if b.borderLeft {
					screen.SetContent(b.x, b.y+b.height-1, bottomLeft, nil, border)
				} else {
					screen.SetContent(b.x, b.y+b.height-1, horizontal, nil, border)
				}
				if b.borderRight {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, bottomRight, nil, border)
				} else {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, horizontal, nil, border)
				}
			}

			if b.borderLeft {
				for y := b.y + 1; y < b.y+b.height-1; y++ {
					screen.SetContent(b.x, y, vertical, nil, border)
				}

				if b.borderTop {
					screen.SetContent(b.x, b.y, topLeft, nil, border)
				} else {
					screen.SetContent(b.x, b.y, vertical, nil, border)
				}

				if b.borderBottom {
					screen.SetContent(b.x, b.y+b.height-1, bottomLeft, nil, border)
				} else {
					screen.SetContent(b.x, b.y+b.height-1, vertical, nil, border)
				}
			}

			if b.borderRight {
				for y := b.y + 1; y < b.y+b.height-1; y++ {
					screen.SetContent(b.x+b.width-1, y, vertical, nil, border)
				}

				if b.borderTop {
					screen.SetContent(b.x+b.width-1, b.y, topRight, nil, border)
				} else {
					screen.SetContent(b.x+b.width-1, b.y, vertical, nil, border)
				}

				if b.borderBottom {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, bottomRight, nil, border)
				} else {
					screen.SetContent(b.x+b.width-1, b.y+b.height-1, vertical, nil, border)
				}
			}
		} else if b.height == 1 && !b.borderTop && !b.borderBottom {
			if b.borderLeft {
				screen.SetContent(b.x, b.y, vertical, nil, border)
			}
			if b.borderRight {
				screen.SetContent(b.x+b.width-1, b.y+b.height-1, vertical, nil, border)
			}
		}

		// Draw title.
		if b.title != "" && b.width >= 4 {
			printed, _ := Print(screen, b.title, b.x+1, b.y, b.width-2, b.titleAlign, b.titleColor)
			if len(b.title)-printed > 0 && printed > 0 {
				_, _, style, _ := screen.GetContent(b.x+b.width-2, b.y)
				fg, _, _ := style.Decompose()
				Print(screen, string(SemigraphicsHorizontalEllipsis), b.x+b.width-2, b.y, 1, AlignLeft, fg)
			}
		}
	}

	// Call custom draw function.
	if b.draw != nil {
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.draw(screen, b.x, b.y, b.width, b.height)
	} else {
		// Remember the inner rect.
		b.innerX = -1
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.GetInnerRect()
	}

	// Clamp inner rect to screen.
	width, height := screen.Size()
	if b.innerX < 0 {
		b.innerWidth += b.innerX
		b.innerX = 0
	}
	if b.innerX+b.innerWidth >= width {
		b.innerWidth = width - b.innerX
	}
	if b.innerY+b.innerHeight >= height {
		b.innerHeight = height - b.innerY
	}
	if b.innerY < 0 {
		b.innerHeight += b.innerY
		b.innerY = 0
	}

	if b.innerWidth < 0 {
		b.innerWidth = 0
	}
	if b.innerHeight < 0 {
		b.innerHeight = 0
	}

	return true
}

// Focus is called when this primitive receives focus.
func (b *Box) Focus(delegate func(p Primitive)) {
	b.hasFocus = true
	if b.onFocus != nil {
		b.onFocus()
	}
}

// Blur is called when this primitive loses focus.
func (b *Box) Blur() {
	b.hasFocus = false
	if b.onBlur != nil {
		b.onBlur()
	}
}

// SetNextFocusableComponents decides which components are to be focused using
// a certain focus direction. If more than one component is passed, the
// priority goes from left-most to right-most. A component will be skipped if
// it is not visible.
func (b *Box) SetNextFocusableComponents(direction FocusDirection, components ...Primitive) {
	b.nextFocusableComponents[direction] = components
}

// NextFocusableComponent decides which component should receive focus next.
// If nil is returned, the focus is retained.
func (b *Box) NextFocusableComponent(direction FocusDirection) Primitive {
	components, avail := b.nextFocusableComponents[direction]
	if avail {
		for _, comp := range components {
			if comp.IsVisible() {
				return comp
			}
		}
	}

	return nil
}

// HasFocus returns whether or not this primitive has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}

// GetFocusable returns the item's Focusable.
func (b *Box) GetFocusable() Focusable {
	return b.focus
}

// SetIndicateOverflow toggles whether overflow arrows can be drawn in order to
// signal that the component contains content that is out of the viewarea.
func (b *Box) SetIndicateOverflow(indicateOverflow bool) *Box {
	b.indicateOverflow = indicateOverflow
	return b
}

func (b *Box) SetParent(parent Primitive) {
	//Reparenting is possible!
	b.parent = parent
}

func (b *Box) GetParent() Primitive {
	return b.parent
}

func (b *Box) drawOverflow(screen tcell.Screen, showTop, showBottom bool) {
	if b.indicateOverflow && b.border && b.borderTop && b.borderBottom && b.height > 1 {
		overflowIndicatorX := b.innerX + b.innerWidth + b.paddingRight - 1
		style := tcell.StyleDefault.Foreground(Styles.InverseTextColor).Background(b.backgroundColor)
		if showTop {
			screen.SetContent(overflowIndicatorX, b.innerY-b.paddingTop-1, '▲', nil, style)
		}
		if showBottom {
			screen.SetContent(overflowIndicatorX, b.innerY+b.innerHeight+b.paddingBottom, '▼', nil, style)
		}
	}
}
