package gecore

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

type MacroModeStruct struct {
	ModeManager *mode.Manager
	RootNode    *keychord.RootNode

	keymacros []tcell.EventKey
	recording bool
	replay    bool
	index     int // 再生位置
	dispatch  func(tcell.EventKey)
}

func NewMacroMode(mm *mode.Manager, key func(*keychord.RootNode, *MacroModeStruct), dispatch func(tcell.EventKey)) *MacroModeStruct {
	m := &MacroModeStruct{
		ModeManager: mm,
		RootNode:    keychord.NewRootNode(),
		keymacros:   make([]tcell.EventKey, 0, 32),
		dispatch:    dispatch,
	}

	if key != nil {
		key(m.RootNode, m)
	}

	return m
}

// --- Mode interface ---
func (m *MacroModeStruct) Name() string             { return "MacroMode" }
func (m *MacroModeStruct) Keys() *keychord.RootNode { return m.RootNode }
func (m *MacroModeStruct) WillEnter()               {}
func (m *MacroModeStruct) WillExit()                {}
func (m *MacroModeStruct) Draw()                    {}

// --- Macro functions ---
func (m *MacroModeStruct) StartRecording() {
	m.keymacros = m.keymacros[:0]
	m.recording = true
	m.replay = false
}

func (m *MacroModeStruct) StopRecording() {
	m.recording = false
	m.RootNode.Reset()
}

func (m *MacroModeStruct) AbortRecording() {
	if !m.recording {
		return
	}
	m.recording = false
}

// --- Mode pointer ---
func (m *MacroModeStruct) CurrentMode() mode.Mode { return m }

func (m *MacroModeStruct) Append(ev tcell.EventKey) {
	if !m.recording {
		return
	}
	// copied := ev
	m.keymacros = append(m.keymacros, ev)
}

func (m *MacroModeStruct) Replay() {
	if m.replay {
		return
	}

	m.replay = true
	for _, ev := range m.keymacros {
		m.dispatch(ev)
	}
	m.replay = false
	Echo.AddText("(Type e to repeat macro)")
}
