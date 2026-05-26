package tree

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/theme"
)

//----------------------------------------------------------
// view op mode
// split screen mode
//----------------------------------------------------------

type OpMode struct {
	Screen    *screen.Screen
	viewNames string
}

func NewOpMode(viewNames string) *OpMode {
	return &OpMode{
		Screen:    screen.Get(),
		viewNames: viewNames,
	}
}

func (v OpMode) Draw() {
	// draw views names
	name := 0
	rootTree.Traverse(func(leaf *Tree) {
		if name >= len(v.viewNames) {
			return
		}
		bg := tcell.ColorNames["blue"]
		if leaf == ActiveTreeGet() {
			bg = tcell.ColorNames["red"]
		}
		r := leaf.Rect
		r.Width = 3
		r.Height = 1
		v.Screen.DrawLabel(r, &screen.LabelParams{Style: theme.ColorDefault.Background(bg), Align: screen.AlignCenter, Ellipsis: 0, CenterEllipsis: false}, string(v.viewNames[name]))
		name++
	})

	// draw splitters
	// r = v.tree.Rect
	r := ActiveTreeGet().Rect
	var x, y int

	// horizontal ----------------------
	hRect := r
	hRect.X += (r.Width - 1) / 2
	hRect.Width = 1
	hRect.Height = 3
	v.Screen.FillRect(hRect, '|', theme.ColorDefault.Foreground(tcell.ColorNames["white"]).Background(tcell.ColorNames["red"]))

	x = hRect.X
	y = hRect.Y + 1
	v.Screen.SetContent(x, y, 'h', nil, theme.ColorDefault.Foreground(tcell.ColorNames["white"]).Background(tcell.ColorNames["red"]).Bold(true))

	// vertical ----------------------
	vRect := r
	vRect.Y += (r.Height - 1) / 2
	vRect.Height = 1
	vRect.Width = 5
	v.Screen.DrawLabel(vRect, &screen.LabelParams{Style: theme.ColorDefault.Foreground(tcell.ColorNames["white"]).Background(tcell.ColorNames["red"]), Align: screen.AlignCenter, Ellipsis: 0, CenterEllipsis: false}, "--v--")
}

func (v OpMode) SelectName(str string) *Tree {
	if len(str) == 0 {
		return nil
	}
	ch := []rune(str)[0]

	var sel *Tree = nil
	name := 0
	// v.tree.Traverse(func(leaf *tree.Tree) {
	rootTree.Traverse(func(leaf *Tree) {
		if name >= len(v.viewNames) {
			return
		}
		if rune(v.viewNames[name]) == ch {
			sel = leaf
		}
		name++
	})

	return sel
}
