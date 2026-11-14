// gecore/echo.go
package gecore

import (
	"context"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

var Echo *EchoStruct = NewEcho()

// The echo line has two parts.
//
//   - text (SetText / AddText): a message of the moment. It is shown by the
//     next Draw and then cleared.
//   - usage (PushUsage / SetUsage): the state and the keys of a mode. It is
//     kept until ClearUsage, so it stays while keys are typed.
//
// Draw shows the usage first, then the text:
//
//	Space:set/unset  Enter:finish  C-g:quit | saved, 3 cursors
//
// When the line is too narrow the text has the priority: the usage loses
// its last items (a "..." marks that), and is left out when even the first
// item does not fit.
//
// Required implement Overlay interface
type EchoStruct struct {
	textArray   []echoText
	usage       []echoText // kept by Draw, cleared by ClearUsage
	overlayRect screen.Rect
	height      int
	chain       *chainStruct
}

type echoText struct {
	text  string
	style *tcell.Style
}

func NewEcho() *EchoStruct {
	e := &EchoStruct{
		textArray: []echoText{},
		usage:     []echoText{},
		height:    1,
	}
	e.chain = &chainStruct{echo: e}
	return e
}

// chainStruct is returned by the functions that add text, so that
// WithStyle styles what the last call added. usage tells which of the two
// lists the last call added to.
type chainStruct struct {
	// isShowCursor  bool
	echo       *EchoStruct
	styleStart int
	styleEnd   int
	usage      bool
}

/* func (c *chainStruct) ShowCursor(b bool) *chainStruct {
	c.isShowCursor = b
	return c
}
*/

func (c *chainStruct) WithStyle(f tcell.Style) *chainStruct {
	if c.echo != nil {
		list := c.echo.textArray
		if c.usage {
			list = c.echo.usage
		}
		for i := c.styleStart; i < c.styleEnd && i < len(list); i++ {
			list[i].style = &f
		}
	}
	return c
}

// Overlay interface
func (e *EchoStruct) RequiredHeight() int {
	return e.height
}

// Overlay interface
func (e *EchoStruct) IsActive() bool {
	// The minibuffer and echo line are mutually exclusive and are never displayed simultaneously.
	// return len(e.textArray) > 0

	// Always displayed
	return true
}

func (e *EchoStruct) Resize(overlayRect screen.Rect) {
	// 計算は overlay.Layout が行う
	e.overlayRect = overlayRect
}

func (e *EchoStruct) Draw(ctx context.Context) bool {
	scr := screen.Get()

	for x := 0; x < e.overlayRect.Width; x++ {
		scr.SetContent(x, e.overlayRect.Y, ' ', nil, theme.ColorEchoLine)
	}

	for _, c := range layoutEcho(e.overlayRect.Width, e.usage, e.textArray) {
		scr.DrawString(c.x, e.overlayRect.Y, c.width, c.text, *e.styleFor(c.style))
	}

	/*
		if e.chain.isShowCursor {
			w := utils.WidthOnScreen([]byte(s))
			screen.ShowCursor(w, e.overlayRect.Y+e.overlayRect.Height-1)
			e.chain.isShowCursor = false
		}
	*/

	// Clear text. The usage is kept.
	e.textArray = e.textArray[:0]
	e.chain.styleStart = 0
	e.chain.styleEnd = 0

	return false
}

// Clear clears the text. The usage is kept (see ClearUsage).
func (e *EchoStruct) Clear() {
	e.textArray = e.textArray[:0]
	e.chain.styleStart = 0
	e.chain.styleEnd = 0
	// e.chain.isShowCursor = false
}

func (e *EchoStruct) UniversalCancel() {}

// テキストを更新

func (e *EchoStruct) SetText(s string) *chainStruct {
	e.textArray = e.textArray[:0]
	e.chain.usage = false
	e.chain.styleStart = 0
	e.chain.styleEnd = 0
	if s != "" {
		e.textArray = append(e.textArray, echoText{text: s})
		e.chain.styleEnd = len(e.textArray)
	}
	return e.chain
}

func (e *EchoStruct) AddText(s ...string) *chainStruct {
	e.chain.usage = false
	e.chain.styleStart = len(e.textArray)
	e.chain.styleEnd = e.chain.styleStart
	for _, str := range s {
		if str == "" {
			continue
		}

		if len(e.textArray) == 0 || e.textArray[len(e.textArray)-1].text != str {
			e.textArray = append(e.textArray, echoText{text: str})
			e.chain.styleEnd = len(e.textArray)
		}
	}
	return e.chain
}

func (e *EchoStruct) styleFor(style *tcell.Style) *tcell.Style {
	if style == nil {
		return &theme.ColorEchoLine
	}
	return style
}

// ------------------------------
// usage: the state and the keys of a mode, shown until ClearUsage
// ------------------------------

// PushUsage adds items to the usage. WithStyle of the result styles them.
//
// A mode sets its usage in WillEnter and clears it in WillExit: mode.Manager
// calls WillExit of the mode that is left before WillEnter of the next one,
// and WillEnter again of the one that is returned to.
func (e *EchoStruct) PushUsage(s ...string) *chainStruct {
	e.chain.usage = true
	e.chain.styleStart = len(e.usage)
	e.chain.styleEnd = e.chain.styleStart
	for _, str := range s {
		if str == "" {
			continue
		}

		if len(e.usage) == 0 || e.usage[len(e.usage)-1].text != str {
			e.usage = append(e.usage, echoText{text: str})
			e.chain.styleEnd = len(e.usage)
		}
	}
	return e.chain
}

// SetUsage replaces the usage.
func (e *EchoStruct) SetUsage(s ...string) *chainStruct {
	e.usage = e.usage[:0]
	return e.PushUsage(s...)
}

// Usage returns the items of the usage (for tests and for code that needs to
// know what is shown).
func (e *EchoStruct) Usage() []string {
	r := make([]string, len(e.usage))
	for i, u := range e.usage {
		r[i] = u.text
	}
	return r
}

func (e *EchoStruct) ClearUsage() {
	e.usage = e.usage[:0]
	if e.chain.usage {
		e.chain.styleStart = 0
		e.chain.styleEnd = 0
	}
}

// ------------------------------
// layout
// ------------------------------

const (
	echoUsageSep = "  "  // between the items of the usage
	echoBlockSep = " | " // between the usage and the text
	echoTextSep  = ", "  // between the items of the text
	echoMore     = "..." // the usage lost its last items
)

// echoCell is a piece of the echo line to draw at x. style nil is the
// default style of the echo line.
type echoCell struct {
	x     int
	width int
	text  string
	style *tcell.Style
}

func echoWidth(s string) int { return utils.WidthOnScreen([]byte(s)) }

// layoutEcho places the usage and the text on a line of width cells. It
// does not draw, so that the rules can be tested.
func layoutEcho(width int, usage, text []echoText) []echoCell {
	if width <= 0 {
		return nil
	}

	// The text is shown whole if it can be: the usage gets what is left.
	usageLimit := width
	if len(text) > 0 {
		textW := 0
		for i, t := range text {
			if i > 0 {
				textW += echoWidth(echoTextSep)
			}
			textW += echoWidth(t.text)
		}
		usageLimit = width - textW - echoWidth(echoBlockSep)
	}

	var cells []echoCell
	x := 0
	if len(usage) > 0 && usageLimit > 0 {
		cells = placeUsage(usage, usageLimit)
		if n := len(cells); n > 0 {
			x = cells[n-1].x + cells[n-1].width
		}
	}
	if x > 0 && len(text) > 0 {
		w := echoWidth(echoBlockSep)
		if x+w <= width {
			cells = append(cells, echoCell{x: x, width: w, text: echoBlockSep})
			x += w
		}
	}

	for i, t := range text {
		if i > 0 {
			w := echoWidth(echoTextSep)
			if x+w > width {
				break
			}
			cells = append(cells, echoCell{x: x, width: w, text: echoTextSep})
			x += w
		}

		w := echoWidth(t.text)
		if remaining := width - x; w > remaining {
			w = remaining
		}
		if w <= 0 {
			break
		}
		cells = append(cells, echoCell{x: x, width: w, text: t.text, style: t.style})
		x += w
		if x >= width {
			break
		}
	}
	return cells
}

// placeUsage places the items of the usage from x = 0 within limit cells.
// The items that do not fit are left out, and "..." takes their place when
// it fits. Nothing is placed when not even the first item fits.
func placeUsage(usage []echoText, limit int) []echoCell {
	type end struct{ n, x int } // number of cells and width after an item
	var (
		cells    []echoCell
		ends     []end
		x        int
		complete = true
	)
	sepW := echoWidth(echoUsageSep)

	for _, u := range usage {
		w := echoWidth(u.text)
		if w == 0 {
			continue
		}
		start := x
		if len(ends) > 0 {
			start += sepW
		}
		if start+w > limit {
			complete = false
			break
		}
		if len(ends) > 0 {
			cells = append(cells, echoCell{x: x, width: sepW, text: echoUsageSep})
		}
		cells = append(cells, echoCell{x: start, width: w, text: u.text, style: u.style})
		x = start + w
		ends = append(ends, end{n: len(cells), x: x})
	}

	if !complete {
		moreW := echoWidth(echoMore)
		// Make room for the sign by dropping the items from the end.
		for len(ends) > 0 && x+sepW+moreW > limit {
			ends = ends[:len(ends)-1]
			if len(ends) > 0 {
				last := ends[len(ends)-1]
				cells, x = cells[:last.n], last.x
			} else {
				cells, x = nil, 0
			}
		}
		if len(ends) > 0 {
			cells = append(cells,
				echoCell{x: x, width: sepW, text: echoUsageSep},
				echoCell{x: x + sepW, width: moreW, text: echoMore})
		}
	}
	return cells
}
