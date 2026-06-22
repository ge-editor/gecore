package styleresolver

import "github.com/gdamore/tcell/v3"

type CursorLineResolver struct{}

func (r *CursorLineResolver) Resolve(
	ctx *Context,
	current tcell.Style,
) (tcell.Style, ResolveFlag) {

	if ctx.IsCursorLine {
		return current.Underline(true), Changed
	}

	return current, Changed
}
