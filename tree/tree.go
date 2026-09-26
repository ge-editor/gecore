package tree

import (
	"github.com/ge-editor/gecore/screen"
)

// Context Root に名を変更 2026-01-24 Sat

// Want to be able to perform dynamic screen switching by being able to replace the RootTree
var rootTree *Tree

func GetRootTree() *Tree {
	return rootTree
}

func SetRootTree(tr *Tree) {
	rootTree = tr
}

var activeTree *Tree

func ActiveTreeSet(tr *Tree) {
	if activeTree != nil {
		leaf := activeTree.GetLeaf()
		if leaf != nil {
			activeTree.GetLeaf().Active(false)
		}
	}
	activeTree = tr
	activeTree.GetLeaf().Active(true)
}

func ActiveTreeGet() *Tree {
	return activeTree
}

// Create a new Tree and assign view
func NewRootTree(lt LeafType) *Tree {
	return &Tree{
		parent: nil,
		leaf:   lt.NewLeaf(),
	}
}

// Required implement Overlay interface
type Tree struct {
	// At the same time only one of these groups can be valid:
	// 1) 'left', 'right' and 'split'
	// 2) 'top', 'bottom' and 'split'
	// 3) 'leaf'
	parent *Tree

	left  *Tree
	right *Tree

	top    *Tree
	bottom *Tree

	split float32

	screen.Rect

	leaf Leaf
}

func (tr *Tree) SetLeaf(leaf Leaf) {
	tr.leaf = leaf
}

func (tr *Tree) GetLeaf() Leaf {
	return tr.leaf
}

/*
	 func isActive(t *Tree) Stat {
		if t == ActiveTreeGet() {
			return Active
		}
		return Inactive
	}
*/

func (tr *Tree) Draw() bool {
	if tr.left != nil {
		if tr.left.Draw() {
			return true
		}
		return tr.right.Draw()
	} else if tr.top != nil {
		if tr.top.Draw() {
			return true
		}
		return tr.bottom.Draw()
	} else {
		return tr.leaf.Draw()
	}
}

func (tr *Tree) UniversalCancel() {
}

func (tr *Tree) ForEachLeaf(fn func(Leaf)) {
	if tr.left != nil {
		tr.left.ForEachLeaf(fn)
		tr.right.ForEachLeaf(fn)
	} else if tr.top != nil {
		tr.top.ForEachLeaf(fn)
		tr.bottom.ForEachLeaf(fn)
	} else {
		fn(tr.leaf)
	}
}

/*
// A
func GetLeafTypeByRegisterName(name string) []Leaf {
	leaves := []Leaf{}
	collectLeaves(leaves, rootTree, name)
	return leaves
}

func collectLeaves(leaves []Leaf, tr *Tree, name string) {
	if tr.left != nil {
		collectLeaves(leaves, tr.left, name)
		collectLeaves(leaves, tr.right, name)
	} else if tr.top != nil {
		collectLeaves(leaves, tr.top, name)
		collectLeaves(leaves, tr.bottom, name)
	} else {
		if tr.leaf.LeafType().Name() == name {
			leaves = append(leaves, tr.leaf)
		}
	}
}
*/

/* func (tr *Tree) Type() overlay.OverlayType {
	return overlay.OverlayFlow
}
*/

// tree の高さは、Overlay の残り全てを使用するので意味を持たない
func (tr *Tree) RequiredHeight() int {
	return -1
}

// or error
func (tr *Tree) Resize(rect screen.Rect) {
	tr.Rect = rect

	if tr.left != nil {
		// horizontal split, use 'w'
		w := rect.Width
		/*
			if w > 0 {
				// reserve one line for splitter, if we have one line
				w--
			}
		*/
		lw := int(float32(w) * tr.split)
		rw := w - lw
		// gelog.Info("v.Rect, rect %v %v, lw,rw %d,%d", tr.Rect, rect, lw, rw)
		tr.left.Resize(screen.Rect{X: rect.X, Y: rect.Y, Width: lw, Height: rect.Height})
		tr.right.Resize(screen.Rect{X: rect.X + lw /* + 1 */, Y: rect.Y, Width: rw, Height: rect.Height})
	} else if tr.top != nil {
		// vertical split, use 'h', no need to reserve one line for
		// splitter, because splitters are part of the buffer's output
		// (their modelines act like a splitter)
		h := rect.Height
		th := int(float32(h) * tr.split)
		bh := h - th
		tr.top.Resize(screen.Rect{X: rect.X, Y: rect.Y, Width: rect.Width, Height: th})
		tr.bottom.Resize(screen.Rect{X: rect.X, Y: rect.Y + th, Width: rect.Width, Height: bh})
	} else {
		// s := screen.Get()
		tr.leaf.Resize( /* s.Width, s.Height, */ rect)
	}
}

// Return the Leaf for the new Tree created by split screen
func (tr *Tree) newLeaf(direction string) Leaf {
	viewName := tr.GetLeaf().LeafType().Name()              // Get View Name
	v, ok := LeafTypes.GetLeafTypeByReginsterName(viewName) // Get View by View Name
	if !ok {

	}
	// gelog.Debug("tr", tr, "tr.GetLeaf()", tr.GetLeaf())
	return v.NewSiblingLeaf(direction, tr.GetLeaf()) // New Leaf
}

// Split Tree. Set new Active Tree to active.
func (tr *Tree) SplitVertically() {
	tr.top = &Tree{
		parent: tr,
		leaf:   tr.leaf,
	}
	tr.bottom = &Tree{
		parent: tr,
		leaf:   tr.newLeaf("bottom"),
	}
	tr.split = 0.5
	tr.leaf = nil

	ActiveTreeSet(tr.top)
	tr.Resize(tr.Rect)
}

// Split Tree. Set new Active Tree to active.
func (tr *Tree) SplitHorizontally() {
	tr.left = &Tree{
		parent: tr,
		leaf:   tr.leaf,
	}
	tr.right = &Tree{
		parent: tr,
		leaf:   tr.newLeaf("right"),
	}
	tr.split = 0.5
	tr.leaf = nil

	ActiveTreeSet(tr.left)
	tr.Resize(tr.Rect)
}

func (tr *Tree) InsertTop() {
	target := tr.parent
	if target == nil {
		return
	}

	// target の現在の中身を新しい bottom に退避する。
	bottom := &Tree{
		parent: target,
		left:   target.left,
		right:  target.right,
		top:    target.top,
		bottom: target.bottom,
		split:  target.split,
		leaf:   target.leaf,
		Rect:   target.Rect,
	}

	if bottom.left != nil {
		bottom.left.parent = bottom
		bottom.right.parent = bottom
	} else if bottom.top != nil {
		bottom.top.parent = bottom
		bottom.bottom.parent = bottom
	}

	// target 自身を新しい vertical split node に変身させる。
	target.left = nil
	target.right = nil
	target.top = &Tree{
		parent: target,
		leaf:   tr.newLeaf("top"),
	}
	target.bottom = bottom
	target.leaf = nil
	target.split = 0.5

	ActiveTreeSet(target.top)
	target.Resize(target.Rect)
}

func (tr *Tree) InsertBottom() {
	target := tr.parent
	if target == nil {
		return
	}

	// target の現在の中身を新しい top に退避する。
	top := &Tree{
		parent: target,
		left:   target.left,
		right:  target.right,
		top:    target.top,
		bottom: target.bottom,
		split:  target.split,
		leaf:   target.leaf,
		Rect:   target.Rect,
	}

	if top.left != nil {
		top.left.parent = top
		top.right.parent = top
	} else if top.top != nil {
		top.top.parent = top
		top.bottom.parent = top
	}

	// target 自身を新しい vertical split node に変身させる。
	target.left = nil
	target.right = nil
	target.top = top
	target.bottom = &Tree{
		parent: target,
		leaf:   tr.newLeaf("bottom"),
	}
	target.leaf = nil
	target.split = 0.5

	ActiveTreeSet(target.bottom)
	target.Resize(target.Rect)
}

func (tr *Tree) InsertRight() {
	target := tr.parent
	if target == nil {
		return
	}

	// target の現在の中身を新しい left に退避する。
	left := &Tree{
		parent: target,
		left:   target.left,
		right:  target.right,
		top:    target.top,
		bottom: target.bottom,
		split:  target.split,
		leaf:   target.leaf,
		Rect:   target.Rect,
	}

	if left.left != nil {
		left.left.parent = left
		left.right.parent = left
	} else if left.top != nil {
		left.top.parent = left
		left.bottom.parent = left
	}

	// target 自身を新しい horizontal split node に変身させる。
	target.left = left
	target.right = &Tree{
		parent: target,
		leaf:   tr.newLeaf("right"),
	}
	target.top = nil
	target.bottom = nil
	target.leaf = nil
	target.split = 0.5

	ActiveTreeSet(target.right)
	target.Resize(target.Rect)
}

func (tr *Tree) InsertLeft() {
	target := tr.parent
	if target == nil {
		return
	}

	// target の現在の中身を新しい right に退避する。
	right := &Tree{
		parent: target,
		left:   target.left,
		right:  target.right,
		top:    target.top,
		bottom: target.bottom,
		split:  target.split,
		leaf:   target.leaf,
		Rect:   target.Rect,
	}

	if right.left != nil {
		right.left.parent = right
		right.right.parent = right
	} else if right.top != nil {
		right.top.parent = right
		right.bottom.parent = right
	}

	// target 自身を新しい horizontal split node に変身させる。
	target.left = &Tree{
		parent: target,
		leaf:   tr.newLeaf("left"),
	}
	target.right = right
	target.top = nil
	target.bottom = nil
	target.leaf = nil
	target.split = 0.5

	ActiveTreeSet(target.left)
	target.Resize(target.Rect)
}

var tree1, tree2 *Tree

func (tr *Tree) SwitchSplitDirection() {
	if tr.parent.top != nil {
		if tr.parent.top == tree1 && tr.parent.bottom == tree2 {
			tr.parent.left = tr.parent.bottom
			tr.parent.right = tr.parent.top
		} else {
			tr.parent.left = tr.parent.top
			tr.parent.right = tr.parent.bottom
		}
		tr.parent.top = nil
		tr.parent.bottom = nil

		tree1 = tr.parent.left
		tree2 = tr.parent.right
	} else if tr.parent.left != nil {
		tr.parent.top = tr.parent.left
		tr.parent.bottom = tr.parent.right
		tr.parent.left = nil
		tr.parent.right = nil
	}

	tr.parent.Resize(tr.parent.Rect)
}

// This method recursively traverses the tree and applies a callback function cb to each node it visits.
func (tr *Tree) Traverse(cb func(*Tree)) {
	if tr.left != nil {
		tr.left.Traverse(cb)
		tr.right.Traverse(cb)
	} else if tr.top != nil {
		tr.top.Traverse(cb)
		tr.bottom.Traverse(cb)
	} else {
		cb(tr)
	}
}

func (tr *Tree) NearestVSplit() *Tree {
	tr = tr.parent
	for tr != nil {
		if tr.top != nil {
			return tr
		}
		tr = tr.parent
	}
	return nil
}

func (tr *Tree) NearestHSplit() *Tree {
	tr = tr.parent
	for tr != nil {
		if tr.left != nil {
			return tr
		}
		tr = tr.parent
	}
	return nil
}

func (tr *Tree) oneStep() float32 {
	if tr.top != nil {
		return 1.0 / float32(tr.Height)
	} else if tr.left != nil {
		return 1.0 / float32(tr.Width-1)
	}
	return 0.0
}

func (tr *Tree) normalizeSplit() {
	var off int
	if tr.top != nil {
		off = int(float32(tr.Height) * tr.split)
	} else {
		off = int(float32(tr.Width-1) * tr.split)
	}
	tr.split = float32(off) * tr.oneStep()
}

func (tr *Tree) StepResize(n int) {
	if tr.Width <= 1 || tr.Height <= 0 {
		// avoid division by zero, result is really bad
		return
	}

	one := tr.oneStep()
	tr.normalizeSplit()
	tr.split += one*float32(n) + (one * 0.5)
	if tr.split > 1.0 {
		tr.split = 1.0
	}
	if tr.split < 0.0 {
		tr.split = 0.0
	}
	tr.Resize(tr.Rect)
}

// Find subling Tree
func (tr *Tree) Sibling() *Tree {
	if tr.parent == nil {
		return nil
	}
	switch {
	case tr == tr.parent.left:
		return tr.parent.right
	case tr == tr.parent.right:
		return tr.parent.left
	case tr == tr.parent.top:
		return tr.parent.bottom
	case tr == tr.parent.bottom:
		return tr.parent.top
	}
	panic("sibling unreachable")
}

func (tr *Tree) firstLeafNode() *Tree {
	if tr.left != nil {
		return tr.left.firstLeafNode()
	} else if tr.top != nil {
		return tr.top.firstLeafNode()
	} else if tr.leaf != nil {
		return tr
	}
	panic("unreachable")
}

// delete-window
// C-x 0
func (tr *Tree) DeleteWindow() {
	if ActiveTreeGet().parent == nil { // Root Tree
		return
	}

	sib2 := ActiveTreeGet().Sibling()
	var sib Tree = *sib2
	ActiveTreeGet().parent.left = sib.left
	ActiveTreeGet().parent.right = sib.right
	ActiveTreeGet().parent.top = sib.top
	ActiveTreeGet().parent.bottom = sib.bottom
	// ActiveTreeGet().parent.split = sib.split // The size is the same as the parent size.
	ActiveTreeGet().parent.leaf = sib.leaf
	/*
		if ActiveTreeGet().parent.leaf != nil {
			(*ActiveTreeGet().parent.leaf).SetParentTree(ActiveTreeGet().parent)
		}
	*/

	if ActiveTreeGet().parent.left != nil {
		ActiveTreeGet().parent.left.parent = sib.parent
		ActiveTreeGet().parent.right.parent = sib.parent
	} else if ActiveTreeGet().parent.top != nil {
		ActiveTreeGet().parent.top.parent = sib.parent
		ActiveTreeGet().parent.bottom.parent = sib.parent
	}

	ActiveTreeSet(sib.parent.firstLeafNode())

	GetRootTree().Resize(screen.Get().RootRect())
}

func (tr *Tree) NextInCycle() {
	next := ActiveTreeGet().nextInCycle()
	if next != nil && next.leaf != nil {
		ActiveTreeSet(next)
	}
}

// nextInCycle returns the next view
// after v, eventually cycling
// through all views.
func (tr *Tree) nextInCycle() (ret *Tree) {

	root := tr
	for root.parent != nil {
		root = root.parent
	}

	// find our number
	var k, our int
	root.Traverse(func(w *Tree) {
		if w == tr {
			our = k
		}
		k++
	})
	// return the view after us
	tot := k
	k = 0
	root.Traverse(func(w *Tree) {
		if k == (our+1)%tot {
			ret = w
		}
		k++
	})
	return
}

// Implement overlay
func (tr *Tree) IsActive() bool {
	return true
}
