package styleresolver

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/locale"
)

type Context struct {
	// Current drawing row
	RowIndex int
	ColIndex int

	Cursor screen.Cursor

	// Context 生成時点で判断不能な場合がある
	// RowIndex, ColIndex と Cursor から判定するにはロジックが必要
	IsCursorLine bool

	// 例えば leaf が Active/Inactive による
	IsCursorEnabled bool

	IsLastAtRow bool
	IsEndRow    bool

	Cell locale.Cell

	Meta any
}

type ResolveResult struct {
	Style tcell.Style
}
