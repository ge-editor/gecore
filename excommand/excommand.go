// gecore/excommand/excommand.go

package excommand

import "context"

type ExCommand interface {
	Name() string
	Aliases() []string
	Description() string

	// handled=false なら親、または次の registry に解決を続ける。
	Execute(ctx context.Context, args []string) (handled bool, err error)
}

type CommandProvider interface {
	Commands() []ExCommand
}
