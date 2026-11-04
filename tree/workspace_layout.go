package tree

import (
	"encoding/json"
	"fmt"
)

// WorkspaceLeafData is an optional interface for leaves that persist
// workspace-specific state alongside their layout.
type WorkspaceLeafData interface {
	MarshalWorkspaceData() (json.RawMessage, error)
	UnmarshalWorkspaceData(json.RawMessage) error
}

// LayoutNode is the JSON representation of a split tree and its leaves.
type LayoutNode struct {
	Type      string  `json:"type"`
	Direction string  `json:"direction,omitempty"`
	Ratio     float32 `json:"ratio,omitempty"`
	// SplitSaved is true while the split is maximized (Tree.splitSaved).
	// Ratio is then the maximized ratio and SavedRatio the one to restore.
	SplitSaved bool            `json:"splitSaved,omitempty"`
	SavedRatio float32         `json:"savedRatio,omitempty"`
	Children   []*LayoutNode   `json:"children,omitempty"`
	LeafType   string          `json:"leafType,omitempty"`
	// Active marks the leaf that was the RootTree's active window
	// (RootTreeStruct.Active) when the layout was saved.
	Active bool `json:"active,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// MarshalLayout exports a tree's layout and optional per-leaf data.
func (tr *Tree) MarshalLayout() (*LayoutNode, error) {
	return tr.MarshalLayoutWithActive(nil)
}

// MarshalLayoutWithActive exports a tree's layout and marks active as the
// active leaf. active should be a leaf node contained in tr.
func (tr *Tree) MarshalLayoutWithActive(active *Tree) (*LayoutNode, error) {
	return tr.marshalLayout(active)
}

func (tr *Tree) marshalLayout(active *Tree) (*LayoutNode, error) {
	if tr == nil {
		return nil, fmt.Errorf("tree: nil layout")
	}
	if tr.left != nil || tr.top != nil {
		n := &LayoutNode{Type: "split", Ratio: tr.split}
		if tr.splitSaved {
			n.SplitSaved, n.SavedRatio = true, tr.savedSplit
		}
		if tr.left != nil {
			n.Direction = "horizontal"
			left, err := tr.left.marshalLayout(active)
			if err != nil {
				return nil, err
			}
			right, err := tr.right.marshalLayout(active)
			if err != nil {
				return nil, err
			}
			n.Children = []*LayoutNode{left, right}
		} else {
			n.Direction = "vertical"
			top, err := tr.top.marshalLayout(active)
			if err != nil {
				return nil, err
			}
			bottom, err := tr.bottom.marshalLayout(active)
			if err != nil {
				return nil, err
			}
			n.Children = []*LayoutNode{top, bottom}
		}
		return n, nil
	}
	if tr.leaf == nil || tr.leaf.LeafType() == nil {
		return nil, fmt.Errorf("tree: layout contains leaf without type")
	}
	n := &LayoutNode{Type: "leaf", LeafType: tr.leaf.LeafType().Name(), Active: tr == active}
	if state, ok := tr.leaf.(WorkspaceLeafData); ok {
		data, err := state.MarshalWorkspaceData()
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			n.Data = data
		}
	}
	return n, nil
}

// ActiveLeaf returns the leaf node marked active in a restored layout, or nil
// when the layout has no active marker (as in older workspace files).
func (tr *Tree) ActiveLeaf() *Tree {
	if tr == nil {
		return nil
	}
	if tr.leaf != nil {
		if tr.restoreActive {
			return tr
		}
		return nil
	}
	if tr.left != nil {
		if leaf := tr.left.ActiveLeaf(); leaf != nil {
			return leaf
		}
		return tr.right.ActiveLeaf()
	}
	if tr.top != nil {
		if leaf := tr.top.ActiveLeaf(); leaf != nil {
			return leaf
		}
		return tr.bottom.ActiveLeaf()
	}
	return nil
}

// UnmarshalLayout constructs a tree from a saved layout using registered leaf types.
func UnmarshalLayout(n *LayoutNode) (*Tree, error) {
	if n == nil {
		return nil, fmt.Errorf("tree: nil layout node")
	}
	switch n.Type {
	case "leaf":
		if LeafTypes == nil {
			return nil, fmt.Errorf("tree: leaf types are not initialized")
		}
		lt, ok := LeafTypes.GetLeafTypeByReginsterName(n.LeafType)
		if !ok {
			return nil, fmt.Errorf("tree: unknown leaf type %q", n.LeafType)
		}
		leaf := lt.NewLeaf()
		if state, ok := leaf.(WorkspaceLeafData); ok && len(n.Data) > 0 {
			if err := state.UnmarshalWorkspaceData(n.Data); err != nil {
				return nil, fmt.Errorf("tree: restore %s leaf: %w", n.LeafType, err)
			}
		}
		return &Tree{leaf: leaf, restoreActive: n.Active}, nil
	case "split":
		if len(n.Children) != 2 {
			return nil, fmt.Errorf("tree: split requires two children")
		}
		if n.Ratio <= 0 || n.Ratio >= 1 {
			return nil, fmt.Errorf("tree: invalid split ratio %v", n.Ratio)
		}
		a, err := UnmarshalLayout(n.Children[0])
		if err != nil {
			return nil, err
		}
		b, err := UnmarshalLayout(n.Children[1])
		if err != nil {
			return nil, err
		}
		tr := &Tree{split: n.Ratio}
		// An invalid saved ratio is dropped: the maximized state is lost
		// but the layout still loads.
		if n.SplitSaved && n.SavedRatio > 0 && n.SavedRatio < 1 {
			tr.splitSaved, tr.savedSplit = true, n.SavedRatio
		}
		a.parent, b.parent = tr, tr
		switch n.Direction {
		case "horizontal":
			tr.left, tr.right = a, b
		case "vertical":
			tr.top, tr.bottom = a, b
		default:
			return nil, fmt.Errorf("tree: invalid split direction %q", n.Direction)
		}
		return tr, nil
	default:
		return nil, fmt.Errorf("tree: invalid layout node type %q", n.Type)
	}
}

// FirstLeaf returns the first leaf node in display order.
func (tr *Tree) FirstLeaf() *Tree {
	if tr.left != nil {
		return tr.left.FirstLeaf()
	}
	if tr.top != nil {
		return tr.top.FirstLeaf()
	}
	return tr
}
