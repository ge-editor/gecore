package gecore

import (
	"strings"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/keychord"
)

var (
	CancelKeyOnly = keychord.NewRootNode()
	GlobalKey     = keychord.NewRootNode()
	RootKey       = keychord.NewRootNode()

	ModeManager = mode.NewManager(RootKey)

	// Keyboard macro mode
	MacroKey         *keychord.RootNode = keychord.NewRootNode()
	MacroModeManager *mode.Manager      = mode.NewManager(MacroKey)
	MacroMode        *MacroModeStruct
)

// Must be modified to execute events for each view compared to all registered view types
func eventActiveTreeLeaf(ev tcell.EventKey) (tcell.EventKey, error) {
	// gelog.Info("tree.ActiveLeaf", tree.ActiveLeaf)
	/* 	if tree.ActiveLeaf == nil {
	   		return ev, nil
	   	}
	*/
	leaf := tree.ActiveTreeGet().GetLeaf()
	_, res := leaf.DispatchKey(ev)

	// _, res := (*tree.ActiveLeaf).DispatchKey(ev)
	switch res {
	case keychord.DispatchNotFound:
	case keychord.DispatchPrefix:
	case keychord.DispatchExecuted:
	}
	/* 	if res.Consumed {
	   		// Leaf が処理したので、イベントは止める
	   		return ev, nil
	   	}
	*/
	// 未消費 → 上位に流す
	return ev, nil
}

var prefix []string // echo 表示用

// 2026-01-18 Sun 14:20:51
// leaf で ESC > が動作しなくなったが
// 下記で解消している
func Dispatch(ev tcell.EventKey) {
	/* 	gelog.Key("EventKey",
	   		"Str", ev.Str(),
	   		"Key", ev.Key(),
	   		"Name", ev.Name(),
	   		"Ctrl", ev.Modifiers()&tcell.ModCtrl,
	   		"Shift", ev.Modifiers()&tcell.ModShift,
	   		"Alt", ev.Modifiers()&tcell.ModAlt,
	   		"Meta", ev.Modifiers()&tcell.ModMeta,
	   	)
	*/

	if _, res := CancelKeyOnly.Dispatch(ev); res == keychord.DispatchExecuted {
		keychord.ResetAllRootNodes()
		return
	}

	// Keyboard macro
	macroActive := MacroModeManager.ActiveKeys()
	keyStatus, res := macroActive.Dispatch(ev)
	switch res {
	case keychord.DispatchExecuted:
		keychord.ResetAllRootNodes()
		return
	}
	MacroMode.Append(ev)

	// Global Root を最優先
	keyStatus, gres := GlobalKey.Dispatch(ev)
	switch gres {
	case keychord.DispatchExecuted:
		// リセットするべきか？
		keychord.ResetAllRootNodes()
		// rootKey.Reset()
		ModeManager.ActiveKeys().Reset()
		prefix = []string{}
		return
	case keychord.DispatchPrefix:
		prefix = append(prefix, keyStatus)
	case keychord.DispatchNotFound:
		prefix = []string{}
	}

	// gelog.Info("minibuff", "Dispatch")
	minibuff := MinibufferManager()
	// minibuffer session に dispatch される
	res = minibuff.Dispatch(ev)
	if res == keychord.DispatchExecuted {
		overlay.OverlayManager().Layout(screen.Get().Rect) // Resize minibuffer height
		return
	}
	if minibuff.IsActive() {
		return // Consume in the minibuffer
	}

	// モード / 通常 Root を処理
	active := ModeManager.ActiveKeys()
	// gelog.Info("modeManager", "ActiveKeys", "Dispatch")
	keyStatus, res = active.Dispatch(ev)
	switch res {
	case keychord.DispatchExecuted:
		// グローバルキーの状態をリセットしておく
		// globalKey.Reset()
		keychord.ResetAllRootNodes()
		return
	case keychord.DispatchPrefix:
		// プレフィックスキー 状態表示
		if len(prefix) == 0 || (len(prefix) > 0 && prefix[len(prefix)-1] != keyStatus) {
			prefix = append(prefix, keyStatus)
		}
		// printLine(7, fmt.Sprintf("Prefix key: %q, waiting next", strings.Join(prefix, ",")))
		Echo.AddText(strings.Join(prefix, ","))
		// return
	case keychord.DispatchNotFound:
		if gres == keychord.DispatchNotFound {
			// 無効キー リセット
			// keychord.ResetAllRootNodes()
			// globalKey.Reset()
			// rootKey.Reset()
			// active.Reset()
			prefix = []string{}
		}
	}

	// gelog.Info("eventActiveTreeLeaf")
	ev, _ = eventActiveTreeLeaf(ev) // to
	/*
		 	if err == gecore.ErrCodeAction {
				gecore.Echo.AddText("")
			}
	*/
}
