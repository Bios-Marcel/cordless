package tview

import (
	"sort"
	"strings"
	"time"

	tcell "github.com/gdamore/tcell/v2"
)

// Tree navigation events.
const (
	treeNone int = iota
	treeHome
	treeEnd
	treeUp
	treeDown
	treePageUp
	treePageDown
)

// TreeNode represents one node in a tree view.
type TreeNode struct {
	// The reference object.
	reference interface{}

	// This node's child nodes.
	children []*TreeNode

	// The item's text.
	text string

	// This text is a prefixes in front of the normal text.
	prefixes []string

	// The text color.
	color tcell.Color

	blinking bool

	underline bool

	// Whether or not this node can be selected.
	selectable bool

	// Whether or not this node's children should be displayed.
	expanded bool

	// The additional horizontal indent of this node's text.
	indent int

	// An optional function which is called when the user selects this node.
	selected func()

	// Temporary member variables.
	parent    *TreeNode // The parent node (nil for the root).
	level     int       // The hierarchy level (0 for the root, 1 for its children, and so on).
	graphicsX int       // The x-coordinate of the left-most graphics rune.
	textX     int       // The x-coordinate of the first rune of the text.
}

// NewTreeNode returns a new tree node.
func NewTreeNode(text string) *TreeNode {
	return &TreeNode{
		text:       text,
		color:      Styles.PrimaryTextColor,
		indent:     2,
		expanded:   true,
		selectable: true,
	}
}

// Walk traverses this node's subtree in depth-first, pre-order (NLR) order and
// calls the provided callback function on each traversed node (which includes
// this node) with the traversed node and its parent node (nil for this node).
// The callback returns whether traversal should continue with the traversed
// node's child nodes (true) or not recurse any deeper (false).
func (n *TreeNode) Walk(callback func(node, parent *TreeNode) bool) *TreeNode {
	n.parent = nil
	nodes := []*TreeNode{n}
	for len(nodes) > 0 {
		// Pop the top node and process it.
		node := nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
		if !callback(node, node.parent) {
			// Don't add any children.
			continue
		}

		// Add children in reverse order.
		for index := len(node.children) - 1; index >= 0; index-- {
			node.children[index].parent = node
			nodes = append(nodes, node.children[index])
		}
	}

	return n
}

// SetReference allows you to store a reference of any type in this node. This
// will allow you to establish a mapping between the TreeView hierarchy and your
// internal tree structure.
func (n *TreeNode) SetReference(reference interface{}) *TreeNode {
	n.reference = reference
	return n
}

// GetReference returns this node's reference object.
func (n *TreeNode) GetReference() interface{} {
	return n.reference
}

// SetChildren sets this node's child nodes.
func (n *TreeNode) SetChildren(childNodes []*TreeNode) *TreeNode {
	n.children = childNodes
	return n
}

// GetText returns this node's text.
func (n *TreeNode) GetText() string {
	return n.text
}

// GetPrefixes returns this node's prefix text.
func (n *TreeNode) GetPrefixes() []string {
	return n.prefixes
}

// GetParent returns a refrence to this nodes parent node or nil if this is a
// root node.
func (n *TreeNode) GetParent() *TreeNode {
	return n.parent
}

// GetChildren returns this node's children.
func (n *TreeNode) GetChildren() []*TreeNode {
	return n.children
}

// ClearChildren removes all child nodes from this node.
func (n *TreeNode) ClearChildren() *TreeNode {
	n.children = nil
	return n
}

// AddChild adds a new child node to this node.
func (n *TreeNode) AddChild(node *TreeNode) *TreeNode {
	n.children = append(n.children, node)
	return n
}

// SetSelectable sets a flag indicating whether this node can be selected by
// the user.
func (n *TreeNode) SetSelectable(selectable bool) *TreeNode {
	n.selectable = selectable
	return n
}

// IsSelectable returns wether the node is selectable or not.
func (n *TreeNode) IsSelectable() bool {
	return n.selectable
}

// SetSelectedFunc sets a function which is called when the user selects this
// node by hitting Enter when it is selected.
func (n *TreeNode) SetSelectedFunc(handler func()) *TreeNode {
	n.selected = handler
	return n
}

// SetExpanded sets whether or not this node's child nodes should be displayed.
func (n *TreeNode) SetExpanded(expanded bool) *TreeNode {
	n.expanded = expanded
	return n
}

// Expand makes the child nodes of this node appear.
func (n *TreeNode) Expand() *TreeNode {
	n.expanded = true
	return n
}

// Collapse makes the child nodes of this node disappear.
func (n *TreeNode) Collapse() *TreeNode {
	n.expanded = false
	return n
}

// ExpandAll expands this node and all descendent nodes.
func (n *TreeNode) ExpandAll() *TreeNode {
	n.Walk(func(node, parent *TreeNode) bool {
		node.expanded = true
		return true
	})
	return n
}

// CollapseAll collapses this node and all descendent nodes.
func (n *TreeNode) CollapseAll() *TreeNode {
	n.Walk(func(node, parent *TreeNode) bool {
		n.expanded = false
		return true
	})
	return n
}

// IsExpanded returns whether the child nodes of this node are visible.
func (n *TreeNode) IsExpanded() bool {
	return n.expanded
}

// SetText sets the node's text which is displayed.
func (n *TreeNode) SetText(text string) *TreeNode {
	n.text = text
	return n
}

// AddPrefix sets the node's prefix text which is displayed. Duplicates are
// ignored.
func (n *TreeNode) AddPrefix(newPrefix string) *TreeNode {
	for _, prefix := range n.prefixes {
		if prefix == newPrefix {
			return n
		}
	}

	n.prefixes = append(n.prefixes, newPrefix)
	return n
}

// RemovePrefix removes the given prefix, maintaining the order of the items.
func (n *TreeNode) RemovePrefix(prefix string) {
	for removeIndex, oldPrefix := range n.prefixes {
		if oldPrefix == prefix {
			n.prefixes = append(n.prefixes[:removeIndex], n.prefixes[removeIndex+1:]...)
			break
		}
	}
}

// ClearPrefixes removes all prefixes by nulling the underlying array.
func (n *TreeNode) ClearPrefixes() {
	n.prefixes = nil
}

// SortPrefixes sorts all currently set prefixes. This function sin't executed
// once new prefixes are added.
func (n *TreeNode) SortPrefixes(lessFunction func(a, b string) bool) {
	sort.Slice(n.prefixes, func(a, b int) bool {
		aItem := n.prefixes[a]
		bItem := n.prefixes[b]
		return lessFunction(aItem, bItem)
	})
}

// GetColor returns the node's color.
func (n *TreeNode) GetColor() tcell.Color {
	return n.color
}

// SetColor sets the node's text color.
func (n *TreeNode) SetColor(color tcell.Color) *TreeNode {
	n.color = color
	return n
}

func (n *TreeNode) SetUnderline(underline bool) {
	n.underline = underline
}

func (n *TreeNode) SetBlinking(blinking bool) {
	n.blinking = blinking
}

// SetIndent sets an additional indentation for this node's text. A value of 0
// keeps the text as far left as possible with a minimum of line graphics. Any
// value greater than that moves the text to the right.
func (n *TreeNode) SetIndent(indent int) *TreeNode {
	n.indent = indent
	return n
}

// TreeView displays tree structures. A tree consists of nodes (TreeNode
// objects) where each node has zero or more child nodes and exactly one parent
// node (except for the root node which has no parent node).
//
// The SetRoot() function is used to specify the root of the tree. Other nodes
// are added locally to the root node or any of its descendents. See the
// TreeNode documentation for details on node attributes. (You can use
// SetReference() to store a reference to nodes of your own tree structure.)
//
// Nodes can be selected by calling SetCurrentNode(). The user can navigate the
// selection or the tree by using the following keys:
//
//   - j, down arrow, right arrow: Move (the selection) down by one node.
//   - k, up arrow, left arrow: Move (the selection) up by one node.
//   - g, home: Move (the selection) to the top.
//   - G, end: Move (the selection) to the bottom.
//   - Ctrl-F, page down: Move (the selection) down by one page.
//   - Ctrl-B, page up: Move (the selection) up by one page.
//
// Selected nodes can trigger the "selected" callback when the user hits Enter.
//
// The root node corresponds to level 0, its children correspond to level 1,
// their children to level 2, and so on. Per default, the first level that is
// displayed is 0, i.e. the root node. You can call SetTopLevel() to hide
// levels.
//
// If graphics are turned on (see SetGraphics()), lines indicate the tree's
// hierarchy. Alternative (or additionally), you can set different prefixes
// using SetPrefixes() for different levels, for example to display hierarchical
// bullet point lists.
//
// See https://github.com/Bios-Marcel/cordless/tview/wiki/TreeView for an example.
type TreeView struct {
	*Box

	// The root node.
	root *TreeNode

	// The currently selected node or nil if no node is selected.
	currentNode *TreeNode

	// The movement to be performed during the call to Draw(), one of the
	// constants defined above.
	movement int

	// The top hierarchical level shown. (0 corresponds to the root level.)
	topLevel int

	// Strings drawn before the nodes, based on their level.
	bulletCharacters []string

	// This decides whether the selection will cycle when reaching the end
	// or the beginning of the tree.
	cycleSelection bool

	// Vertical scroll offset.
	offsetY int

	// If set to true, all node texts will be aligned horizontally.
	align bool

	// If set to true, the tree structure is drawn using lines.
	graphics bool

	// The color of the lines.
	graphicsColor tcell.Color

	// An optional function which is called when the user has navigated to a new
	// tree node.
	changed func(node *TreeNode)

	// An optional function which is called when a tree item was selected.
	selected func(node *TreeNode)

	// The visible nodes, top-down, as set by process().
	nodes []*TreeNode

	// Decides wether it allows usage of vim bindings for navigation.
	vimBindings bool

	// Decides wether a search and selection will be triggered on rune input.
	searchOnType bool

	jumpTime   time.Time
	jumpBuffer string
}

// NewTreeView returns a new tree view.
func NewTreeView() *TreeView {
	return &TreeView{
		Box:           NewBox(),
		graphics:      true,
		graphicsColor: Styles.GraphicsColor,
		searchOnType:  true,
		vimBindings:   false,
	}
}

// SetRoot sets the root node of the tree.
func (t *TreeView) SetRoot(root *TreeNode) *TreeView {
	t.root = root
	return t
}

// GetRoot returns the root node of the tree. If no such node was previously
// set, nil is returned.
func (t *TreeView) GetRoot() *TreeNode {
	return t.root
}

// SetCurrentNode sets the currently selected node. Provide nil to clear all
// selections. Selected nodes must be visible and selectable, or else the
// selection will be changed to the top-most selectable and visible node.
//
// This function does NOT trigger the "changed" callback.
func (t *TreeView) SetCurrentNode(node *TreeNode) *TreeView {
	t.currentNode = node
	return t
}

// GetCurrentNode returns the currently selected node or nil of no node is
// currently selected.
func (t *TreeView) GetCurrentNode() *TreeNode {
	return t.currentNode
}

// SetTopLevel sets the first tree level that is visible with 0 referring to the
// root, 1 to the root's child nodes, and so on. Nodes above the top level are
// not displayed.
func (t *TreeView) SetTopLevel(topLevel int) *TreeView {
	t.topLevel = topLevel
	return t
}

// SetBulletCharacters defines the strings drawn before the nodes' texts.
// This is a slice of strings where each element corresponds to a node's
// hierarchy level, i.e. 0 for the root, 1 for the root's children, and
// so on (levels will cycle).
//
// For example, to display a hierarchical list with bullet points:
//
//   treeView.SetGraphics(false).
//     SetBulletCharacters([]string{"* ", "- ", "x "})
func (t *TreeView) SetBulletCharacters(prefixes []string) *TreeView {
	t.bulletCharacters = prefixes
	return t
}

// SetAlign controls the horizontal alignment of the node texts. If set to true,
// all texts except that of top-level nodes will be placed in the same column.
// If set to false, they will indent with the hierarchy.
func (t *TreeView) SetAlign(align bool) *TreeView {
	t.align = align
	return t
}

// SetCycleSelection controls the selection behaviour when there is no more
// node that is eligible for selection. For example if you are index 1 and
// index 0 isn't selectable anymore, it will try selecting the last index.
func (t *TreeView) SetCycleSelection(cycleSelection bool) *TreeView {
	t.cycleSelection = cycleSelection
	return t
}

// SetGraphics sets a flag which determines whether or not line graphics are
// drawn to illustrate the tree's hierarchy.
func (t *TreeView) SetGraphics(showGraphics bool) *TreeView {
	t.graphics = showGraphics
	return t
}

// SetGraphicsColor sets the colors of the lines used to draw the tree structure.
func (t *TreeView) SetGraphicsColor(color tcell.Color) *TreeView {
	t.graphicsColor = color
	return t
}

// SetChangedFunc sets the function which is called when the user navigates to
// a new tree node.
func (t *TreeView) SetChangedFunc(handler func(node *TreeNode)) *TreeView {
	t.changed = handler
	return t
}

// SetVimBindingsEnabled decides wether the usage of vim bindings for
// navigation is possible or not. This setting disables the search on
// rune input.
func (t *TreeView) SetVimBindingsEnabled(enabled bool) *TreeView {
	t.vimBindings = enabled

	if enabled {
		t.searchOnType = false
	}

	return t
}

// SetSearchOnTypeEnabled enables / disables search through the tree on rune
// input.
func (t *TreeView) SetSearchOnTypeEnabled(enabled bool) *TreeView {
	t.searchOnType = enabled

	if enabled {
		t.vimBindings = false
	}

	return t
}

// SetSelectedFunc sets the function which is called when the user selects a
// node by pressing Enter on the current selection.
func (t *TreeView) SetSelectedFunc(handler func(node *TreeNode)) *TreeView {
	t.selected = handler
	return t
}

// GetScrollOffset returns the number of node rows that were skipped at the top
// of the tree view. Note that when the user navigates the tree view, this value
// is only updated after the tree view has been redrawn.
func (t *TreeView) GetScrollOffset() int {
	return t.offsetY
}

// GetRowCount returns the number of "visible" nodes. This includes nodes which
// fall outside the tree view's box but notably does not include the children
// of collapsed nodes. Note that this value is only up to date after the tree
// view has been drawn.
func (t *TreeView) GetRowCount() int {
	return len(t.nodes)
}

// process builds the visible tree, populates the "nodes" slice, and processes
// pending selection actions.
func (t *TreeView) process() {
	_, _, _, height := t.GetInnerRect()

	// Determine visible nodes and their placement.
	var graphicsOffset, maxTextX int
	t.nodes = nil
	selectedIndex := -1
	topLevelGraphicsX := -1
	if t.graphics {
		graphicsOffset = 1
	}
	t.root.Walk(func(node, parent *TreeNode) bool {
		// Set node attributes.
		node.parent = parent
		if parent == nil {
			node.level = 0
			node.graphicsX = 0
			node.textX = 0
		} else {
			node.level = parent.level + 1
			node.graphicsX = parent.textX
			node.textX = node.graphicsX + graphicsOffset + node.indent
		}
		if !t.graphics && t.align {
			// Without graphics, we align nodes on the first column.
			node.textX = 0
		}
		if node.level == t.topLevel {
			// No graphics for top level nodes.
			node.graphicsX = 0
			node.textX = 0
		}

		// Add the node to the list.
		if node.level >= t.topLevel {
			// This node will be visible.
			if node.textX > maxTextX {
				maxTextX = node.textX
			}
			if node == t.currentNode && node.selectable {
				selectedIndex = len(t.nodes)
			}

			// Maybe we want to skip this level.
			if t.topLevel == node.level && (topLevelGraphicsX < 0 || node.graphicsX < topLevelGraphicsX) {
				topLevelGraphicsX = node.graphicsX
			}

			t.nodes = append(t.nodes, node)
		}

		// Recurse if desired.
		return node.expanded
	})

	// Post-process positions.
	for _, node := range t.nodes {
		// If text must align, we correct the positions.
		if t.align && node.level > t.topLevel {
			node.textX = maxTextX
		}

		// If we skipped levels, shift to the left.
		if topLevelGraphicsX > 0 {
			node.graphicsX -= topLevelGraphicsX
			node.textX -= topLevelGraphicsX
		}
	}

	// Process selection. (Also trigger events if necessary.)
	if selectedIndex >= 0 {
		// Move the selection.
		newSelectedIndex := selectedIndex
	MovementSwitch:
		switch t.movement {
		case treeUp:
			for newSelectedIndex > 0 {
				newSelectedIndex--
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}

			//Cycle to the bottom if we are at the top.
			if t.cycleSelection {
				newSelectedIndex = len(t.nodes)
				for newSelectedIndex > selectedIndex {
					newSelectedIndex--
					if t.nodes[newSelectedIndex].selectable {
						break MovementSwitch
					}
				}
			}

			newSelectedIndex = selectedIndex

		case treeDown:
			for newSelectedIndex < len(t.nodes)-1 {
				newSelectedIndex++
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}

			//Cycle to the top if we are at the bottom.
			if t.cycleSelection {
				newSelectedIndex = -1
				for newSelectedIndex < selectedIndex {
					newSelectedIndex++
					if t.nodes[newSelectedIndex].selectable {
						break MovementSwitch
					}
				}
			}

			newSelectedIndex = selectedIndex
		case treeHome:
			for newSelectedIndex = 0; newSelectedIndex < len(t.nodes); newSelectedIndex++ {
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}
			newSelectedIndex = selectedIndex
		case treeEnd:
			for newSelectedIndex = len(t.nodes) - 1; newSelectedIndex >= 0; newSelectedIndex-- {
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}
			newSelectedIndex = selectedIndex
		case treePageDown:
			if newSelectedIndex+height < len(t.nodes) {
				newSelectedIndex += height
			} else {
				newSelectedIndex = len(t.nodes) - 1
			}
			for ; newSelectedIndex < len(t.nodes); newSelectedIndex++ {
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}
			newSelectedIndex = selectedIndex
		case treePageUp:
			if newSelectedIndex >= height {
				newSelectedIndex -= height
			} else {
				newSelectedIndex = 0
			}
			for ; newSelectedIndex >= 0; newSelectedIndex-- {
				if t.nodes[newSelectedIndex].selectable {
					break MovementSwitch
				}
			}
			newSelectedIndex = selectedIndex
		}
		t.currentNode = t.nodes[newSelectedIndex]
		if newSelectedIndex != selectedIndex {
			t.movement = treeNone
			if t.changed != nil {
				t.changed(t.currentNode)
			}
		}
		selectedIndex = newSelectedIndex

		// Move selection into viewport.
		if selectedIndex-t.offsetY >= height {
			t.offsetY = selectedIndex - height + 1
		}
		if selectedIndex < t.offsetY {
			t.offsetY = selectedIndex
		}
	} else {
		// If selection is not visible or selectable, select the first candidate.
		if t.currentNode != nil {
			for index, node := range t.nodes {
				if node.selectable {
					selectedIndex = index
					t.currentNode = node
					break
				}
			}
		}
		if selectedIndex < 0 {
			t.currentNode = nil
		}
	}
}

// Draw draws this primitive onto the screen.
func (t *TreeView) Draw(screen tcell.Screen) bool {
	res := t.Box.Draw(screen)
	if !res {
		return false
	}

	if t.root == nil {
		return false
	}

	t.process()

	// Scroll the tree.
	x, y, width, height := t.GetInnerRect()
	switch t.movement {
	case treeUp:
		t.offsetY--
	case treeDown:
		t.offsetY++
	case treeHome:
		t.offsetY = 0
	case treeEnd:
		t.offsetY = len(t.nodes)
	case treePageUp:
		t.offsetY -= height
	case treePageDown:
		t.offsetY += height
	}
	t.movement = treeNone

	// Fix invalid offsets.
	if t.offsetY >= len(t.nodes)-height {
		t.offsetY = len(t.nodes) - height
	}
	if t.offsetY < 0 {
		t.offsetY = 0
	}

	// Draw the tree.
	posY := y
	lineStyle := tcell.StyleDefault.Background(t.backgroundColor).Foreground(t.graphicsColor)
	for index, node := range t.nodes {
		// Skip invisible parts.
		if posY >= y+height+1 {
			break
		}
		if index < t.offsetY {
			continue
		}

		// Draw the graphics.
		if t.graphics {
			// Draw ancestor branches.
			ancestor := node.parent
			for ancestor != nil && ancestor.parent != nil && ancestor.parent.level >= t.topLevel {
				if ancestor.graphicsX >= width {
					continue
				}

				// Draw a branch if this ancestor is not a last child.
				if ancestor.parent.children[len(ancestor.parent.children)-1] != ancestor {
					if posY-1 >= y && ancestor.textX > ancestor.graphicsX {
						PrintJoinedSemigraphics(screen, x+ancestor.graphicsX, posY-1, Borders.Vertical, t.graphicsColor)
					}
					if posY < y+height {
						screen.SetContent(x+ancestor.graphicsX, posY, Borders.Vertical, nil, lineStyle)
					}
				}
				ancestor = ancestor.parent
			}

			if node.textX > node.graphicsX && node.graphicsX < width {
				// Connect to the node above.
				if posY-1 >= y && t.nodes[index-1].graphicsX <= node.graphicsX && t.nodes[index-1].textX > node.graphicsX {
					PrintJoinedSemigraphics(screen, x+node.graphicsX, posY-1, Borders.TopLeft, t.graphicsColor)
				}

				// Join this node.
				if posY < y+height {
					screen.SetContent(x+node.graphicsX, posY, Borders.BottomLeft, nil, lineStyle)
					for pos := node.graphicsX + 1; pos < node.textX && pos < width; pos++ {
						screen.SetContent(x+pos, posY, Borders.Horizontal, nil, lineStyle)
					}
				}
			}
		}

		// Draw the prefix and the text.
		if node.textX < width && posY < y+height {
			// Prefix.
			var bulletCharacterWidth int
			if len(t.bulletCharacters) > 0 {
				_, bulletCharacterWidth = Print(screen, t.bulletCharacters[(node.level-t.topLevel)%len(t.bulletCharacters)], x+node.textX, posY, width-node.textX, AlignLeft, node.color)
			}

			// Text.
			if node.textX+bulletCharacterWidth < width {
				style := tcell.StyleDefault.Foreground(node.color)
				if node == t.currentNode {
					if IsVtxxx {
						style = tcell.StyleDefault.Reverse(true)
					} else {
						style = tcell.StyleDefault.Background(node.color).Foreground(t.backgroundColor)
					}
				}
				if node.blinking {
					style = style.Blink(true)
				}
				if node.underline {
					style = style.Underline(true)
				}
				var fullPrefix string
				for _, prefix := range node.prefixes {
					fullPrefix += prefix
				}
				printWithStyle(screen, fullPrefix+node.text, x+node.textX+bulletCharacterWidth, posY, width-node.textX-bulletCharacterWidth, AlignLeft, style)
			}
		}

		// Advance.
		posY++
	}

	t.drawOverflow(screen, t.offsetY != 0, (t.offsetY != len(t.nodes)-t.innerHeight) && len(t.nodes) > t.innerHeight)

	return true
}

// InputHandler returns the handler for this primitive.
func (t *TreeView) InputHandler() InputHandlerFunc {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) *tcell.EventKey {
		selectNode := func() {
			if t.currentNode != nil {
				if t.selected != nil {
					t.selected(t.currentNode)
				}
				if t.currentNode.selected != nil {
					t.currentNode.selected()
				}
			}
		}

		// Because the tree is flattened into a list only at drawing time, we also
		// postpone the (selection) movement to drawing time.
		switch key := event.Key(); key {
		case tcell.KeyTab, tcell.KeyDown, tcell.KeyRight:
			t.movement = treeDown
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			t.movement = treeUp
		case tcell.KeyHome:
			t.movement = treeHome
		case tcell.KeyEnd:
			t.movement = treeEnd
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			t.movement = treePageDown
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			t.movement = treePageUp
		case tcell.KeyRune:
			if t.vimBindings {
				switch event.Rune() {
				case 'g':
					t.movement = treeHome
				case 'G':
					t.movement = treeEnd
				case 'j':
					t.movement = treeDown
				case 'k':
					t.movement = treeUp
				default:
					return event
				}
			} else if t.searchOnType {
				if time.Since(t.jumpTime) > (500 * time.Millisecond) {
					t.jumpBuffer = ""
				}

				if event.Key() == tcell.KeyRune {
					t.jumpTime = time.Now()
					t.jumpBuffer += strings.ToLower(string(event.Rune()))

					node := t.FindFirstSelectableNode(t.GetRoot(), t.jumpBuffer)
					if node != nil {
						t.SetCurrentNode(node)
					}
				}
			}
		case tcell.KeyEnter:
			selectNode()
		default:
			return event
		}

		t.process()
		return nil
	})
}

// FindFirstSelectableNode iterates through the tree from top to bottom, trying
// to find a node that is selectable and has the given text as its prefix. The
// search is case-insensitive.
func (t *TreeView) FindFirstSelectableNode(node *TreeNode, text string) *TreeNode {
	for _, child := range node.GetChildren() {
		if len(child.GetChildren()) == 0 {
			if child.IsSelectable() && strings.HasPrefix(strings.ToLower(child.GetText()), text) {
				return child
			}
		} else {
			subResult := t.FindFirstSelectableNode(child, text)
			if subResult != nil {
				return subResult
			}
		}
	}

	return nil
}
