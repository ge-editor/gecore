package screen

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

var (
	screen     *Screen
	screenOnce sync.Once
)

// Get returns the screen, creating it (and putting the terminal into the
// full-screen mode) on the first call. It is deliberately not done in
// init(): "ge --remote ..." is a plain command line tool that imports this
// package too, and must leave the terminal alone.
// Get returns nil when the terminal cannot be initialized.
func Get() *Screen {
	screenOnce.Do(func() {
		var err error
		screen, err = newScreen()
		if err != nil {
			fmt.Println(err)
		}
		// screen.EnableMouse()
	})
	return screen
}

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

type Screen struct {
	tcell.Screen
	Rect
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
		Rect:   Rect{X: 0, Y: 0, Width: w, Height: h},
	}
	return screen, nil
}

// Return Screen Rect without Minibuffer
func (s *Screen) RootRect() Rect {
	return Rect{X: 0, Y: 0, Width: s.Width, Height: s.Height - 1}
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

// cellScreen is what FillRect needs of the screen: tcell.Screen has both.
type cellScreen interface {
	Get(x, y int) (string, tcell.Style, int)
	SetContent(x, y int, primary rune, combining []rune, style tcell.Style)
}

func (s *Screen) FillRect(dest Rect, r rune, style tcell.Style) {
	fillRect(s.Screen, s.Rect, dest, r, style)
}

// fillRect fills dest (clipped to bounds) with r in style.
//
// A wide character that is cut by an edge of dest is replaced by spaces, so
// that no half of it is left. It writes only on the row of the cut: the left
// neighbor of the first cell and the right neighbor of the last one.
func fillRect(scr cellScreen, bounds, dest Rect, r rune, style tcell.Style) {
	dest = bounds.Intersection(dest)
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
		if _, st, w := scr.Get(startX-1, y); w == 2 {
			scr.SetContent(startX-1, y, ' ', nil, st)
		}

		// --- 右端のクリーンアップ ---
		// endX から始まる幅2の文字は、右半分が endX+1 (dest の外) にある。
		// 同じ行の endX+1 を空白にして、文字が分断されるのを防ぐ。
		// (endX, y+1 ではない: 下の行は、すでに描画済みかもしれない)
		//
		// ただし endX+1 が、その文字の右半分のまま (同じスタイル) のときだけ。
		// 行番号の帯を、テキストの後で塗ることがある (drawLineNumber)。
		// そのとき endX+1 はテキスト領域の先頭の桁で、もう描画済みである。
		// ここを古い文字のスタイルの空白で上書きすると、popupmenu が全角文字
		// を帯とテキストの境目に描いていた場合に、閉じた後も反転した空白が
		// 残り、行頭の文字が消える。
		if _, st, w := scr.Get(endX, y); w == 2 {
			_, next, _ := scr.Get(endX+1, y)
			scr.SetContent(endX, y, ' ', nil, style)
			if next == st {
				scr.SetContent(endX+1, y, ' ', nil, st)
			}
		}

		// --- メインの描画ループ ---
		for x := startX; x <= endX; {
			// 残りの幅が step 未満なら、文字を置けないので空白で埋める
			if x+step-1 > endX {
				scr.SetContent(x, y, ' ', nil, style)
				x++
				continue
			}

			scr.SetContent(x, y, r, nil, style)
			x += step // 文字幅分だけ進める
		}
	}
}

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

func (s *Screen) DrawLabel(dest Rect, params *LabelParams, text string) {

	ellipsisWidth := utils.RuneWidth(params.Ellipsis)
	runs := []rune(text)

	ellipsisFlag := false
	var leftWidth, rightWidth int
	var leftRunes, rightRunes []rune

	for leftIndex, rightIndex := 0, len(runs)-1; leftIndex <= rightIndex; {
		if (params.Align == AlignRight || params.CenterEllipsis) && leftWidth > rightWidth {
			ch := runs[rightIndex]
			w := utils.RuneWidth(ch)
			if leftWidth+rightWidth+w > dest.Width {
				ellipsisFlag = true
				break
			}
			rightWidth += w
			rightRunes = append([]rune{ch}, rightRunes...)
			rightIndex--
		} else {
			ch := runs[leftIndex]
			w := utils.RuneWidth(ch)
			if leftWidth+rightWidth+w > dest.Width {
				ellipsisFlag = true
				break
			}
			leftWidth += w
			leftRunes = append(leftRunes, ch)
			leftIndex++
		}
	}

	if ellipsisFlag {
		// Delete letters to make room for ellipsis
		for leftWidth+rightWidth+ellipsisWidth > dest.Width {
			if leftWidth > rightWidth {
				if len(leftRunes) == 0 {
					break
				}
				ch := leftRunes[len(leftRunes)-1]
				leftWidth -= utils.RuneWidth(ch)
				leftRunes = leftRunes[:len(leftRunes)-1]
			} else {
				if len(rightRunes) == 0 {
					break
				}
				ch := rightRunes[0]
				rightWidth -= utils.RuneWidth(ch)
				rightRunes = rightRunes[1:]
			}
		}
	}

	max := dest.Width - leftWidth - rightWidth // for space
	var str string
	if ellipsisFlag {
		max -= ellipsisWidth
		if params.CenterEllipsis {
			str = string(leftRunes) + string(params.Ellipsis) + string(rightRunes)
		} else {
			str = string(leftRunes) + string(rightRunes) + string(params.Ellipsis)
		}
	} else {
		str = string(leftRunes) + string(rightRunes)
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
