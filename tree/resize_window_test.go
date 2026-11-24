package tree

import (
	"testing"

	"github.com/ge-editor/gecore/screen"
)

// hsplit builds a tree split into a left and a right window, 100 cells wide.
func hsplit() (root, left, right *Tree) {
	left = &Tree{leaf: &fakeLeaf{}}
	right = &Tree{leaf: &fakeLeaf{}}
	root = &Tree{left: left, right: right, split: 0.5}
	left.parent, right.parent = root, root
	root.Resize(screen.Rect{Width: 101, Height: 40})
	return
}

// vsplit builds a tree split into a top and a bottom window, 40 lines high.
func vsplit() (root, top, bottom *Tree) {
	top = &Tree{leaf: &fakeLeaf{}}
	bottom = &Tree{leaf: &fakeLeaf{}}
	root = &Tree{top: top, bottom: bottom, split: 0.5}
	top.parent, bottom.parent = root, root
	root.Resize(screen.Rect{Width: 101, Height: 40})
	return
}

func TestStepResizeWindowHorizontal(t *testing.T) {
	_, left, right := hsplit()
	l0, r0 := left.Width, right.Width

	left.StepResizeWindow(true, 1) // enlarge the left window
	if left.Width <= l0 || right.Width >= r0 {
		t.Errorf("enlarge left: widths %d,%d -> %d,%d", l0, r0, left.Width, right.Width)
	}

	l1, r1 := left.Width, right.Width
	right.StepResizeWindow(true, 1) // enlarge the right window
	if right.Width <= r1 || left.Width >= l1 {
		t.Errorf("enlarge right: widths %d,%d -> %d,%d", l1, r1, left.Width, right.Width)
	}

	l2, r2 := left.Width, right.Width
	right.StepResizeWindow(true, -1) // shrink the right window
	if right.Width >= r2 || left.Width <= l2 {
		t.Errorf("shrink right: widths %d,%d -> %d,%d", l2, r2, left.Width, right.Width)
	}
}

func TestStepResizeWindowVertical(t *testing.T) {
	_, top, bottom := vsplit()
	t0, b0 := top.Height, bottom.Height

	top.StepResizeWindow(false, 1)
	if top.Height <= t0 || bottom.Height >= b0 {
		t.Errorf("enlarge top: heights %d,%d -> %d,%d", t0, b0, top.Height, bottom.Height)
	}

	t1, b1 := top.Height, bottom.Height
	bottom.StepResizeWindow(false, 1)
	if bottom.Height <= b1 || top.Height >= t1 {
		t.Errorf("enlarge bottom: heights %d,%d -> %d,%d", t1, b1, top.Height, bottom.Height)
	}
}

// A window in a split of the other direction only is not resized.
func TestStepResizeWindowNoSplit(t *testing.T) {
	_, top, bottom := vsplit()
	t0, b0 := top.Height, bottom.Height
	top.StepResizeWindow(true, 1) // there is no horizontal split
	if top.Height != t0 || bottom.Height != b0 {
		t.Errorf("heights changed: %d,%d -> %d,%d", t0, b0, top.Height, bottom.Height)
	}
}

// The nearest split of the direction is used, through splits of the other
// direction.
func TestStepResizeWindowNested(t *testing.T) {
	root, left, right := hsplit()
	top := &Tree{leaf: &fakeLeaf{}}
	bottom := &Tree{leaf: &fakeLeaf{}}
	right.leaf = nil
	right.top, right.bottom, right.split = top, bottom, 0.5
	top.parent, bottom.parent = right, right
	root.Resize(root.Rect)

	l0, b0 := left.Width, bottom.Width
	bottom.StepResizeWindow(true, 1) // enlarge the bottom-right window to the left
	if left.Width >= l0 || bottom.Width <= b0 {
		t.Errorf("widths: left %d -> %d, bottom %d -> %d", l0, left.Width, b0, bottom.Width)
	}
}
