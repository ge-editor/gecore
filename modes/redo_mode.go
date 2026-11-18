// gecore/modes/redo_mode.go
// 最もシンプルなモード
package modes

import (
	"context"

	"github.com/ge-editor/gecore"
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
func (m *RedoMode) Draw(ctx context.Context) {}

// WillEnter shows the usage in the echo line until the mode is left.
func (m *RedoMode) WillEnter() { gecore.Echo.SetUsage(redoUsage()...) }

// WillExit clears the usage.
func (m *RedoMode) WillExit() { gecore.Echo.ClearUsage() }

// redoUsage is what the echo line shows in the mode: its name and the keys
// (see the key handler in ge/keyset_editorleaf_common.go). Arrow keys and
// keys with a modifier leave the mode as well.
func redoUsage() []string {
	return []string{
		"Redo",
		"/:redo",
		"C-g:quit",
	}
}

// --- Mode pointer ---
func (m *RedoMode) CurrentMode() mode.Mode { return m }
