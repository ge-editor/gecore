package tree

var Views *ViewsStruct // User Views
var ActiveLeaf *Leaf

type View interface {
	NewLeaf() *Leaf

	// Mainly used for splitting the screen.
	// In the new window created by splitting the screen,
	// a new tree.Leaf of the same type as the parent tree.Leaf before splitting is cloned.
	//
	// For editorview.Editor:
	// Create a new tree.Leaf (Editor) and make it the same as leaf *tree.Leaf
	// direction: "right", "bottom" are not referenced
	NewSiblingLeaf(direction string, leaf *Leaf) *Leaf

	Name() string // this view name
	// WillClose() // Should I move it here?
}
