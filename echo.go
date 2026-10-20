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

// Required implement Overlay interface
type EchoStruct struct {
	textArray   []echoText
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
		height:    1,
	}
	e.chain = &chainStruct{echo: e}
	return e
}

type chainStruct struct {
	// isShowCursor  bool
	echo       *EchoStruct
	styleStart int
	styleEnd   int
}

/* func (c *chainStruct) ShowCursor(b bool) *chainStruct {
	c.isShowCursor = b
	return c
}
*/

func (c *chainStruct) WithStyle(f tcell.Style) *chainStruct {
	if c.echo != nil {
		for i := c.styleStart; i < c.styleEnd && i < len(c.echo.textArray); i++ {
			c.echo.textArray[i].style = &f
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
	screen := screen.Get()

	for x := 0; x < e.overlayRect.Width; x++ {
		screen.SetContent(x, e.overlayRect.Y, ' ', nil, theme.ColorEchoLine)
	}

	x := 0
	for i, item := range e.textArray {
		if i > 0 {
			if x+2 > e.overlayRect.Width {
				break
			}
			screen.DrawString(x, e.overlayRect.Y, 2, ", ", theme.ColorEchoLine)
			x += 2
		}

		style := e.styleFor(item.style)
		width := utils.WidthOnScreen([]byte(item.text))
		if remaining := e.overlayRect.Width - x; width > remaining {
			width = remaining
		}
		if width <= 0 {
			break
		}
		screen.DrawString(x, e.overlayRect.Y, width, item.text, *style)
		x += width
		if x >= e.overlayRect.Width {
			break
		}
	}

	/*
		if e.chain.isShowCursor {
			w := utils.WidthOnScreen([]byte(s))
			screen.ShowCursor(w, e.overlayRect.Y+e.overlayRect.Height-1)
			e.chain.isShowCursor = false
		}
	*/

	// Clear text
	e.textArray = e.textArray[:0]
	e.chain.styleStart = 0
	e.chain.styleEnd = 0

	return false
}

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
	e.chain.styleStart = 0
	e.chain.styleEnd = 0
	if s != "" {
		e.textArray = append(e.textArray, echoText{text: s})
		e.chain.styleEnd = len(e.textArray)
	}
	return e.chain
}

func (e *EchoStruct) AddText(s ...string) *chainStruct {
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
