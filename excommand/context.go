package excommand

import (
	"github.com/gdamore/tcell/v3"
	"github.com/ge-editor/gecore/tree"
)

type Context interface {
	Root() *tree.Tree
	CurrentLeaf() tree.Leaf
	Screen() tcell.Screen

	/*
		Editor() *EditorLeaf
		Selection() *Selection

		CommandLine() *ExCommandLine

		Message(msg string)
		Error(err error)
	*/

	Message(msg string)
	Error(err error)
}

/* type ExContext struct {
	root *RootNode
}

func NewExContext(root *RootNode) *ExContext {
	return &ExContext{
		root: root,
	}
}

func (c *ExContext) Root() *RootNode {
	return c.root
}

func (c *ExContext) CurrentLeaf() Leaf {
	return c.root.ActiveLeaf()
}

func (c *ExContext) Screen() tcell.Screen {
	return screen.Get()
}

func (c *ExContext) Editor() *EditorLeaf {
	leaf := c.CurrentLeaf()

	editor, ok := leaf.(*EditorLeaf)
	if !ok {
		return nil
	}

	return editor
}

func (c *ExContext) Selection() *Selection {
	return c.root.Selection()
}

func (c *ExContext) Message(msg string) {
	c.root.StatusBar().SetMessage(msg)
}

func (c *ExContext) Error(err error) {
	c.root.StatusBar().SetError(err.Error())
}
*/
