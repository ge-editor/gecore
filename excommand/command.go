package excommand

import "context"

type Command struct {
	CommandName        string
	CommandAliases     []string
	CommandDescription string
	Run                func(ctx context.Context, args []string) (handled bool, err error)
}

func (c *Command) Name() string        { return c.CommandName }
func (c *Command) Aliases() []string   { return c.CommandAliases }
func (c *Command) Description() string { return c.CommandDescription }

func (c *Command) Execute(ctx context.Context, args []string) (bool, error) {
	if c.Run == nil {
		return false, nil
	}
	return c.Run(ctx, args)
}
