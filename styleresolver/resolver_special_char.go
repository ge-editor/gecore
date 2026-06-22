package styleresolver

import (
	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore/define"
	"github.com/ge-editor/locale"
	"github.com/ge-editor/theme"
)

type SpecialCharResolver struct{}

func (r *SpecialCharResolver) Resolve(
	ctx *Context,
	current tcell.Style,
) (tcell.Style, ResolveFlag) {

	ch := ctx.Cell.Ch

	switch {
	case ch == define.EOF &&
		ctx.IsLastAtRow &&
		ctx.IsEndRow:

		return theme.ColorMarkEOF, Changed

	case ch == '\t':
		return theme.ColorTab, Changed

	case ch == define.LF:
		return theme.ColorMarkNewline, Changed

	case locale.Is(ctx.Cell, locale.CONTROLCODE):
		return theme.ColorControlCode, Changed
	}

	return current, Continue
}
