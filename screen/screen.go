package screen

import (
	"fmt"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

var screen *Screen

func Get() *Screen {
	return screen
}

func init() {
	var err error
	screen, err = newScreen()
	if err != nil {
		fmt.Println(err)
	}
}

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

type Screen struct {
	tcell.Screen
	utils.Rect
	CX, CY int // cursor position
	echo   []string
}

// called to get the new size after a resize.
func newScreen() (*Screen, error) {
	tcellScreen, err := tcell.NewScreen()
	if err != nil {
		return screen, err
	}
	err = tcellScreen.Init()
	if err != nil {
		return screen, err
	}

	w, h := tcellScreen.Size()

	screen = &Screen{
		Screen: tcellScreen,
		Rect:   utils.Rect{X: 0, Y: 0, Width: w, Height: h},
	}
	return screen, nil
}

// Return Screen Rect without Minibuffer
func (s *Screen) RootRect() utils.Rect {
	return utils.Rect{X: 0, Y: 0, Width: s.Width, Height: s.Height - 1}
}

// ShowCursor sets the cursor position to (x, y).
func (s *Screen) ShowCursor(x, y int) {
	s.CX, s.CY = x, y
	s.Screen.ShowCursor(x, y)
}

func (s *Screen) HideCursor() {
	s.CX, s.CY = -s.CX, -s.CY
	s.Screen.ShowCursor(-1, -1)
}

// Fills an area which is an intersection between buffer and 'dest' with 'proto'.
/*
func (s *Screen) Fill(dest utils.Rect, proto Cell) {
	// m.unsafe_fill(m.Rect.Intersection(dst), proto)
	// Unsafe part of the fill operation, doesn't check for bounds.
	// func (m *Screen) unsafe_fill(dest utils.Rect, proto Cell) {
	dest = s.Rect.Intersection(dest)
	// gelog.Info("unsafe fill proto='%#v', dest='%#v'", proto, dest)
	runeWidth := utils.RuneWidth(proto.Ch)
	// stride := m.Width
	// off := m.Width*dest.Y + dest.X
	// 1個前の cell の文字幅が 2なら cell を空ににする
	for y := 0; y < dest.Height; y++ {
		for x := 0; x < dest.Width; x++ {
			if x == 0 {
				// 1個前の cell の文字幅が 2なら cell を空ににする
				// if _, _, _, w := s.Screen.GetContent(dest.X-1, dest.Y); w == 2 { // tcell/v2
				if _, _, w := s.Screen.Get(dest.X-1, dest.Y); w == 2 { // tcell/v3
					s.Screen.SetContent(dest.X-1, dest.Y, 0, nil, theme.ColorDefault)
				}
			}
			if runeWidth == 2 {
				if dest.X+x+runeWidth <= dest.Width {
					s.Screen.SetContent(dest.X+x, dest.Y+y, proto.Ch, nil, proto.Style)
					x++
					s.Screen.SetContent(dest.X+x, dest.Y+y, 0, nil, proto.Style)
				} else {
					s.Screen.SetContent(dest.X+x, dest.Y+y, 0, nil, proto.Style)
				}
			} else {
				s.Screen.SetContent(dest.X+x, dest.Y+y, proto.Ch, nil, proto.Style)
			}
		}
		// off += stride
	}
	// }
}
*/

func (s *Screen) FillRect(dest utils.Rect, r rune, style tcell.Style) {
	dest = s.Rect.Intersection(dest)
	if dest.Width <= 0 || dest.Height <= 0 {
		return
	}

	step := utils.RuneWidth(r)
	if step < 1 {
		step = 1
	} // 最低1幅を保証

	startY := dest.Y
	endY := dest.Y + dest.Height - 1
	startX := dest.X
	endX := dest.X + dest.Width - 1

	for y := startY; y <= endY; y++ {
		// --- 左端のクリーンアップ ---
		// startX に書き込むことで、startX-1 にある幅2の文字の右半分が消えるのを防ぐ
		if _, st, w := s.Screen.Get(startX-1, y); w == 2 {
			s.Screen.SetContent(startX-1, y, ' ', nil, st)
		}

		// --- 右端のクリーンアップ ---
		// endX (またはその付近) に書き込むことで、endX+1 にある幅2の文字が分断されるのを防ぐ
		// 以下のループ内で処理しても良いが、事前に境界をクリアするのが確実
		if _, st, w := s.Screen.Get(endX, y); w == 2 {
			s.Screen.SetContent(endX, y, ' ', nil, style)
			s.Screen.SetContent(endX, y+1, ' ', nil, st)
		}

		// --- メインの描画ループ ---
		for x := startX; x <= endX; {
			// 残りの幅が step 未満なら、文字を置けないので空白で埋める
			if x+step-1 > endX {
				s.Screen.SetContent(x, y, ' ', nil, style)
				x++
				continue
			}

			s.Screen.SetContent(x, y, r, nil, style)
			x += step // 文字幅分だけ進める
		}
	}
}

// Sets a cell at specified position
/*
func (m *Screen) Set(x, y int, proto Cell) {
	runewidth := utils.RuneWidth(proto.Ch)

	//pp("top of Set(x=%v, y=%v). this='%#v'", x, y, this)
	// if x < 0 || x >= m.Width {
	if x < 0 || x+runewidth > m.Width {
		return
	}
	if y < 0 || y >= m.Height {
		return
	}
	// off := m.Width*y + x
	// m.Cells[off] = proto

	if x > 0 {
		if _, _, _, w := m.Screen.GetContent(x-1, y); w == 2 {
			m.Screen.SetContent(x-1, y, 0, nil, proto.Style)
		}
	}
	m.Screen.SetContent(x, y, proto.Ch, nil, proto.Style)
	if runewidth == 2 {
		m.Screen.SetContent(x+1, y, 0, nil, proto.Style)
	}
}
*/

// Resizes the Buffer, buffer contents are invalid after the resize.
func (s *Screen) Resize(w, h int) {
	s.Width = w
	s.Height = h
}

type LabelParams struct {
	Style          tcell.Style
	Align          Alignment
	Ellipsis       rune
	CenterEllipsis bool
}

// func (m *Screen) DrawLabel(dest utils.Rect, params *LabelParams, text []byte) {
func (s *Screen) DrawLabel(dest utils.Rect, params *LabelParams, text string) {
	// gelog.Info("DrawLabel, text = '%s', param='%#v'. dest='%#v'", string(text), params, dest)

	/*
		// 高さを 1 にする
		if dest.Height != 1 {
			dest.Height = 1
		}

		// m と dest の重なった領域を取得
		dest = m.Rect.Intersection(dest)
		if dest.Height == 0 || dest.Width == 0 {
			return
		}
	*/

	ellipsisWidth := utils.RuneWidth(params.Ellipsis)
	runs := []rune(text)

	ellipsisFlag := false
	var leftWidth, rightWidth int
	var leftStr, rightStr string
	for leftIndex, rightIndex := 0, len(runs)-1; leftIndex <= rightIndex; {
		if (params.Align == AlignRight || params.CenterEllipsis) && leftWidth > rightWidth {
			ch := runs[rightIndex]
			w := utils.RuneWidth(ch)
			if leftWidth+rightWidth+w > dest.Width {
				ellipsisFlag = true
				break
			}
			rightWidth += w
			rightStr = string(ch) + rightStr
			rightIndex--
		} else {
			ch := runs[leftIndex]
			w := utils.RuneWidth(ch)
			if leftWidth+rightWidth+w > dest.Width {
				ellipsisFlag = true
				break
			}
			leftWidth += w
			leftStr += string(ch)
			leftIndex++
		}
	}
	if ellipsisFlag {
		// Delete letters to make room for ellipsis
		for leftWidth+rightWidth+ellipsisWidth > dest.Width {
			if leftWidth > rightWidth {
				leftWidth -= utils.RuneWidth(rune(leftStr[len(leftStr)-1]))
				leftStr = leftStr[:len(leftStr)-1]
			} else {
				rightWidth -= utils.RuneWidth(rune(rightStr[0]))
				rightStr = rightStr[1:]
			}
		}
	}

	max := dest.Width - leftWidth - rightWidth // for space
	str := ""
	if ellipsisFlag {
		max -= ellipsisWidth
		if params.CenterEllipsis {
			str = leftStr + string(params.Ellipsis) + rightStr
		} else {
			str = leftStr + rightStr + string(params.Ellipsis)
		}
	} else {
		str = leftStr + rightStr
	}
	if params.Align == AlignCenter {
		max1 := max / 2
		for i := 0; i < max1; i++ {
			str = " " + str
		}
		for i := 0; i < max-max1; i++ {
			str += " "
		}
	} else {
		space := ""
		for i := 0; i < max; i++ {
			space += " "
		}
		if params.Align == AlignLeft {
			str += space
		} else {
			str = space + str
		}
	}
	s.DrawString(dest.X, dest.Y, dest.Width, str, params.Style)
}

// fill space if width > 0
func (s *Screen) DrawString(x, y, width int, str string, style tcell.Style) {
	// if _, _, _, w := s.GetContent(x-1, y); w == 2 { // tcell/v2
	if _, _, w := s.Get(x-1, y); w == 2 { // tcell/v3
		s.SetContent(x-1, y, 0, nil, theme.ColorDefault)
	}

	width += x
	for _, ch := range str {
		w := utils.RuneWidth(ch)
		if width > 0 && x+w > width {
			break
		}
		s.SetContent(x, y, ch, nil, style)
		for i := 1; i < w; i++ {
			s.SetContent(x+i, y, 0, nil, style)
		}
		x += w
	}

	// fill space if width > 0
	for ; x < width; x++ {
		s.SetContent(x, y, ' ', nil, style)
	}

}
