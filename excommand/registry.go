package excommand

import (
	"errors"
	"strings"

	"github.com/ge-editor/utils"
)

type ExCommandRegistry struct {
	root     []*ExCommand
	indexMap map[string]*ExCommand
}

func NewExCommandRegistry() *ExCommandRegistry {
	return &ExCommandRegistry{
		indexMap: map[string]*ExCommand{},
	}
}

func (r *ExCommandRegistry) Register(cmd *ExCommand) {
	r.root = append(r.root, cmd)
	r.registerRecursive(cmd, nil)
}

func (r *ExCommandRegistry) registerRecursive(
	cmd *ExCommand,
	parent *ExCommand,
) {
	cmd.Parent = parent

	// index
	path := cmd.FullName()

	if _, exists := r.indexMap[path]; exists {
		panic("duplicate ex command: " + path)
	}

	r.indexMap[path] = cmd

	// recursive
	for _, child := range cmd.Children {
		r.registerRecursive(child, cmd)
	}
}

var (
	ErrCommandNotFound = errors.New("Extended command not found")
)

func (r *ExCommandRegistry) Find(
	input string,
) (*ExCommand, error) {

	tokens := strings.Fields(input)

	var current *ExCommand
	children := r.root

	for _, token := range tokens {

		found := prefixMatch(children, token)

		if found == nil {
			return nil, ErrCommandNotFound
		}

		current = found
		children = found.Children
	}

	return current, nil
}

func prefixMatch(children []*ExCommand, token string) *ExCommand {
	for _, child := range children {
		if utils.PrefixMatchTokens(child.Name, token) {
			return child
		}
	}
	return nil
}

/*
var splitCmd = &ExCommand{
	Name:        "split",
	Description: "Split editor",
	Children: []*ExCommand{
		{
			Name: "vertical",
			Run: func(ctx Context, args []string) error {
				return SplitVertical(ctx)
			},
		},
		{
			Name: "horizontal",
			Run: func(ctx Context, args []string) error {
				return SplitHorizontal(ctx)
			},
		},
	},
}

registry.Register(splitCmd)
*/
