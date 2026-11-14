package tree

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gelog"
)

// Context Root に名を変更 2026-01-24 Sat

// RootTree の所有者は roottrees.go の RootTrees。
// 以下は「現在表示中の RootTree」への互換アクセサ。
func GetRootTree() *Tree {
	rt := currentRoot.Load()
	if rt == nil {
		return nil
	}
	return rt.Root
}

// SetRootTree sets the Root of the current RootTreeStruct.
// If none exists yet, a "default" RootTreeStruct is created, registered
// in RootTrees and made current.
func SetRootTree(tr *Tree) {
	rt := currentRoot.Load()
	if rt == nil {
		rt = &RootTreeStruct{Name: "default"}
		AddRootTree(rt)
		currentRoot.Store(rt)
	}
	rt.Root = tr
}

// SetActiveTree sets the active Tree of the current RootTreeStruct.
//
// Exactly one leaf of the RootTree ends up with Active(true). Every other
// leaf is deactivated by walking the tree instead of trusting rt.Active:
// the split operations (SplitHorizontally, InsertRight, ...) turn the
// previously active node into a split node (its leaf is moved into a child
// and tr.leaf becomes nil) before calling this function, so rt.Active.GetLeaf()
// is already nil there and the old leaf would otherwise stay active. Two
// active leaves make the cursor (drawn by whichever active leaf is drawn
// last) disagree with the window that receives keys.
func SetActiveTree(tr *Tree) {
	rt := currentRoot.Load()
	if rt == nil || tr == nil {
		return
	}
	next := tr.GetLeaf()

	if rt.Active != nil && rt.Active != tr {
		if l := rt.Active.GetLeaf(); l != nil && l != next {
			l.Active(false)
		}
	}
	if rt.Root != nil {
		rt.Root.ForEachLeaf(func(l Leaf) bool {
			if l != nil && l != next {
				l.Active(false)
			}
			return true
		})
	}

	rt.Active = tr
	if next != nil {
		next.Active(true)
	}
}

func GetActiveTree() *Tree {
	rt := currentRoot.Load()
	if rt == nil {
		return nil
	}
	return rt.Active
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
	// restoreActive is populated from a saved LayoutNode.Active marker.
	restoreActive bool

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

// SelectLeaf shows the leaf of the registered type name in this pane.
//
// The leaf that was shown is parked in the pane, not destroyed: it is not
// drawn, but it keeps its state and is still visited by ForEachLeaf (so
// buffers shared with it stay in sync, the quit guard sees its unsaved
// changes, and closing the pane closes it). Selecting its type again brings
// the same instance back. A type that has no parked leaf in this pane is
// created with its LeafType's NewLeaf.
func (tr *Tree) SelectLeaf(name string) error {
	if tr == nil || tr.leaf == nil {
		return errors.New("tree: not a leaf pane")
	}
	cur := tr.leaf
	if cur.LeafType() != nil && cur.LeafType().Name() == name {
		return nil
	}

	var next Leaf
	if LeafTypes == nil {
		return errors.New("tree: leaf types are not initialized")
	}
	lt, ok := LeafTypes.GetLeafTypeByReginsterName(name)
	if !ok {
		return fmt.Errorf("tree: unknown leaf type %q", name)
	}
	next = lt.NewLeaf()

	isActive := tr == GetActiveTree()

	cur.Active(false)
	tr.leaf = next
	if isActive {
		next.Active(true)
	}

	tr.Resize(tr.Rect)
	next.Resume()
	return nil
}

func (tr *Tree) Draw(ctx context.Context) bool {
	if tr.left != nil {
		if tr.left.Draw(ctx) {
			return true
		}
		return tr.right.Draw(ctx)
	} else if tr.top != nil {
		if tr.top.Draw(ctx) {
			return true
		}
		return tr.bottom.Draw(ctx)
	} else {
		return tr.leaf.Draw(ctx)
	}
}

func (tr *Tree) UniversalCancel() {
}

func (tr *Tree) ForEachLeaf(fn func(Leaf) bool) bool {
	if tr.left != nil {
		if !tr.left.ForEachLeaf(fn) {
			return false
		}
		return tr.right.ForEachLeaf(fn)
	} else if tr.top != nil {
		if !tr.top.ForEachLeaf(fn) {
			return false
		}
		return tr.bottom.ForEachLeaf(fn)
	}

	return fn(tr.leaf)
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
	leafName := tr.GetLeaf().LeafType().Name()              // Get View Name
	v, ok := LeafTypes.GetLeafTypeByReginsterName(leafName) // Get View by View Name
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

	// The leaf being removed must release what it registered elsewhere
	// (e.g. Editorleaf unsubscribes from its buffer's syntax Worker).
	if l := GetActiveTree().GetLeaf(); l != nil {
		l.WillClose()
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

// replaceWith makes tr take over the contents of sib (its child).
// The children of sib are re-parented to tr. Without this their parent
// would keep pointing at the discarded sib node, and later operations that
// walk or edit through parent (DeleteWindow, Sibling, SwitchSplitDirection,
// nextInCycle) would work on a node that is no longer in the tree.
func (tr *Tree) replaceWith(sib *Tree) {
	tr.left = sib.left
	tr.right = sib.right
	tr.top = sib.top
	tr.bottom = sib.bottom
	tr.split = sib.split
	tr.savedSplit = sib.savedSplit
	tr.splitSaved = sib.splitSaved
	tr.leaf = sib.leaf

	if tr.left != nil {
		tr.left.parent = tr
		tr.right.parent = tr
	} else if tr.top != nil {
		tr.top.parent = tr
		tr.bottom.parent = tr
	}
}

// removeLeaf は target を木から取り除き、sibling を親に昇格させる。
// root ポインタは維持される。
func (tr *Tree) removeLeaf(target Leaf) {
	if tr == nil || tr.leaf != nil {
		return
	}

	if tr.left != nil { // 左右分割
		if tr.left.leaf == target {
			tr.replaceWith(tr.right) // 左を削除 → 右を昇格
			return
		}
		if tr.right.leaf == target {
			tr.replaceWith(tr.left) // 右を削除 → 左を昇格
			return
		}
		tr.left.removeLeaf(target)
		tr.right.removeLeaf(target)
		return
	}

	if tr.top != nil { // 上下分割
		if tr.top.leaf == target {
			tr.replaceWith(tr.bottom) // 上を削除 → 下を昇格
			return
		}
		if tr.bottom.leaf == target {
			tr.replaceWith(tr.top) // 下を削除 → 上を昇格
			return
		}
		tr.top.removeLeaf(target)
		tr.bottom.removeLeaf(target)
	}
}

// findTreeForLeaf は指定 leaf を持っている Tree ノードを返す。
func (tr *Tree) findTreeForLeaf(target Leaf) *Tree {
	if tr == nil {
		return nil
	}
	if tr.leaf == target {
		return tr
	}
	if tr.left != nil {
		if t := tr.left.findTreeForLeaf(target); t != nil {
			return t
		}
		return tr.right.findTreeForLeaf(target)
	}
	if tr.top != nil {
		if t := tr.top.findTreeForLeaf(target); t != nil {
			return t
		}
		return tr.bottom.findTreeForLeaf(target)
	}
	return nil
}

// DeleteOtherWindows is delete-other-windows (C-x 1): it closes the leaves
// other than the active one.
//
// targets and excludes are leaf type names (LeafType().Name()). When targets
// is not empty, only leaves of those types are closed; leaves of the types in
// excludes are never closed. The leaves that are not closed stay in the
// layout, so with targets or excludes more than one leaf can remain.
func (tr *Tree) DeleteOtherWindows(targets, excludes []string) {
	if !deleteOtherWindows(targets, excludes) {
		return
	}
	GetRootTree().Resize(screen.Get().RootRect())
}

// deleteOtherWindows does the work of DeleteOtherWindows without the final
// re-layout (which needs the terminal screen). It reports whether the tree
// was changed.
func deleteOtherWindows(targets, excludes []string) bool {
	active := GetActiveTree()
	root := GetRootTree()
	if active == nil || root == nil {
		return false
	}
	// すでに単一なら何もしない
	if root.leaf != nil {
		return false
	}

	keep := active.leaf
	if keep == nil {
		return false
	}

	contains := func(list []string, target string) bool {
		for _, v := range list {
			if v == target {
				return true
			}
		}
		return false
	}

	// 1. 閉じる対象を先に収集（走査中に木を変えない）
	var toClose []Leaf
	root.ForEachLeaf(func(l Leaf) bool {
		if l == keep {
			return true
		}
		name := ""
		if lt := l.LeafType(); lt != nil {
			name = lt.Name()
		}

		// targets が指定されている場合、含まれていなければ閉じない
		if len(targets) > 0 && !contains(targets, name) {
			return true
		}
		// excludes に含まれていれば閉じない
		if len(excludes) > 0 && contains(excludes, name) {
			return true
		}

		toClose = append(toClose, l)
		return true
	})

	// 2. 収集した leaf を順に削除（リソース解放 + 木の再構成）
	for _, l := range toClose {
		cleanupAfterCloseLeaf(l)
		root.removeLeaf(l)
	}

	// 3. keep が乗っているノードを探して active にする
	//    （昇格で元の active ノードが消えている可能性があるため）
	if at := root.findTreeForLeaf(keep); at != nil {
		SetActiveTree(at)
	} else {
		// 念のため
		SetActiveTree(root)
	}

	return true
}

// 失敗
// delete-other-windows
// C-x 1
// Kill all views but active
// root.leaf.LeafType().Name() == targets
// root.leaf.LeafType().Name() == excludes
/* func (tr *Tree) DeleteOtherWindows__(targets, excludes []string) {
	active := GetActiveTree()
	root := GetRootTree()
	if active == nil || root == nil {
		return
	}
	// Do nothing if it is already a single view
	if root.leaf != nil {
		return
	}

	// Save the leaf to keep
	keep := active.leaf
	if keep == nil {
		return
	}

	// Helper function to check if the slice contains the target string
	contains := func(list []string, target string) bool {
		for _, v := range list {
			if v == target {
				return true
			}
		}
		return false
	}

	// 再帰的に不要なリーフを削除し、空になった枝を畳む（縮約する）ヘルパー関数
	var prune func(t *Tree) bool
	prune = func(t *Tree) bool {
		if t == nil {
			return false
		}

		// リーフの場合
		if t.leaf != nil {
			if t.leaf == keep {
				return false // 残す
			}

			name := t.leaf.LeafType().Name()

			// Target check: 指定されたターゲットに含まれない場合は残す（閉じない）
			if len(targets) > 0 && !contains(targets, name) {
				return false
			}

			// Exclude check: 除外リストに含まれている場合は残す（閉じない）
			if len(excludes) > 0 && contains(excludes, name) {
				return false
			}

			// 削除対象のウィンドウを閉じる
			cleanupAfterCloseLeaf(t.leaf)
			return true // このリーフが削除され、空になったことを通知
		}

		// 左右分割の場合
		if t.left != nil && t.right != nil {
			leftDeleted := prune(t.left)
			rightDeleted := prune(t.right)

			if leftDeleted && rightDeleted {
				t.left = nil
				t.right = nil
				return true
			} else if leftDeleted {
				*t = *t.right // 左が消えたので右を昇格
				return false
			} else if rightDeleted {
				*t = *t.left // 右が消えたので左を昇格
				return false
			}
			return false
		}

		// 上下分割の場合
		if t.top != nil && t.bottom != nil {
			topDeleted := prune(t.top)
			bottomDeleted := prune(t.bottom)

			if topDeleted && bottomDeleted {
				t.top = nil
				t.bottom = nil
				return true
			} else if topDeleted {
				*t = *t.bottom // 上が消えたので下を昇格
				return false
			} else if bottomDeleted {
				*t = *t.top // 下が消えたので上を昇格
				return false
			}
			return false
		}

		return false
	}

	// ツリー全体の整理を実行
	prune(root)

	SetActiveTree(active)
	root.Resize(screen.Get().RootRect())
} */

// delete-other-windows
// C-x 1
// Kill all views but active
/* func (tr *Tree) DeleteOtherWindows_original() {
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
	root.ForEachLeaf(func(l Leaf) bool {
		if l != keep {
			cleanupAfterCloseLeaf(l)
		}
		return true
	})

	// keep 以外の leaf が消える
	root.left = nil
	root.right = nil
	root.top = nil
	root.bottom = nil
	root.split = 0
	root.leaf = keep

	SetActiveTree(root)
	root.Resize(screen.Get().RootRect())
}
*/

// 破棄対象の leaf に後始末が必要な場合のみ実装する
func cleanupAfterCloseLeaf(l Leaf) {
	if l == nil {
		return
	}
	l.WillClose()
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
func (tr *Tree) nextInCycle() *Tree {
	root := tr
	for root.parent != nil {
		root = root.parent
	}

	var nodes []*Tree
	root.Traverse(func(w *Tree) {
		nodes = append(nodes, w)
	})
	if len(nodes) == 0 {
		return nil
	}

	// If tr is not in the tree (index -1), start from the first leaf
	// rather than silently treating tr as the first one.
	idx := slices.Index(nodes, tr)
	return nodes[(idx+1)%len(nodes)]
}

// Implement overlay
func (tr *Tree) IsActive() bool {
	return true
}
