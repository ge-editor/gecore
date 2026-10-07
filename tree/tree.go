package tree

import (
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gelog"
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

func SetActiveTree(tr *Tree) {
	if activeTree != nil {
		leaf := activeTree.GetLeaf()
		if leaf != nil {
			activeTree.GetLeaf().Active(false)
		}
	}
	activeTree = tr
	activeTree.GetLeaf().Active(true)
}

func GetActiveTree() *Tree {
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
	// maximize 前の split を保存する (splitSaved が true の間のみ有効)
	savedSplit float32
	splitSaved bool

	screen.Rect

	leaf Leaf
}

func (tr *Tree) Parent() *Tree {
	return tr.parent
}

func (tr *Tree) SetLeaf(leaf Leaf) {
	tr.leaf = leaf
}

func (tr *Tree) GetLeaf() Leaf {
	return tr.leaf
}

/*
	 func isActive(t *Tree) Stat {
		if t == GetActiveTree() {
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
	gelog.Debug("tr", tr)
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

	SetActiveTree(tr.top)
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

	SetActiveTree(tr.left)
	tr.Resize(tr.Rect)
}

func (tr *Tree) InsertTop(target *Tree) (*Tree, *Tree) {
	// target := tr.parent
	// target := GetRootTree()
	if target == nil {
		return nil, nil
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

	SetActiveTree(target.top)
	target.Resize(target.Rect)

	return target.top, target.bottom
}

func (tr *Tree) InsertBottom(target *Tree) (*Tree, *Tree) {
	// target := tr.parent
	// target := GetRootTree()
	if target == nil {
		return nil, nil
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

	SetActiveTree(target.bottom)
	target.Resize(target.Rect)

	return target.top, target.bottom
}

func (tr *Tree) InsertRight(target *Tree) (*Tree, *Tree) {
	// target := tr.parent
	// target := GetRootTree()
	if target == nil {
		return nil, nil
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

	SetActiveTree(target.right)
	target.Resize(target.Rect)

	return target.left, target.right
}

func (tr *Tree) InsertLeft(target *Tree) (*Tree, *Tree) {
	// target := tr.parent
	// target := GetRootTree()
	if target == nil {
		return nil, nil
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

	SetActiveTree(target.left)
	target.Resize(target.Rect)

	return target.left, target.right
}

// 最大化時に、押しつぶされる側へ最低限残すセル数
const minOtherCells = 1

// maximize-active-window
// active leaf から root までの経路上の分割比を ratio (例: 0.99) に寄せ、
// それ以外の leaf を最小化する。元の split は初回のみ保存する。
func (tr *Tree) MaximizeActive(ratio float32) {
	active := GetActiveTree()
	if active == nil || active.parent == nil { // 単一 view なら何もしない
		return
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	child := active
	for p := child.parent; p != nil; child, p = p, p.parent {
		// 連続して呼ばれても、最初の(本来の) split を保持する
		if !p.splitSaved {
			p.savedSplit = p.split
			p.splitSaved = true
		}

		// 反対側に最低 minOtherCells を残す (0 サイズの leaf を避ける)
		size := p.Rect.Height
		if p.left != nil {
			size = p.Rect.Width
		}
		r := ratio
		if size > 0 {
			if limit := 1 - float32(minOtherCells)/float32(size); r > limit {
				r = limit
			}
		}

		// split は left/top の割合なので、active が right/bottom 側なら反転する
		if child == p.right || child == p.bottom {
			r = 1 - r
		}
		p.split = r
	}

	GetRootTree().Resize(screen.Get().RootRect())
}

// restore-window-sizes
// MaximizeActive で保存した split を全ノードに復元する。
func (tr *Tree) RestoreSplits() {
	root := GetRootTree()
	if root == nil {
		return
	}
	root.restoreSplits()
	root.Resize(screen.Get().RootRect())
}

func (tr *Tree) restoreSplits() {
	if tr.splitSaved {
		tr.split = tr.savedSplit
		tr.splitSaved = false
	}
	if tr.left != nil {
		tr.left.restoreSplits()
		tr.right.restoreSplits()
	} else if tr.top != nil {
		tr.top.restoreSplits()
		tr.bottom.restoreSplits()
	}
}

// 最大化されているか
func (tr *Tree) IsMaximized() bool {
	maximized := false
	var walk func(*Tree)
	walk = func(t *Tree) {
		if t.splitSaved {
			maximized = true
			return
		}
		if t.left != nil {
			walk(t.left)
			walk(t.right)
		} else if t.top != nil {
			walk(t.top)
			walk(t.bottom)
		}
	}
	walk(GetRootTree())
	return maximized
}

// トグル用 (最大化中なら復元、そうでなければ最大化)
func (tr *Tree) ToggleMaximizeActive(ratio float32) {
	if tr.IsMaximized() {
		tr.RestoreSplits()
	} else {
		tr.MaximizeActive(ratio)
	}
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
	if GetActiveTree().parent == nil { // Root Tree
		return
	}

	sib2 := GetActiveTree().Sibling()
	var sib Tree = *sib2
	GetActiveTree().parent.left = sib.left
	GetActiveTree().parent.right = sib.right
	GetActiveTree().parent.top = sib.top
	GetActiveTree().parent.bottom = sib.bottom
	// GetActiveTree().parent.split = sib.split // The size is the same as the parent size.
	GetActiveTree().parent.leaf = sib.leaf
	/*
		if GetActiveTree().parent.leaf != nil {
			(*GetActiveTree().parent.leaf).SetParentTree(GetActiveTree().parent)
		}
	*/

	if GetActiveTree().parent.left != nil {
		GetActiveTree().parent.left.parent = sib.parent
		GetActiveTree().parent.right.parent = sib.parent
	} else if GetActiveTree().parent.top != nil {
		GetActiveTree().parent.top.parent = sib.parent
		GetActiveTree().parent.bottom.parent = sib.parent
	}

	SetActiveTree(sib.parent.firstLeafNode())

	GetRootTree().Resize(screen.Get().RootRect())
}

// delete-other-windows
// C-x 1
// Kill all views but active
func (tr *Tree) DeleteOtherWindows() {
	active := GetActiveTree()
	root := GetRootTree()
	if active == nil || root == nil {
		return
	}
	// すでに単一 view なら何もしない
	if root.leaf != nil {
		return
	}

	// 残す leaf を退避する
	keep := active.leaf
	if keep == nil {
		return
	}

	// 他の leaf を破棄する
	root.ForEachLeaf(func(l Leaf) {
		if l != keep {
			closeLeaf(l)
		}
	})

	// root を leaf ノードに戻す
	root.left = nil
	root.right = nil
	root.top = nil
	root.bottom = nil
	root.split = 0
	root.leaf = keep

	SetActiveTree(root)
	root.Resize(screen.Get().RootRect())
}

// 破棄対象の leaf に後始末が必要な場合のみ実装する
func closeLeaf(l Leaf) {
	if c, ok := l.(interface{ Close() }); ok {
		c.Close()
	}
}

func (tr *Tree) NextInCycle() {
	next := GetActiveTree().nextInCycle()
	if next != nil && next.leaf != nil {
		SetActiveTree(next)
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
