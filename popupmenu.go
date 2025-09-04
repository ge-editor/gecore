package gecore

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/theme"
)

const maxThreshold = 2

func NewPopupmenu(rect utils.Rect, items []string, itemIndex int) *Popupmenu {
	pm := &Popupmenu{
		screen:     screen.Get(),
		items:      items,
		itemIndex:  itemIndex,
		Rect:       rect,
		KeyPointer: KeyMapper(),
	}
	return pm
}

type Popupmenu struct {
	screen     *screen.Screen
	items      []string
	itemIndex  int
	startIndex int
	utils.Rect
	dispHeight int
	*KeyPointer
}

// -----------------

func (pm *Popupmenu) Item() (int, string) {
	// verb.PP("items %d, %d, %v", pm.itemIndex, len(pm.items), pm.items)
	if pm == nil || len(pm.items) == 0 || pm.itemIndex < 0 || pm.itemIndex >= len(pm.items) {
		return -1, ""
	}
	return pm.itemIndex, pm.items[pm.itemIndex]
}

func (pm *Popupmenu) Index() int {
	return pm.itemIndex
}

func (pm *Popupmenu) Set(items []string, index int) {
	pm.items = items
	if index > len(pm.items)-1 {
		index = len(pm.items) - 1
	}
	pm.itemIndex = index
}

func (pm *Popupmenu) sliderPosAndRune() (int, rune) {
	max := len(pm.items) - pm.Height
	progress := int((float32(pm.startIndex) / float32(max)) * float32(pm.Height*2))
	/*
		if max <= 0 && len(m.items) > 0 {
			return 0, '▀'
		}
	*/
	if pm.startIndex == max {
		return pm.Height - 1, '▄'
	}
	var r rune
	if progress&1 != 0 {
		r = '▄'
	} else {
		r = '▀'
	}
	// m.screen.Echo(fmt.Sprintf("%d/2 %d", progress, progress/2))
	return progress / 2, r
}

// Calculate Popupmenu.startIndex and set
// Calculate startIndex from itemIndex and dispHeight.
func (pm *Popupmenu) calcStartIndex() {
	threshold := utils.Threshold(maxThreshold, pm.dispHeight)
	// threshold = 2

	if pm.itemIndex <= pm.startIndex+threshold {
		pm.startIndex = pm.itemIndex - threshold
	}

	if pm.itemIndex >= pm.startIndex+pm.dispHeight-threshold {
		pm.startIndex = pm.itemIndex + threshold - pm.dispHeight + 1
	}

	if pm.startIndex < 0 {
		pm.startIndex = 0
	}
	if pm.startIndex > len(pm.items)-pm.dispHeight {
		pm.startIndex = len(pm.items) - pm.dispHeight
	}
}

func (pm *Popupmenu) Draw() {
	if pm.itemIndex < 0 {
		pm.itemIndex = 0
	}
	if pm.itemIndex > len(pm.items)-1 {
		pm.itemIndex = len(pm.items) - 1
	}

	// x
	minW := pm.Width
	x := pm.X
	verb.PP("x %d, minW %d, width %d", x, minW, pm.screen.Width)
	c := x + minW - pm.screen.Width
	if c > 0 {
		x -= c
	}
	if x < 0 {
		x = 0
	}

	// y

	// above
	aStart, aEnd := pm.Y-1, pm.Y-1-(pm.Height-1)
	if aEnd < 0 {
		aEnd = 0
	}

	// bellow
	bStart, bEnd := pm.Y+1, pm.Y+1+(pm.Height-1)
	if bEnd >= pm.screen.Height {
		bEnd = pm.screen.Height - 1
	}

	if (aStart - aEnd) > (bEnd - bStart) {
		if (aStart - aEnd) > len(pm.items)-1 {
			aEnd = aStart - (len(pm.items) - 1)
		}

		// swap
		aStart, aEnd = aEnd, aStart
	} else {
		if (bEnd - bStart) > len(pm.items)-1 {
			bEnd = bStart + (len(pm.items) - 1)
		}

		// change
		aStart = bStart
		aEnd = bEnd
	}

	pm.dispHeight = aEnd - aStart + 1
	pm.calcStartIndex() // use m.dispHeight
	progress, ch := pm.sliderPosAndRune()

	j := 0
	i := pm.startIndex
	for y := aStart; y <= aEnd; y++ {
		var style tcell.Style
		if i == pm.itemIndex {
			style = theme.ColorPopupmenuForeground
		} else {
			style = theme.ColorPopupmenuBackground
		}
		pm.screen.DrawLabel(utils.Rect{X: x, Y: y, Width: pm.Width - 1, Height: 1},
			&screen.LabelParams{
				Style:          style,
				Align:          screen.AlignLeft,
				Ellipsis:       '…',
				CenterEllipsis: false,
			}, pm.items[i])

		r := ' '
		if progress == j {
			r = ch
		}
		pm.screen.SetContent(x+pm.Width-1, y, r, nil, theme.ColorPopupmenuForeground)

		i++
		j++
	}
}

//-------------------

func (pm *Popupmenu) CursorForward() {
	if pm.itemIndex >= len(pm.items)-1 {
		return
	}
	pm.itemIndex++
}

func (pm *Popupmenu) CursorBackward() {
	if pm.itemIndex == 0 {
		return
	}
	pm.itemIndex--
}

func (pm *Popupmenu) CursorHome() {
	pm.itemIndex = 0
}

func (pm *Popupmenu) CursorEnd() {
	pm.itemIndex = len(pm.items) - 1
}

// func (m *Popupmenu) Event(tev *EventKey) *EventKey {
func (pm *Popupmenu) Event(tev *tcell.EventKey) *tcell.EventKey {
	// m.Execute(tev, false)

	switch tev.Key() {
	case tcell.KeyCtrlF, tcell.KeyRight:
		fallthrough
	case tcell.KeyCtrlN, tcell.KeyDown:
		pm.CursorForward()
	case tcell.KeyCtrlB, tcell.KeyLeft:
		fallthrough
	case tcell.KeyCtrlP, tcell.KeyUp:
		pm.CursorBackward()
	case tcell.KeyCtrlE, tcell.KeyEnd:
		pm.CursorEnd()
	// mac delete-key is this
	case tcell.KeyCtrlA, tcell.KeyHome:
		pm.CursorHome()
	default:
	}

	return tev
}
