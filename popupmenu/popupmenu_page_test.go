package popupmenu

import (
	"testing"

	"github.com/ge-editor/gecore/screen"
)

// PgDn / PgUp move the selection by the number of rows that are shown and stop
// at the ends.
func TestCursorPage(t *testing.T) {
	p := NewPopupmenu()
	items := make([]string, 50)
	for i := range items {
		items[i] = "item"
	}

	// Resize takes one row for the space above the popup: 10 rows are shown.
	p.Resize(screen.Rect{X: 0, Y: 0, Width: 80, Height: 11})
	p.SetItems(items, 0)

	forward := []int{10, 20, 30, 40, 49, 49}
	for _, want := range forward {
		p.CursorPageForward()
		if i, _ := p.GetItem(); i != want {
			t.Fatalf("PageForward: index = %d, want %d", i, want)
		}
	}

	backward := []int{39, 29, 19, 9, 0, 0}
	for _, want := range backward {
		p.CursorPageBackward()
		if i, _ := p.GetItem(); i != want {
			t.Fatalf("PageBackward: index = %d, want %d", i, want)
		}
	}
}

// Before the first layout no row is shown: a page is one item, and an empty
// menu does not move.
func TestCursorPageWithoutLayout(t *testing.T) {
	p := NewPopupmenu()
	p.CursorPageForward() // empty: must not panic
	p.CursorPageBackward()

	p.SetItems([]string{"a", "b", "c"}, 0)
	p.CursorPageForward()
	if i, _ := p.GetItem(); i != 1 {
		t.Fatalf("index = %d, want 1", i)
	}
	p.CursorPageBackward()
	if i, _ := p.GetItem(); i != 0 {
		t.Fatalf("index = %d, want 0", i)
	}
}
