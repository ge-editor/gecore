package tree

import "github.com/ge-editor/gecore"

var ActiveLeaf Leaf

type LeafContext struct {
	CancelManager *gecore.EventCancelManager
}

// Leaf を生成・複製するための Factory + Policy の集合
type LeafType interface {
	NewLeaf() Leaf

	// Mainly used for splitting the screen.
	// In the new window created by splitting the screen,
	// a new tree.Leaf of the same type as the parent tree.Leaf before splitting is cloned.
	//
	// For editorleaf.Editor:
	// Create a new tree.Leaf (Editor) and make it the same as leaf *tree.Leaf
	// direction: "right", "bottom" are not referenced
	NewSiblingLeaf(direction string, leaf Leaf) Leaf

	RealName() string         // Immutable canonical name (e.g. "github.com/ge-editor/editorleaf")
	Name() string             // Mutable registered name (e.g. "editorleaf")
	SetRegisteredName(string) // Overrides Name()

	SetCtx(*LeafContext)
	CancelManager() *gecore.EventCancelManager
}
