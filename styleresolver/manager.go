package styleresolver

import (
	"github.com/gdamore/tcell/v3"
)

type Manager struct {
	resolvers []Resolver
}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) Add(r Resolver) {
	m.resolvers = append(m.resolvers, r)
}

func (m *Manager) Resolve(
	ctx *Context,
	base tcell.Style,
) (tcell.Style, ResolveFlag) {
	style := base

	if m == nil {
		return style, None
	}

	for _, r := range m.resolvers {
		if r == nil {
			continue
		}

		var flags ResolveFlag

		style, flags = r.Resolve(ctx, style)

		if flags&Continue != 0 {
			continue
		}
		if flags&Changed != 0 {
			return style, flags
		}
		if flags&Stop != 0 {
			return style, flags
		}
	}

	return style, None
}
