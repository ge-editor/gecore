// gecore/modes/redo_mode.go
// 最もシンプルなモード
package modes

import (
	"context"

	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

type RedoMode struct {
	ModeManager *mode.Manager
	RootNode    *keychord.RootNode
}

func NewRedoMode(mm *mode.Manager, key func(*keychord.RootNode, *RedoMode)) *RedoMode {
	m := &RedoMode{
		ModeManager: mm,
		RootNode:    keychord.NewRootNode(),
	}

	if key != nil {
		key(m.RootNode, m)
	}

	return m
}

// OverMinibuffer を実装したモードは、minibuffer が active でもキーを先に受け取る。
func (m *RedoMode) OverMinibuffer() bool { return true }

// --- Mode interface ---
func (m *RedoMode) Name() string             { return "RedoMode" }
func (m *RedoMode) Keys() *keychord.RootNode { return m.RootNode }
func (m *RedoMode) WillEnter()               {}
func (m *RedoMode) WillExit()                {}
func (m *RedoMode) Draw(ctx context.Context) {}

// --- Mode pointer ---
func (m *RedoMode) CurrentMode() mode.Mode { return m }
