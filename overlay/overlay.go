// gecore/overlay/overlay.go

package overlay

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
)

type Overlay interface {
	RequiredHeight() int
	Resize(overlayRect screen.Rect)
	Draw() bool
	IsActive() bool
}

var overlayManager = overlayManagerStruct{}

type overlayManagerStruct struct {
	Tree         Overlay
	Minibuffer   Overlay
	Echo         Overlay
	freeOverlays []Overlay
}

func OverlayManager() *overlayManagerStruct {
	return &overlayManager
}

// Add registers an overlay.
// Order matters: later-added overlays are stacked lower (drawn later).
func (m *overlayManagerStruct) Add(o Overlay) {
	if o == nil {
		return
	}

	m.freeOverlays = append(m.freeOverlays, o)
}

// Remove unregisters an overlay.
func (m *overlayManagerStruct) Remove(o Overlay) {
	if o == nil {
		return
	}
	for i, v := range m.freeOverlays {
		if v == o {
			m.freeOverlays = append(m.freeOverlays[:i], m.freeOverlays[i+1:]...)
			return
		}
	}
}

func (m *overlayManagerStruct) SetTree(o Overlay) {
	if o == nil {
		return
	}

	m.Tree = o
}

func (m *overlayManagerStruct) SetMinibuffer(o Overlay) {
	if o == nil {
		return
	}

	m.Minibuffer = o
}

func (m *overlayManagerStruct) SetEcho(o Overlay) {
	if o == nil {
		return
	}

	m.Echo = o
}

// Layout calculates overlay rectangles from the bottom of screenRect upward,
// calls Resize on each overlay, and returns the remaining rect for tree.
// stack from bottom (last overlay is bottom-most)
func (m *overlayManagerStruct) Layout(screenRect screen.Rect) screen.Rect {
	rect := screenRect
	act := m.Minibuffer.IsActive()

	// Echo
	if m.Echo != nil {
		// o.Resize() must be called even when overlay.Height() is 0.

		// Echo is normally inactive and displayed exclusively with the minibuffer.
		// The minibuffer and echo line are mutually exclusive and are never displayed simultaneously.
		// When Echo is active, it is displayed unconditionally.
		h := 0
		if m.Echo.IsActive() || !act {
			h = m.Echo.RequiredHeight()
		}

		if h > rect.Height {
			h = rect.Height
		}

		y := rect.Y + rect.Height - h

		overlayRect := screen.Rect{
			X:      rect.X,
			Y:      y,
			Width:  rect.Width,
			Height: h,
		}

		m.Echo.Resize(overlayRect)

		rect.Height -= h // Remaining height

		if rect.Height <= 0 {
			rect.Height = 0
			return rect
		}
	}

	// Minibuffer
	if m.Minibuffer != nil {
		// overlay.Height() が 0 の場合でも o.Resize() を呼び出す必要がある
		// Echo は通常 inactive で minibuffer と排他表示, active な場合は強制表示
		h := m.Minibuffer.RequiredHeight()
		if h > rect.Height {
			h = rect.Height
		}

		y := rect.Y + rect.Height - h

		useInCalcMinibufferMaxHeight := h
		if act {
			useInCalcMinibufferMaxHeight = screen.Get().Height // 画面全体の高さ
			// Minibuffer の高さの最大値は画面全体の高さから割合で算出している
		}

		overlayRect := screen.Rect{
			X:      rect.X,
			Y:      y,
			Width:  rect.Width,
			Height: useInCalcMinibufferMaxHeight,
		}

		m.Minibuffer.Resize(overlayRect)

		rect.Height -= h // Remaining height

		if rect.Height <= 0 {
			rect.Height = 0
			return rect
		}
	}

	// Tree
	if m.Tree != nil {
		/*
			h := m.Tree.RequiredHeight()
			if h > rect.Height {
				h = rect.Height
			}
		*/

		// 残り全ての高さを使用
		h := rect.Height

		overlayRect := screen.Rect{
			X:      rect.X,
			Y:      0,
			Width:  rect.Width,
			Height: h,
		}

		m.Tree.Resize(overlayRect)

		// Free overlays
		for _, ov := range m.freeOverlays {
			ov.Resize(overlayRect) // Tree のサイズを渡す
		}
		return rect
	}

	return rect
}

// Draw draws all overlays in registration order.
func (m *overlayManagerStruct) Draw() bool {
	flowOverlays := []Overlay{m.Tree, m.Minibuffer, m.Echo}
	for _, o := range flowOverlays {
		if o == nil {
			continue
		}
		if o.Draw() {
			return true
		}
	}
	for _, o := range m.freeOverlays {
		if o == nil {
			continue
		}
		if o.Draw() {
			return true
		}
	}
	return false
}

func (m *overlayManagerStruct) Resize(ev tcell.EventResize) {
	w, h := ev.Size()
	if w <= 0 || h <= 0 {
		return
	}

	screenRect := screen.Rect{
		X:      0,
		Y:      0,
		Width:  w,
		Height: h,
	}

	// Layout が各 overlay の Resize を呼ぶ
	m.Layout(screenRect)
}
