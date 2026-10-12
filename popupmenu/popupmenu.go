// gecore/popupmenu/popupmenu.go
package popupmenu

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/keychord"
	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

const MIN_WIDTH = 21

type Session struct {
	Popupmenu *PopupmenuStruct
	Keymap    *keychord.RootNode
}

func NewSession(
	items []string,
	bind func(*keychord.RootNode, *PopupmenuStruct),
) *Session {
	pm := NewPopupmenu()
	pm.SetItems(items, 0)

	km := keychord.NewRootNode()
	if bind != nil {
		bind(km, pm)
	}
	pm.SetKeyDispatcher(km)
	pm.Active(true)

	return &Session{
		Popupmenu: pm,
		Keymap:    km,
	}
}

// ------------------------

const CURSOR_THRESHOLD = 2

type PopupmenuStruct struct {
	columnTitle string
	items       []string
	itemIndex   int
	active      bool

	overlayRect screen.Rect // popupmenu を配置する領域
	popupRect   screen.Rect // popupmenu の位置とサイズ

	startIndex int

	keyDispatcher *keychord.RootNode
}

func NewPopupmenu() *PopupmenuStruct {
	return &PopupmenuStruct{
		items:     []string{},
		itemIndex: 0,
		popupRect: screen.Rect{
			X:      8,
			Y:      10,
			Width:  40,
			Height: 20,
		},
	}
}

func (p *PopupmenuStruct) Active(b bool) {
	p.active = b
}

func (p *PopupmenuStruct) IsActive() bool {
	return p.active
}

func (p *PopupmenuStruct) SetKeyDispatcher(km *keychord.RootNode) {
	p.keyDispatcher = km
}

// Overlay interface
func (p *PopupmenuStruct) RequiredHeight() int {
	return p.popupRect.Height
}

func (p *PopupmenuStruct) Resize(rect screen.Rect) {
	p.overlayRect = rect
	// p.overlayRect.Height -= 2 // top space + mode line
	p.overlayRect.Height -= 1 // top space
	p.overlayRect.Y += 1      // top space

	p.calcPupupHeight()
	p.calcPupupWidth()
}

/*
	 func (p *PopupmenuStruct) SetPopupRect(rect screen.Rect) {
		p.popupRect = rect
	}
*/

func (p *PopupmenuStruct) SetPopupPos(x, y int) {
	p.popupRect.X = x
	p.popupRect.Y = y
}

func (p *PopupmenuStruct) SetColumnTitle(title string) {
	p.columnTitle = title
	p.calcPupupHeight()
	p.calcPupupWidth()
}

func (p *PopupmenuStruct) Draw() {
	if !p.active || len(p.items) == 0 {
		return
	}

	// clamp index
	if p.itemIndex < 0 {
		p.itemIndex = 0
	}
	if p.itemIndex >= len(p.items) {
		p.itemIndex = len(p.items) - 1
	}

	p.calcStartIndex()

	progress, ch := p.sliderPosAndRune()
	rowOffset := 0
	if p.columnTitle != "" {
		screen.Get().DrawLabel(
			screen.Rect{X: p.popupRect.X, Y: p.popupRect.Y, Width: p.popupRect.Width - 1, Height: 1},
			&screen.LabelParams{Style: theme.ColorPopupmenuForeground, Align: screen.AlignLeft, Ellipsis: '…'},
			" "+p.columnTitle,
		)
		screen.Get().SetContent(p.popupRect.X+p.popupRect.Width-1, p.popupRect.Y, ' ', nil, theme.ColorPopupmenuForeground)
		rowOffset = 1
	}

	for row := 0; row < p.contentHeight(); row++ {
		i := p.startIndex + row
		if i >= len(p.items) {
			break
		}

		style := theme.ColorPopupmenuBackground
		if i == p.itemIndex {
			style = theme.ColorPopupmenuForeground
		}

		screen.Get().DrawLabel(
			screen.Rect{
				X:      p.popupRect.X,
				Y:      p.popupRect.Y + rowOffset + row,
				Width:  p.popupRect.Width - 1,
				Height: 1,
			},
			&screen.LabelParams{
				Style:    style,
				Align:    screen.AlignLeft,
				Ellipsis: '…',
			},
			" "+p.items[i],
		)

		r := ' '
		if row == progress {
			r = ch
		}
		screen.Get().SetContent(p.popupRect.X+p.popupRect.Width-1, p.popupRect.Y+rowOffset+row, r, nil, theme.ColorPopupmenuForeground)
	}
}

func (p *PopupmenuStruct) DispatchKey(ev tcell.EventKey) (string, keychord.KeyDispatchTransition) {
	if p.keyDispatcher == nil {
		// gelog.Info("pm.keyDispatcher == nil")
		return "", keychord.DispatchNotFound
	}

	s, res := p.keyDispatcher.Dispatch(ev)
	/* 	switch res {
	   	case keychord.DispatchNotFound:
	   	case keychord.DispatchPrefix:
	   	case keychord.DispatchExecuted:
	   	}
	*/
	return s, res
}

// ------------------------
// Items
// ------------------------

func (p *PopupmenuStruct) SetItems(items []string, index int) {
	p.items = items
	if index < 0 {
		index = 0
	}
	if index >= len(items) {
		index = len(items) - 1
	}
	p.itemIndex = index

	p.calcPupupHeight()
	p.calcPupupWidth()
}

func (p *PopupmenuStruct) GetItem() (int, string) {
	if len(p.items) == 0 {
		return -1, ""
	}
	return p.itemIndex, p.items[p.itemIndex]
}

// ------------------------
// Cursor
// ------------------------

func (p *PopupmenuStruct) CursorForward() {
	if p.itemIndex < len(p.items)-1 {
		p.itemIndex++
	}
}

func (p *PopupmenuStruct) CursorBackward() {
	if p.itemIndex > 0 {
		p.itemIndex--
	}
}

func (p *PopupmenuStruct) CursorHome() {
	p.itemIndex = 0
}

func (p *PopupmenuStruct) CursorEnd() {
	p.itemIndex = len(p.items) - 1
}

// ------------------------
// Scroll logic
// ------------------------

func (p *PopupmenuStruct) calcPupupHeight() {
	itemHeight := p.overlayRect.Height
	if p.columnTitle != "" {
		itemHeight--
	}
	if itemHeight < 0 {
		itemHeight = 0
	}
	if len(p.items) < itemHeight {
		itemHeight = len(p.items)
	}
	height := itemHeight
	if p.columnTitle != "" && p.overlayRect.Height > 0 {
		height++
	}

	// in case of interaction minibuffer
	p.popupRect.Height = height
	p.popupRect.Y = p.overlayRect.Y + (p.overlayRect.Height - height)
}

func (p *PopupmenuStruct) calcPupupWidth() {
	itemMaxWidth := MIN_WIDTH
	if p.columnTitle != "" {
		if w := utils.WidthOnScreen([]byte(p.columnTitle)) + 2; w > itemMaxWidth {
			itemMaxWidth = w
		}
	}
	for _, item := range p.items {
		w := utils.WidthOnScreen([]byte(item)) + 2 // +2: 多分 drawlabel がおかしい?
		if w > itemMaxWidth {
			itemMaxWidth = w
		}
	}
	width := p.overlayRect.Width - p.popupRect.X - 3 // -3: scrollbar of popup, left space of popup and right space of screen
	if itemMaxWidth < width {
		width = itemMaxWidth
	}
	p.popupRect.Width = width
}

func (p *PopupmenuStruct) calcStartIndex() {
	height := p.contentHeight()
	if height == 0 {
		p.startIndex = 0
		return
	}
	threshold := utils.Threshold(CURSOR_THRESHOLD, height)

	if p.itemIndex <= p.startIndex+threshold {
		p.startIndex = p.itemIndex - threshold
	}
	if p.itemIndex >= p.startIndex+height-threshold {
		p.startIndex = p.itemIndex + threshold - height + 1
	}

	max := len(p.items) - height
	if p.startIndex > max {
		p.startIndex = max
	}

	if p.startIndex < 0 {
		p.startIndex = 0
	}
}

func (p *PopupmenuStruct) contentHeight() int {
	height := p.popupRect.Height
	if p.columnTitle != "" {
		height--
	}
	if height < 0 {
		return 0
	}
	return height
}

func (p *PopupmenuStruct) sliderPosAndRune() (int, rune) {
	height := p.contentHeight()
	if len(p.items) <= 1 || height <= 0 {
		return 0, ' '
	}

	// Position the thumb from the selected item rather than the scroll offset,
	// so it moves from the first item and reaches the bottom at the last item.
	steps := height*2 - 1
	indexRange := len(p.items) - 1
	progress := (p.itemIndex*steps + indexRange - 1) / indexRange
	if p.itemIndex >= len(p.items)-1 {
		return height - 1, '▄'
	}
	if progress >= steps {
		progress = steps - 1
	}

	if progress&1 == 0 {
		return progress / 2, '▀'
	}
	return progress / 2, '▄'
}

/*
func (p *PopupmenuStruct) sliderPosAndRune() (int, rune) {
	max := len(p.items) - p.popupRect.Height
	if max <= 0 {
		return 0, ' '
	}

	progress := int(float32(p.startIndex) / float32(max) * float32(p.popupRect.Height*2))
	if p.startIndex == max {
		return p.popupRect.Height - 1, '▄'
	}

	if progress&1 == 0 {
		return progress / 2, '▀'
	}
	return progress / 2, '▄'
}
*/
