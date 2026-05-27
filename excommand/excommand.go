// ESC-x, execute-extended-command

package excommand

type ExCommand struct {
	Name        string
	Aliases     []string
	Description string

	Parent   *ExCommand
	Children []*ExCommand

	// 候補生成
	Complete CompletionProvider

	// 実行
	Run func(ctx Context, args []string) error
}

func (c *ExCommand) FullName() string {
	if c.Parent == nil {
		return c.Name
	}

	return c.Parent.FullName() + " " + c.Name
}

// type CompletionProvider func(ctx Context) []string
type CompletionProvider interface {
	Complete(ctx Context, input string) []Candidate
}

type ParsedExCommand struct {
	Name string
	Args []string
}

type ExCommandHistory struct {
	items []string
	index int
}

type ExCommandError struct {
	Message string
	Hint    string
}
