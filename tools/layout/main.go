package main

import (
	"github.com/Bios-Marcel/cordless/tview"
	"github.com/gdamore/tcell"
)

func main() {
	app := tview.NewApplication()
	err := app.Run()
	if err != nil {
		panic(err)
	}
}

// CoreLayout implements tview.Primitive and represents the core component layout
// of the application, handling not only the layouting, but also the borders.
type CoreLayout struct {
	visible       bool
	x, y          int
	width, height int

	dm       tview.Primitive
	servers  tview.Primitive
	channels tview.Primitive
	messages tview.Primitive
	input    tview.Primitive
	terminal tview.Primitive
	userList tview.Primitive
}

func NewCoreLayout() *CoreLayout {
	return &CoreLayout{
		visible:  true,
		dm:       NewDemoComponent('d'),
		servers:  NewDemoComponent('s'),
		channels: NewDemoComponent('c'),
		messages: NewDemoComponent('m'),
		input:    NewDemoComponent('i'),
		terminal: NewDemoComponent('t'),
		userList: NewDemoComponent('u'),
	}
}

func (c *CoreLayout) Draw(screen tcell.Screen) bool {
	return true
}

func (c *CoreLayout) SetVisible(visible bool) {
	c.visible = visible
}

func (c *CoreLayout) IsVisible() bool {
	return c.visible
}

func (c *CoreLayout) GetRect() (int, int, int, int) {
	return c.x, c.y, c.width, c.height
}

func (c *CoreLayout) SetRect(x, y, width, height int) {
	c.x = x
	c.y = y
	c.width = width
	c.height = height
}

func (c *CoreLayout) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	panic("implement me")
}

func (c *CoreLayout) Focus(delegate func(p tview.Primitive)) {
	panic("implement me")
}

func (c *CoreLayout) Blur() {
	panic("implement me")
}

func (c *CoreLayout) SetOnFocus(handler func()) {
	panic("implement me")
}

func (c *CoreLayout) SetOnBlur(handler func()) {
	panic("implement me")
}

func (c *CoreLayout) GetFocusable() tview.Focusable {
	panic("implement me")
}

func (c *CoreLayout) NextFocusableComponent(direction tview.FocusDirection) tview.Primitive {
	panic("implement me")
}

func NewDemoComponent(r rune) *demoComponent {
	return &demoComponent{visible: true, fillWith: r}
}

type demoComponent struct {
	visible       bool
	x, y          int
	width, height int

	fillWith rune
}

func genRuneArray(r rune, length int) []rune {
	array := make([]rune, length, length)
	for i := 0; i < length; i++ {
		array[i] = r
	}
	return array
}

func (d *demoComponent) Draw(screen tcell.Screen) bool {
	if !d.visible {
		return false
	}

	for h := 0; h < d.height; h++ {
		screen.SetCell(d.x, d.y+h, tcell.StyleDefault, genRuneArray(d.fillWith, d.width)...)
	}

	return true
}

func (d *demoComponent) SetVisible(visible bool) {
	d.visible = visible
}

func (d *demoComponent) IsVisible() bool {
	return d.visible
}

func (d *demoComponent) GetRect() (int, int, int, int) {
	return d.x, d.y, d.width, d.height
}

func (d *demoComponent) SetRect(x, y, width, height int) {
	d.x = x
	d.y = y
	d.width = width
	d.height = height
}

func (d *demoComponent) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

	}
}

func (d *demoComponent) Focus(delegate func(p tview.Primitive)) {
	//Do nothing
}

func (d *demoComponent) Blur() {
	//Do nothing
}

func (d *demoComponent) SetOnFocus(handler func()) {
	//Do nothing
}

func (d *demoComponent) SetOnBlur(handler func()) {
	//Do nothing
}

func (d *demoComponent) GetFocusable() tview.Focusable {
	return nil
}

func (d *demoComponent) NextFocusableComponent(direction tview.FocusDirection) tview.Primitive {
	return nil
}
