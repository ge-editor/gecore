package tree

import (
	"context"
	"errors"
	"slices"
	"sync/atomic"

	"github.com/ge-editor/gecore/screen"
)

// RootTreeStruct is one screen layout (a Tree of windows) that can be
// switched to. RootTrees (this package) is the single owner; a Workspace
// only keeps cross references (see RemoveRootTree).
type RootTreeStruct struct {
	Name string
	Root *Tree

	// Active is the active window of this RootTree. It is kept per
	// RootTree so that switching away and back restores the focus.
	Active *Tree

	// Owner is an opaque reference to the owning Workspace. tree does not
	// import the workspace package.
	Owner any
}

// RootTrees holds every RootTree, shown or not.
// Mutate and iterate it from the main (UI) goroutine only. The draw
// goroutine must use CurrentRootOverlay / CurrentRootTree instead.
var RootTrees []*RootTreeStruct

// currentRoot is the RootTree being displayed. It is atomic because the
// draw goroutine reads it while the main goroutine may switch it.
var currentRoot atomic.Pointer[RootTreeStruct]

// CurrentRootTree returns the RootTree being displayed (nil before init).
func CurrentRootTree() *RootTreeStruct {
	return currentRoot.Load()
}

// AddRootTree registers rt. Registering the same rt twice is a no-op.
func AddRootTree(rt *RootTreeStruct) {
	if rt == nil || slices.Contains(RootTrees, rt) {
		return
	}
	RootTrees = append(RootTrees, rt)
}

// RemoveRootTree unregisters rt and calls WillClose on its leaves.
// The current RootTree cannot be removed: SwitchRootTree to another one
// first. The caller (Workspace.Close) also drops its own reference.
func RemoveRootTree(rt *RootTreeStruct) error {
	i := slices.Index(RootTrees, rt)
	if i < 0 {
		return errors.New("tree: root tree is not registered")
	}
	if currentRoot.Load() == rt {
		return errors.New("tree: cannot remove the current root tree; switch first")
	}
	if rt.Root != nil {
		rt.Root.ForEachLeaf(func(l Leaf) {
			if l != nil {
				l.WillClose()
			}
		})
	}
	RootTrees = slices.Delete(RootTrees, i, i+1)
	rt.Active = nil
	return nil
}

// SwitchRootTree makes rt the displayed RootTree.
//
// It does not lay out the screen. The caller must then call
// overlay.OverlayManager().Layout(screen.Get().Rect) (and request a draw)
// so that the leaves of rt are resized.
func SwitchRootTree(rt *RootTreeStruct) error {
	if rt == nil || rt.Root == nil {
		return errors.New("tree: invalid root tree")
	}
	if !slices.Contains(RootTrees, rt) {
		return errors.New("tree: root tree is not registered")
	}
	old := currentRoot.Load()
	if old == rt {
		return nil
	}

	if old != nil && old.Active != nil {
		if l := old.Active.GetLeaf(); l != nil {
			l.Active(false)
		}
	}
	if rt.Active == nil {
		rt.Active = rt.Root.firstLeafNode()
	}
	currentRoot.Store(rt)
	if l := rt.Active.GetLeaf(); l != nil {
		l.Active(true)
	}

	// Window swap helpers refer to *Tree of the previous RootTree.
	tree1, tree2 = nil, nil
	return nil
}

// ForEachLeafAll calls fn for every leaf of every RootTree, including
// the ones that are not displayed. Use it instead of
// GetRootTree().ForEachLeaf wherever state shared through an EditBuffer
// (cursor, boundary cache, marks) must be kept in sync.
// fn must not add or remove RootTrees.
func ForEachLeafAll(fn func(Leaf)) {
	for _, rt := range RootTrees {
		if rt.Root != nil {
			rt.Root.ForEachLeaf(fn)
		}
	}
}

// CurrentRootOverlay implements overlay.Overlay and always delegates to the
// RootTree being displayed, so that overlay.SetTree does not have to be
// called again when the RootTree is switched.
type CurrentRootOverlay struct{}

func (CurrentRootOverlay) RequiredHeight() int { return -1 }
func (CurrentRootOverlay) IsActive() bool      { return true }

func (CurrentRootOverlay) Resize(rect screen.Rect) {
	if rt := currentRoot.Load(); rt != nil && rt.Root != nil {
		rt.Root.Resize(rect)
	}
}

func (CurrentRootOverlay) Draw(ctx context.Context) bool {
	if rt := currentRoot.Load(); rt != nil && rt.Root != nil {
		return rt.Root.Draw(ctx)
	}
	return false
}
