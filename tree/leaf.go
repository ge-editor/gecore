package tree

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/utils"
)

type Leaf interface {
	View() *View
	// SetupNewSibling(currentTree, siblingTree *Leaf)
	// Tree() *Tree
	// SetParentTree(*Tree)

	Resize(viewArea utils.Rect)
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

	MiniBufferMode(int)
}
