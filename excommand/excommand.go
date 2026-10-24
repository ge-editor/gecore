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

// ArgCompleter is implemented by commands that can complete their argument.
// The ESC x prompt calls CompleteArgs for the text typed after the command
// name ("language go" -> "go") and shows the result in the popup menu.
type ArgCompleter interface {
	CompleteArgs(input string) []string
}

type CommandProvider interface {
	Commands() []ExCommand
}
