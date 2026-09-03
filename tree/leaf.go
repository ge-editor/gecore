package tree

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/keychord"
	"github.com/ge-editor/utils"
)

type Leaf interface {
	LeafType() LeafType

	Resize(utils.Rect)
	Draw() bool
	Kill(Leaf, bool) Leaf
	Active(bool)

	DispatchKey(ev tcell.EventKey) (string, keychord.KeyDispatchTransition)

	DispatchMouse(ev tcell.EventMouse) // (string, keychord.KeyDispatchTransition)

	Resume()
	Init()
	WillClose()
}
