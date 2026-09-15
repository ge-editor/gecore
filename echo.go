// gecore/echo.go
package gecore

import (
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/theme"
	"github.com/ge-editor/utils"
)

var Echo *EchoStruct = NewEcho()

type activeFlag int

const (
	EchoNormal activeFlag = iota
	EchoBlue
	EchoGreen
	EchoYellow
	EchoRed
)

// Required implement Overlay interface
type EchoStruct struct {
	textArray   []string
	overlayRect utils.Rect
	style       tcell.Style
	height      int
	chain       *chainStruct
}

func NewEcho() *EchoStruct {
	return &EchoStruct{
		textArray: []string{},
		style:     theme.ColorEchoLine,
		height:    1,
		chain:     &chainStruct{},
	}
}

type chainStruct struct {
	isShowCursor  bool
	justNowActive activeFlag
}

func (c *chainStruct) ShowCursor(b bool) *chainStruct {
	c.isShowCursor = b
	return c
}

func (c *chainStruct) JustNowActive(f activeFlag) *chainStruct {
	c.justNowActive = f
	return c
}

// Overlay interface
func (e *EchoStruct) RequiredHeight() int {
	return e.height
}

// Overlay interface
func (e *EchoStruct) IsActive() bool {
	// The minibuffer and echo line are mutually exclusive and are never displayed simultaneously.
	// return e.chain.justNowActive != EchoNormal

	// Always displayed
	return true
}

func (e *EchoStruct) Resize(overlayRect utils.Rect) {
	// 計算は overlay.Layout が行う
	e.overlayRect = overlayRect
}

func (e *EchoStruct) Draw(ts tcell.Screen) bool {
	screen := screen.Get()

	for x := 0; x < e.overlayRect.Width; x++ {
		screen.SetContent(x, e.overlayRect.Y, ' ', nil, e.style)
	}

	s := strings.Join(e.textArray, ", ")
	screen.DrawString(0, e.overlayRect.Y, e.overlayRect.Width, s, e.style)

	if e.chain.isShowCursor {
		w := utils.WidthOnScreen([]byte(s))
		screen.ShowCursor(w, e.overlayRect.Y+e.overlayRect.Height-1)
		e.chain.isShowCursor = false
	}

	// Clear text
	e.textArray = e.textArray[:0]
	e.chain.justNowActive = EchoNormal

	return false
}

func (e *EchoStruct) Clear() {
	e.textArray = e.textArray[:0]
	e.chain.justNowActive = EchoNormal
	e.chain.isShowCursor = false
}

func (e *EchoStruct) UniversalCancel() {}

// テキストを更新

func (e *EchoStruct) SetText(s string) *chainStruct {
	e.textArray = e.textArray[:0]
	if s != "" {
		e.textArray = append(e.textArray, s)
	}
	return e.chain
}

func (e *EchoStruct) AddText(s ...string) *chainStruct {
	for _, str := range s {
		if str == "" {
			continue
		}

		if len(e.textArray) == 0 || e.textArray[len(e.textArray)-1] != str {
			e.textArray = append(e.textArray, str)
		}
	}
	return e.chain
}
