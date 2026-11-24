package tree

import "github.com/gdamore/tcell/v3"

// leafNodeAt returns the node of tr that holds a leaf and covers the cell
// (x, y), or nil.
func (tr *Tree) leafNodeAt(x, y int) *Tree {
	if tr == nil {
		return nil
	}
	r := tr.Rect
	if x < r.X || y < r.Y || x >= r.X+r.Width || y >= r.Y+r.Height {
		return nil
	}
	for _, child := range []*Tree{tr.left, tr.right, tr.top, tr.bottom} {
		if n := child.leafNodeAt(x, y); n != nil {
			return n
		}
	}
	if tr.leaf != nil {
		return tr
	}
	return nil
}

// DispatchMouse sends a mouse event to the leaf under the pointer of the
// displayed RootTree. A press of a button also makes that leaf the active
// one. Wheel events and releases go to the leaf under the pointer without
// changing the active leaf.
func DispatchMouse(ev tcell.EventMouse) {
	rt := currentRoot.Load()
	if rt == nil || rt.Root == nil {
		return
	}
	x, y := ev.Position()
	n := rt.Root.leafNodeAt(x, y)
	if n == nil {
		return
	}
	pressed := ev.Buttons() & (tcell.Button1 | tcell.Button2 | tcell.Button3)
	if pressed != 0 && rt.Active != n {
		SetActiveTree(n)
	}
	n.leaf.DispatchMouse(ev)
}
