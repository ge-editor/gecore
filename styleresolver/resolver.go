package styleresolver

import "github.com/gdamore/tcell/v3"

type ResolveFlag uint8

const (
	None     ResolveFlag = 0
	Continue             = 1 << iota
	Changed
	Stop
)

type Resolver interface {
	Resolve(
		ctx *Context,
		current tcell.Style,
	) (tcell.Style, ResolveFlag)
}
