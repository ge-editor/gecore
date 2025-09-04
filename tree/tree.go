package tree

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/utils"
)

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
			(*activeTree.GetLeaf()).ViewActive(false)
		}
	}
	activeTree = tr
	(*activeTree.GetLeaf()).ViewActive(true)
}

func ActiveTreeGet() *Tree {
	return activeTree
}

// Create a new Tree and assign view
func NewRootTree(view *View) *Tree {
	return &Tree{
		parent: nil,
		leaf:   (*view).NewLeaf(),
	}
}

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

	utils.Rect

	leaf *Leaf
}

func (tr *Tree) SetLeaf(leaf *Leaf) {
	tr.leaf = leaf
}

func (tr *Tree) GetLeaf() *Leaf {
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

func (tr *Tree) Kill(leaf *Leaf, isActive bool) {
	if tr.left != nil {
		tr.left.Kill(leaf, isActive)
		tr.right.Kill(leaf, isActive)
	} else if tr.top != nil {
		tr.top.Kill(leaf, isActive)
		tr.bottom.Kill(leaf, isActive)
	} else {
		tr.leaf = (*tr.leaf).Kill(leaf, isActive)
	}
}

func (tr *Tree) Draw() {
	if tr.left != nil {
		tr.left.Draw()
		tr.right.Draw()
	} else if tr.top != nil {
		tr.top.Draw()
		tr.bottom.Draw()
	} else {
		(*tr.leaf).Draw()
	}
}

func (tr *Tree) Redraw() {
	if tr.left != nil {
		tr.left.Redraw()
		tr.right.Redraw()
	} else if tr.top != nil {
		tr.top.Redraw()
		tr.bottom.Redraw()
	} else {
		(*tr.leaf).Redraw()
	}
}

func GetLeavesByViewName(name string) []*Leaf {
	leaves := []*Leaf{}
	collectLeaves(&leaves, rootTree, name)
	return leaves
}

func collectLeaves(leaves *[]*Leaf, tr *Tree, name string) {
	if tr.left != nil {
		collectLeaves(leaves, tr.left, name)
		collectLeaves(leaves, tr.right, name)
	} else if tr.top != nil {
		collectLeaves(leaves, tr.top, name)
		collectLeaves(leaves, tr.bottom, name)
	} else {
		if (*(*tr.leaf).View()).Name() == name {
			*leaves = append(*leaves, tr.leaf)
		}
	}
}

func (tr *Tree) Event(ev *tcell.Event) {
	if tr.left != nil {
		tr.left.Event(ev)
		tr.right.Event(ev)
	} else if tr.top != nil {
		tr.top.Event(ev)
		tr.bottom.Event(ev)
	} else {
		(*tr.leaf).Event(ev)
	}
}

// or error
func (tr *Tree) Resize(rect utils.Rect) {
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
		verb.PP("v.Rect, rect %v %v, lw,rw %d,%d", tr.Rect, rect, lw, rw)
		tr.left.Resize(utils.Rect{X: rect.X, Y: rect.Y, Width: lw, Height: rect.Height})
		tr.right.Resize(utils.Rect{X: rect.X + lw /* + 1 */, Y: rect.Y, Width: rw, Height: rect.Height})
	} else if tr.top != nil {
		// vertical split, use 'h', no need to reserve one line for
		// splitter, because splitters are part of the buffer's output
		// (their modelines act like a splitter)
		h := rect.Height
		th := int(float32(h) * tr.split)
		bh := h - th
		tr.top.Resize(utils.Rect{X: rect.X, Y: rect.Y, Width: rect.Width, Height: th})
		tr.bottom.Resize(utils.Rect{X: rect.X, Y: rect.Y + th, Width: rect.Width, Height: bh})
	} else {
		// s := screen.Get()
		(*tr.leaf).Resize( /* s.Width, s.Height, */ rect)
	}
}

// Return the Leaf for the new Tree created by split screen
func (tr *Tree) newLeaf(direction string) *Leaf {
	viewName := (*(*tr.GetLeaf()).View()).Name()        // Get View Name
	v := Views.GetViewByName(viewName)                  // Get View by View Name
	return (*v).NewSiblingLeaf(direction, tr.GetLeaf()) // New Leaf
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
	current := rootTree
	rootTree = &Tree{}
	rootTree.top = &Tree{
		parent: rootTree,
		leaf:   tr.newLeaf("top"),
	}
	if current.leaf == nil {
		rootTree.bottom = current
		current.parent = rootTree
	} else {
		rootTree.bottom = &Tree{
			parent: rootTree,
			leaf:   current.leaf,
		}
	}
	rootTree.Rect = current.Rect
	rootTree.split = 0.5

	ActiveTreeSet(rootTree.top)
	rootTree.Resize(rootTree.Rect)
}

func (tr *Tree) InsertRight() {
	current := rootTree
	rootTree = &Tree{}
	rootTree.right = &Tree{
		parent: rootTree,
		leaf:   tr.newLeaf("right"),
	}
	if current.leaf == nil {
		rootTree.left = current
		current.parent = rootTree
	} else {
		rootTree.left = &Tree{
			parent: rootTree,
			leaf:   current.leaf,
		}
	}
	rootTree.Rect = current.Rect
	rootTree.split = 0.5

	ActiveTreeSet(rootTree.right)
	rootTree.Resize(rootTree.Rect)
}

func (tr *Tree) InsertBottom() {
	current := rootTree
	rootTree = &Tree{}
	rootTree.bottom = &Tree{
		parent: rootTree,
		leaf:   tr.newLeaf("bottom"),
	}
	if current.leaf == nil {
		rootTree.top = current
		current.parent = rootTree
	} else {
		rootTree.top = &Tree{
			parent: rootTree,
			leaf:   current.leaf,
		}
	}
	rootTree.Rect = current.Rect
	rootTree.split = 0.5

	ActiveTreeSet(rootTree.bottom)
	rootTree.Resize(rootTree.Rect)
}

func (tr *Tree) InsertLeft() {
	current := rootTree
	rootTree = &Tree{}
	rootTree.left = &Tree{
		parent: rootTree,
		leaf:   tr.newLeaf("left"),
	}
	if current.leaf == nil {
		rootTree.right = current
		current.parent = rootTree
	} else {
		rootTree.right = &Tree{
			parent: rootTree,
			leaf:   current.leaf,
		}
	}
	rootTree.Rect = current.Rect
	rootTree.split = 0.5

	ActiveTreeSet(rootTree.left)
	rootTree.Resize(rootTree.Rect)
}

func (tr *Tree) SwitchSplitDirection() {
	if tr.parent.top != nil {
		tr.parent.left = tr.parent.top
		tr.parent.right = tr.parent.bottom
		tr.parent.top = nil
		tr.parent.bottom = nil
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
