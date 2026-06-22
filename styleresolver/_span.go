package styleresolver

import "github.com/gdamore/tcell/v3"

type Span struct {
	Start int
	Stop  int

	Style tcell.Style

	Priority uint8
}
