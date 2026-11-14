package screen

import (
	"testing"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/utils"
)

// grid is a cellScreen that remembers where FillRect writes. A wide character
// is a cell of width 2 followed by a cell that has width 0.
type grid struct {
	w, h   int
	ch     map[[2]int]rune
	st     map[[2]int]tcell.Style
	writes [][2]int
}

func newGrid(w, h int) *grid {
	return &grid{w: w, h: h, ch: map[[2]int]rune{}, st: map[[2]int]tcell.Style{}}
}

func (g *grid) Get(x, y int) (string, tcell.Style, int) {
	if x < 0 || y < 0 || x >= g.w || y >= g.h {
		return "", tcell.StyleDefault, 0
	}
	r, ok := g.ch[[2]int{x, y}]
	if !ok {
		return " ", tcell.StyleDefault, 1
	}
	return string(r), g.st[[2]int{x, y}], utils.RuneWidth(r)
}

func (g *grid) SetContent(x, y int, r rune, _ []rune, st tcell.Style) {
	if x < 0 || y < 0 || x >= g.w || y >= g.h {
		return
	}
	g.writes = append(g.writes, [2]int{x, y})
	g.ch[[2]int{x, y}] = r
	g.st[[2]int{x, y}] = st
}

func (g *grid) wide(x, y int, st tcell.Style) {
	g.ch[[2]int{x, y}] = '日'
	g.st[[2]int{x, y}] = st
}

// The row below must not be touched: rows above the cursor are drawn from the
// bottom up, so the row below may already be drawn. A write at (endX, y+1)
// blanked the wide character that sat on the column limit of the row below.
func TestFillRectDoesNotWriteOnTheRowBelow(t *testing.T) {
	over := tcell.StyleDefault.Background(tcell.ColorRed)
	g := newGrid(20, 4)
	g.wide(9, 1, over) // old content of the filled row: a wide character at endX
	g.wide(9, 2, over) // the row below, already drawn: a wide character at endX

	bounds := Rect{X: 0, Y: 0, Width: 20, Height: 4}
	fillRect(g, bounds, Rect{X: 2, Y: 1, Width: 8, Height: 1}, 0, tcell.StyleDefault)

	if r := g.ch[[2]int{9, 2}]; r != '日' {
		t.Errorf("the character on the row below was overwritten with %q", r)
	}
	for _, p := range g.writes {
		if p[1] != 1 {
			t.Errorf("wrote outside the filled row: %v", p)
		}
	}
}

// A wide character cut by the right edge loses both halves, on its own row.
func TestFillRectBlanksWideCharacterCutOnTheRight(t *testing.T) {
	g := newGrid(20, 3)
	g.wide(9, 1, tcell.StyleDefault)

	bounds := Rect{X: 0, Y: 0, Width: 20, Height: 3}
	fillRect(g, bounds, Rect{X: 2, Y: 1, Width: 8, Height: 1}, ' ', tcell.StyleDefault)

	if r := g.ch[[2]int{10, 1}]; r != ' ' {
		t.Errorf("the right half was left: %q", r)
	}
}

// A wide character cut by the left edge loses both halves too.
func TestFillRectBlanksWideCharacterCutOnTheLeft(t *testing.T) {
	g := newGrid(20, 3)
	g.wide(1, 1, tcell.StyleDefault) // cells 1 and 2

	bounds := Rect{X: 0, Y: 0, Width: 20, Height: 3}
	fillRect(g, bounds, Rect{X: 2, Y: 1, Width: 8, Height: 1}, ' ', tcell.StyleDefault)

	if r := g.ch[[2]int{1, 1}]; r != ' ' {
		t.Errorf("the left half was left: %q", r)
	}
}
