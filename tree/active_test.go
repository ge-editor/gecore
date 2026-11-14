package tree

import (
	"context"
	"testing"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/keychord"
)

type fakeLeafType struct{ name string }

func (t *fakeLeafType) NewLeaf() Leaf { return &fakeLeaf{lt: t} }
func (t *fakeLeafType) NewSiblingLeaf(direction string, leaf Leaf) Leaf {
	return &fakeLeaf{lt: t}
}
func (t *fakeLeafType) RealName() string                       { return "fake" }
func (t *fakeLeafType) Name() string                           { return t.name }
func (t *fakeLeafType) SetRegisteredName(n string)             { t.name = n }
func (t *fakeLeafType) SetCtx(*LeafContext)                    {}
func (t *fakeLeafType) CancelManager() *gecore.EventCancelManager { return nil }
func (t *fakeLeafType) DispatchLeafRequest(destination, behavior string, flag OpenerFlag, data any, f func(handled bool, err error)) {
}

type fakeLeaf struct {
	lt     LeafType
	active bool
}

func (l *fakeLeaf) LeafType() LeafType            { return l.lt }
func (l *fakeLeaf) Resize(screen.Rect)            {}
func (l *fakeLeaf) Draw(context.Context) bool     { return false }
func (l *fakeLeaf) Kill(Leaf, bool) Leaf          { return nil }
func (l *fakeLeaf) Active(a bool)                 { l.active = a }
func (l *fakeLeaf) DispatchMouse(tcell.EventMouse) {}
func (l *fakeLeaf) Resume()                       {}
func (l *fakeLeaf) Init()                         {}
func (l *fakeLeaf) WillClose()                    {}
func (l *fakeLeaf) Commands() []excommand.ExCommand { return nil }
func (l *fakeLeaf) DispatchKey(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	var t keychord.KeyDispatchTransition
	return "", t
}

func setupFakeTree(t *testing.T) {
	t.Helper()
	LeafTypes = NewLeafTypes()
	if err := LeafTypes.Register(func() LeafType { return &fakeLeafType{} }, 0, "fake"); err != nil {
		t.Fatal(err)
	}
	if err := LeafTypes.Register(func() LeafType { return &fakeLeafType{} }, 1, "other"); err != nil {
		t.Fatal(err)
	}
	currentRoot.Store(nil)
	RootTrees = nil
	SetRootTree(NewRootTree(LeafTypes.GetDefaultLeafType()))
	SetActiveTree(GetRootTree())
}

// exactly one leaf is active and it is the leaf of the active Tree.
func checkOneActive(t *testing.T, step string) {
	t.Helper()
	n := 0
	GetRootTree().ForEachLeaf(func(l Leaf) bool {
		if l.(*fakeLeaf).active {
			n++
		}
		return true
	})
	if n != 1 {
		t.Fatalf("%s: %d active leaves, want 1", step, n)
	}
	if !GetActiveTree().GetLeaf().(*fakeLeaf).active {
		t.Fatalf("%s: the leaf of the active tree is not active", step)
	}
}

func TestSplitKeepsSingleActiveLeaf(t *testing.T) {
	setupFakeTree(t)
	checkOneActive(t, "init")

	a := GetActiveTree()
	a.InsertRight(a)
	checkOneActive(t, "C-x 5")

	for i := 0; i < 4; i++ {
		GetActiveTree().NextInCycle()
		checkOneActive(t, "C-x o (2 windows)")
	}

	GetActiveTree().InsertBottom(GetActiveTree())
	checkOneActive(t, "C-x 2")

	for i := 0; i < 7; i++ {
		GetActiveTree().NextInCycle()
		checkOneActive(t, "C-x o (3 windows)")
	}

	// New leaf is placed before the old one (top / left).
	GetActiveTree().InsertTop(GetRootTree())
	checkOneActive(t, "insert top")
	GetActiveTree().InsertLeft(GetRootTree())
	checkOneActive(t, "insert left")

	GetActiveTree().SplitHorizontally()
	checkOneActive(t, "split horizontally")
	GetActiveTree().SplitVertically()
	checkOneActive(t, "split vertically")
}

func TestNextInCycleVisitsEveryWindowInOrder(t *testing.T) {
	setupFakeTree(t)
	a := GetActiveTree()
	a.InsertRight(a)
	GetActiveTree().InsertBottom(GetActiveTree())

	var order []*Tree
	GetRootTree().Traverse(func(w *Tree) { order = append(order, w) })
	if len(order) != 3 {
		t.Fatalf("got %d windows, want 3", len(order))
	}

	SetActiveTree(order[0])
	for i := 1; i <= 6; i++ {
		GetActiveTree().NextInCycle()
		want := order[i%3]
		if GetActiveTree() != want {
			t.Fatalf("step %d: moved to the wrong window", i)
		}
	}
}

// every child's parent points at the node that holds it.
func checkParents(t *testing.T, tr *Tree, step string) {
	t.Helper()
	for _, c := range []*Tree{tr.left, tr.right, tr.top, tr.bottom} {
		if c == nil {
			continue
		}
		if c.parent != tr {
			t.Fatalf("%s: child has a stale parent pointer", step)
		}
		checkParents(t, c, step)
	}
}

func leafNames() []string {
	var names []string
	GetRootTree().ForEachLeaf(func(l Leaf) bool {
		names = append(names, l.LeafType().Name())
		return true
	})
	return names
}

// L(other) | R1(fake) / R2(fake); active = R2
func setupThreeWindows(t *testing.T) (l, r1, r2 *Tree) {
	t.Helper()
	setupFakeTree(t)
	a := GetActiveTree()
	a.InsertRight(a)
	GetActiveTree().InsertBottom(GetActiveTree())

	var order []*Tree
	GetRootTree().Traverse(func(w *Tree) { order = append(order, w) })
	if len(order) != 3 {
		t.Fatalf("got %d windows, want 3", len(order))
	}
	if err := order[0].SelectLeaf("other"); err != nil {
		t.Fatal(err)
	}
	SetActiveTree(order[2])
	return order[0], order[1], order[2]
}

func TestDeleteOtherWindowsTargets(t *testing.T) {
	_, _, r2 := setupThreeWindows(t)
	keep := r2.GetLeaf()

	// Only "fake" leaves are closed; the "other" leaf stays.
	if !deleteOtherWindows([]string{"fake"}, nil) {
		t.Fatal("tree was not changed")
	}
	names := leafNames()
	if len(names) != 2 || names[0] != "other" || names[1] != "fake" {
		t.Fatalf("leaves after close = %v, want [other fake]", names)
	}
	checkParents(t, GetRootTree(), "after targets")
	if GetActiveTree().GetLeaf() != keep {
		t.Fatal("active window is not the kept leaf")
	}
	checkOneActive(t, "after targets")
}

func TestDeleteOtherWindowsExcludes(t *testing.T) {
	setupThreeWindows(t)

	// Closes everything except the active leaf and "other".
	deleteOtherWindows(nil, []string{"other"})
	names := leafNames()
	if len(names) != 2 || names[0] != "other" || names[1] != "fake" {
		t.Fatalf("leaves after close = %v, want [other fake]", names)
	}
	checkParents(t, GetRootTree(), "after excludes")
	checkOneActive(t, "after excludes")
}

func TestDeleteOtherWindowsAll(t *testing.T) {
	_, _, r2 := setupThreeWindows(t)
	keep := r2.GetLeaf()

	deleteOtherWindows(nil, nil)
	if GetRootTree().leaf != keep {
		t.Fatal("root does not hold only the kept leaf")
	}
	if GetActiveTree() != GetRootTree() {
		t.Fatal("active tree is not the root")
	}
	checkOneActive(t, "after all")
}

func TestDeleteOtherWindowsKeepsTreeConsistent(t *testing.T) {
	// A nested split where the surviving subtree is promoted.
	setupFakeTree(t)
	a := GetActiveTree()
	a.InsertRight(a)
	GetActiveTree().InsertBottom(GetActiveTree())
	GetActiveTree().InsertRight(GetActiveTree())
	// leaves: L | R1 / (R2a | R2b); active = R2b

	deleteOtherWindows([]string{"fake"}, nil)
	checkParents(t, GetRootTree(), "nested")
	checkOneActive(t, "nested")
	for i := 0; i < 3; i++ {
		GetActiveTree().NextInCycle()
		checkOneActive(t, "cycle after close")
	}
}
