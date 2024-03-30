package tree

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/theme"
)

//----------------------------------------------------------
// view op mode
// split screen mode
//----------------------------------------------------------

type OpMode struct {
	Screen *screen.Screen
}

func NewOpMode() *OpMode {
	return &OpMode{
		Screen: screen.Get(),
	}
}

const viewNames = `1234567890abcdefgijlmnpqstuwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

func (v OpMode) Draw() {
	// draw views names
	name := 0
	rootTree.Traverse(func(leaf *Tree) {
		if name >= len(viewNames) {
			return
		}
		bg := tcell.ColorBlue
		if leaf == ActiveTreeGet() {
			bg = tcell.ColorRed
		}
		r := leaf.Rect
		r.Width = 3
		r.Height = 1
		v.Screen.DrawLabel(r, &screen.LabelParams{Style: theme.ColorDefault.Background(bg), Align: screen.AlignCenter, Ellipsis: 0, CenterEllipsis: false}, string(viewNames[name]))
		name++
	})

	// draw splitters
	// r = v.tree.Rect
	r := ActiveTreeGet().Rect
	var x, y int

	// horizontal ----------------------
	hr := r
	hr.X += (r.Width - 1) / 2
	hr.Width = 1
	hr.Height = 3
	v.Screen.Fill(hr, screen.Cell{
		Style: theme.ColorDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed),
		Ch:    '|',
	})

	x = hr.X
	y = hr.Y + 1
	v.Screen.SetContent(x, y, 'h', nil, theme.ColorDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed).Attributes(tcell.AttrBold))

	// vertical ----------------------
	vr := r
	vr.Y += (r.Height - 1) / 2
	vr.Height = 1
	vr.Width = 5
	v.Screen.DrawLabel(vr, &screen.LabelParams{Style: theme.ColorDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed), Align: screen.AlignCenter, Ellipsis: 0, CenterEllipsis: false}, "--v--")
}

func (v OpMode) SelectName(ch rune) *Tree {
	var sel *Tree = nil
	name := 0
	// v.tree.Traverse(func(leaf *tree.Tree) {
	rootTree.Traverse(func(leaf *Tree) {
		if name >= len(viewNames) {
			return
		}
		if rune(viewNames[name]) == ch {
			sel = leaf
		}
		name++
	})

	return sel
}
