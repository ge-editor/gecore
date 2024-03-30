package tree

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/utils"
)

var Views *ViewsStruct // User Views
var ActiveLeaf *Leaf

type View interface {
	NewLeaf() *Leaf

	// Mainly used for splitting the screen.
	// In the new window created by splitting the screen,
	// a new tree.Leaf of the same type as the parent tree.Leaf before splitting is cloned.
	//
	// For te.Editor:
	// Create a new tree.Leaf (Editor) and make it the same as leaf *tree.Leaf
	// direction: "right", "bottom" are not referenced
	NewSiblingLeaf(direction string, leaf *Leaf) *Leaf

	Name() string // this view name
	// WillClose() // Should I move it here?
}

type Leaf interface {
	View() *View
	// SetupNewSibling(currentTree, siblingTree *Leaf)
	// Tree() *Tree
	// SetParentTree(*Tree)

	Resize(int, int, utils.Rect)
	Draw()
	Redraw()
	Kill(*Leaf, bool) *Leaf
	// ForceDraw()
	// Stat(Stat)
	ViewActive(bool)

	Event(*tcell.Event) *tcell.Event
	// EventInterrupt(*tcell.EventInterrupt) *tcell.EventInterrupt
	// EventResize(*tcell.EventResize) *tcell.EventResize
	// EventKey(*tcell.EventKey) *tcell.EventKey

	Resume()
	Init()
	WillClose() //
}
