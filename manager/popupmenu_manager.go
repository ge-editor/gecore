// gecore/manager/popupmenu_manager.go

package manager

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/popupmenu"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/keychord"
)

// var popupmenuManager = &PopupmenuManagerStruct{}

/* func PopupmenuManager() *PopupmenuManagerStruct {
	return popupmenuManager
}
*/

func NewPopupmenuManager() *PopupmenuManagerStruct {
	return &PopupmenuManagerStruct{}
}

// Required implement Overlay interface
type PopupmenuManagerStruct struct {
	active bool
	// dispatchActive bool
	session     *popupmenu.Session
	overlayRect screen.Rect
}

func (pm *PopupmenuManagerStruct) Popupmenu() *popupmenu.PopupmenuStruct {
	if pm.session == nil {
		return nil
	}
	return pm.session.Popupmenu
}

func (pm *PopupmenuManagerStruct) Start(session *popupmenu.Session) {
	pm.session = session
	pm.active = true
	// pm.dispatchActive = true

	overlay.OverlayManager().Add(pm)
	gecore.CancelManager().Register(pm)

	pm.Resize(pm.overlayRect)
	overlay.OverlayManager().Layout(screen.Get().Rect)
}

func (pm *PopupmenuManagerStruct) Close() {
	if !pm.active {
		return
	}

	pm.session = nil
	pm.active = false
	//pm.dispatchActive = false

	overlay.OverlayManager().Remove(pm)
	gecore.CancelManager().Unregister(pm)

	overlay.OverlayManager().Layout(screen.Get().Rect)
}

func (pm *PopupmenuManagerStruct) Cancel() {
	pm.Close()
}

func (pm *PopupmenuManagerStruct) Priority() int {
	return 32 // 適当
}

func (pm *PopupmenuManagerStruct) Active(b bool) {
	pm.active = b
}

func (pm *PopupmenuManagerStruct) IsActive() bool {
	return pm.active
}

/* func (pm *PopupmenuManagerStruct) DispatchActive(b bool) {
	pm.dispatchActive = b
}

func (pm *PopupmenuManagerStruct) IsDispatchActive() bool {
	return pm.dispatchActive
}
*/

/*
	 func (pm *PopupmenuManagerStruct) Type() overlay.OverlayType {
		return overlay.OverlayFree
	}
*/

func (pm *PopupmenuManagerStruct) RequiredHeight() int {
	if !pm.active || pm.session == nil {
		return 0
	}
	return pm.session.Popupmenu.RequiredHeight()
}

func (pm *PopupmenuManagerStruct) Resize(overallRect screen.Rect) {
	// gelog.Info("PopupmenuManager", "Resize", "rect", rect)
	pm.overlayRect = overallRect

	if !pm.active || pm.session == nil {
		return
	}

	pm.session.Popupmenu.Resize(overallRect)
	// overlay.OverlayManager().Layout(screen.Get().Rect) ///////////////no!
}

func (pm *PopupmenuManagerStruct) Draw() bool {
	if !pm.active || pm.session == nil {
		return false
	}
	pm.session.Popupmenu.Draw()
	return false
}

func (pm *PopupmenuManagerStruct) UniversalCancel() {
	pm.Close()
}

func (pm *PopupmenuManagerStruct) SetItems(items []string) {
	if pm.session == nil {
		return
	}
	index, _ := pm.session.Popupmenu.GetItem()
	pm.session.Popupmenu.SetItems(items, index)
	// pm.session.Popupmenu.Resize(pm.overlayRect) ///////////
}

// 選択結果取得
func (pm *PopupmenuManagerStruct) GetItem() (int, string) {
	if pm.session == nil {
		return -1, ""
	}
	return pm.session.Popupmenu.GetItem()
}

func (pm *PopupmenuManagerStruct) Dispatch(ev tcell.EventKey) keychord.KeyDispatchTransition {
	// gelog.Info("PopupmenuManager", "Dispatch", "ev", ev)
	// if !pm.active || !pm.dispatchActive {
	if !pm.active {
		return keychord.DispatchNotFound
	}

	/* 	if pm.session == nil {
	   		return keychord.DispatchNotFound
	   	}
	*/ //pm.active = false

	_, res := pm.session.Popupmenu.DispatchKey(ev)
	return res
}

/* func (pm *PopupmenuManager) Dispatch(ev tcell.EventKey) bool {
	if !pm.active {
		return false
	}
	pm.session.Popupmenu.DispatchKey(ev)
	return true
}
*/
