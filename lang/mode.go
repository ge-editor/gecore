package lang

import (
	"context"

	"github.com/gdamore/tcell/v2"

	sitter "github.com/smacker/go-tree-sitter"
)

// var Modes *ModesStruct

// イベント構造体
type Event struct {
	Row       uint32      // 行番号 (0-based)
	Column    uint32      // 列番号 (0-based) byte column
	EventType string      // イベントの種類 ("start" or "end")
	Color     tcell.Style // 対応する色
	NodeType  string      // ノードの種類
}

type Mode interface {
	Name() string

	HasMatchingExtension(filePath string) bool

	// Format formats source code and restores the cursor position.
	// The cursor position based on byte indices
	// Format(source [][]byte, cursorRow, cursorCol int) ([][]byte, int, int, error)

	IsFormattingBeforeSave() bool
	Formatting(source []byte) ([]byte, error)

	GetDefaultTabWidth() int
	GetTabWidth() int
	SetTabWidth(int)

	GetDefaultSoftTab() bool
	GetSoftTab() bool
	SetSoftTab(bool)

	// Tree-sitter
	ColorizeEvents(ctx context.Context, oldTree *sitter.Tree, sourceCode []byte) ([]Event, *sitter.Tree, error)

	// Tree-sitter
	EventIndex(ctx context.Context, currentRow, currentCol int, source [][]byte, events []Event, eventIndex int) (int, tcell.Style, error)
}
