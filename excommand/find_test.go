package excommand

import (
	"context"
	"testing"
)

type fakeCmd struct {
	name    string
	aliases []string
}

func (c fakeCmd) Name() string        { return c.name }
func (c fakeCmd) Aliases() []string   { return c.aliases }
func (c fakeCmd) Description() string { return "" }
func (c fakeCmd) Execute(context.Context, []string) (bool, error) {
	return true, nil
}

func TestFindNearestFirstAndAliasIgnoreCase(t *testing.T) {
	near := &fakeCmd{name: "language", aliases: []string{"lang"}}
	far := &fakeCmd{name: "lang"}
	cmds := []ExCommand{near, far}

	if got := Find(cmds, "LANG"); got != ExCommand(near) {
		t.Fatalf("Find should return the nearest command matching an alias, got %v", got)
	}
	if got := Find(cmds, "nope"); got != nil {
		t.Fatalf("Find should return nil when nothing matches, got %v", got)
	}
}
