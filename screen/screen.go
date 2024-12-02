package screen

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/theme"
)

var screen *Screen

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

func Get() *Screen {
	return screen
}

// Return Screen Rect without Minibuffer
func (s *Screen) RootRect() utils.Rect {
	return utils.Rect{X: 0, Y: 0, Width: s.Width, Height: s.Height - 1}
}

// ShowCursor sets the cursor position to (x, y).
func (s *Screen) ShowCursor(x, y int) {
	s.CX, s.CY = x, y
	// _, file, line, _ := runtime.Caller(1)
	// verb.PP("Caller: %s: %d", file, line)
	// verb.PP("ShowCursor %d,%d", m.CX, m.CY)
	s.Screen.ShowCursor(x, y)
}

func (s *Screen) HideCursor() {
	s.CX, s.CY = -s.CX, -s.CY
	// verb.PP("HideCursor %d,%d", m.CX, m.CY)
	s.Screen.ShowCursor(-1, -1)
}

// Fills an area which is an intersection between buffer and 'dest' with 'proto'.
func (s *Screen) Fill(dest utils.Rect, proto Cell) {
	// m.unsafe_fill(m.Rect.Intersection(dst), proto)
	// Unsafe part of the fill operation, doesn't check for bounds.
	// func (m *Screen) unsafe_fill(dest utils.Rect, proto Cell) {
	dest = s.Rect.Intersection(dest)
	// verb.PP("unsafe fill proto='%#v', dest='%#v'", proto, dest)
	runeWidth := utils.RuneWidth(proto.Ch)
	// stride := m.Width
	// off := m.Width*dest.Y + dest.X
	// 1個前の cell の文字幅が 2なら cell を空ににする
	/*
		if _, _, _, w := m.Screen.GetContent(dest.X-1, dest.Y); w == 2 {
			m.Screen.SetContent(dest.X-1, dest.Y, 0, nil, proto.Style)
		}
	*/
	for y := 0; y < dest.Height; y++ {
		for x := 0; x < dest.Width; x++ {
			if x == 0 {
				// 1個前の cell の文字幅が 2なら cell を空ににする
				if _, _, _, w := s.Screen.GetContent(dest.X-1, dest.Y); w == 2 {
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
			/*
				if dest.X+x+runeWidth <= dest.Width {
					m.Screen.SetContent(dest.X+x, dest.Y+y, proto.Ch, nil, proto.Style)
					if runeWidth == 2 {
						x++
						m.Screen.SetContent(dest.X+x, dest.Y+y, 0, nil, proto.Style)
					}
				} else {
					m.Screen.SetContent(dest.X+x, dest.Y+y, 0, nil, proto.Style)
				}
			*/
		}
		// off += stride
	}
	// }
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
	// verb.PP("DrawLabel, text = '%s', param='%#v'. dest='%#v'", string(text), params, dest)

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

// Append message to Screen.echo []string
func (s *Screen) Echo(message string) {
	// Exclude same message
	if len(s.echo) > 0 {
		if s.echo[len(s.echo)-1] == message {
			return
		}
	}
	s.echo = append(s.echo, message)
}

// 引数が指定されている場合には
// 指定された文字列を screen.echo に追加して内容を即表示する
// 引数が指定されていない場合には
// screen.echo の内容を表示する
func (s *Screen) PrintEcho(str ...string) {
	if len(str) > 0 {
		s.echo = append(s.echo, str...)
	} else if len(s.echo) == 0 {
		return
	}

	// Join echo messages
	// message := strings.Join(e.echo, ", ")
	message := ""
	for _, str := range s.echo {
		if str == "" {
			continue
		}
		if message != "" {
			message += ", "
		}
		message += str
	}
	s.DrawString(0, s.Rect.Height-1, s.Rect.Width, message, theme.ColorDefault)
	s.echo = []string{}
}

// fill space if width > 0
func (s *Screen) DrawString(x, y, width int, str string, style tcell.Style) {
	if _, _, _, w := s.GetContent(x-1, y); w == 2 {
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

// x, y int  : 描画領域の始点
// width int : 領域の幅
// maxThreshold int : カーソル位置左右
// cells []Cell, index int : 表示内容, 表示開始位置
// clearStyle tcell.Style
// prefixBytes []byte, prefixWidth int, prefixStyle tcell.Style: prefix 文字と幅、スタイル
// echo bool : 表示のみ
func (s *Screen) DrawMiniBuffer(x, y, width, maxThreshold int, cells []Cell, index int, clearStyle tcell.Style, prefix string, prefixWidth int, prefixStyle tcell.Style, echo bool) (drawStartIndex, drawEndIndex int) {
	threshold := utils.Threshold(maxThreshold, width)

	// index を元に描画を開始する cells[drawStartIndex] を計算する
	drawStartIndex = index
	for w := 0; !echo && drawStartIndex > 0; drawStartIndex-- {
		// for w := 0; drawStartIndex > 0; drawStartIndex-- {
		if width-prefixWidth-threshold < w+cells[drawStartIndex].Width {
			break
		}
		w += cells[drawStartIndex].Width
	}

	// draw prefix
	//for i := 0; i < len(prefixBytes); {
	for _, ch := range prefix {
		//ch, size := utf8.DecodeRune(prefixBytes[i:])
		w := utils.RuneWidth(ch)
		//i += size

		if x+w > width {
			break
		}
		s.SetContent(x, y, ch, nil, prefixStyle)
		for i := 1; i < w; i++ {
			s.SetContent(x+i, y, 0, nil, prefixStyle)
		}
		x += w
	}

	// draw buffer and calcurate cursor x position
	cx := 0
	for i := drawStartIndex; i < len(cells); i++ {
		c := cells[i]
		if x+c.Width > width {
			break
		}
		drawEndIndex = i
		s.SetContent(x, y, c.Ch, nil, c.Style)
		for i := 1; i < c.Width; i++ {
			s.SetContent(x+i, y, 0, nil, c.Style)
		}
		if i == index {
			cx = x
		}
		x += c.Width
	}

	// fill space
	for ; x < width; x++ {
		s.SetContent(x, y, ' ', nil, clearStyle)
	}

	if echo {
		s.HideCursor()
	} else {
		s.ShowCursor(cx, y)
	}
	return
}
