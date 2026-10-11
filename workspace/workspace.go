// Package workspace groups RootTrees (screen layouts) under named
// workspaces.
//
// Ownership: the tree package owns every RootTree (tree.RootTrees).
// Workspace.RootTrees only holds cross references to them, and
// RootTreeStruct.Owner points back to the Workspace. Closing a workspace
// removes its RootTrees from tree.RootTrees as well.
//
// All functions must be called from the main (UI) goroutine.
package workspace

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
)

type Workspace struct {
	Name        string
	Directories []string

	// RootTrees are references only. The owner is tree.RootTrees.
	RootTrees []*tree.RootTreeStruct
}

// Entry is one row of the switch menu.
type Entry struct {
	Workspace *Workspace
	RootTree  *tree.RootTreeStruct
}

var (
	workspaces []*Workspace
	closeHooks []func(*Workspace)
)

// OnClose registers fn to be called while a workspace is being closed,
// after its RootTrees were removed and before it is unregistered.
// (For example, to shut down the language servers of its directories.)
func OnClose(fn func(*Workspace)) {
	closeHooks = append(closeHooks, fn)
}

// All returns the registered workspaces.
func All() []*Workspace {
	return slices.Clone(workspaces)
}

// Find returns the workspace named name, or nil.
func Find(name string) *Workspace {
	for _, w := range workspaces {
		if w.Name == name {
			return w
		}
	}
	return nil
}

// Of returns the workspace that owns rt, or nil.
func Of(rt *tree.RootTreeStruct) *Workspace {
	if rt == nil {
		return nil
	}
	w, _ := rt.Owner.(*Workspace)
	return w
}

// Current returns the workspace of the RootTree being displayed, or nil.
func Current() *Workspace {
	return Of(tree.CurrentRootTree())
}

// New creates and registers a workspace. Names must be unique.
func New(name string, dirs []string) (*Workspace, error) {
	if name == "" {
		return nil, errors.New("workspace: empty name")
	}
	if Find(name) != nil {
		return nil, fmt.Errorf("workspace: %q already exists", name)
	}
	w := &Workspace{Name: name, Directories: slices.Clone(dirs)}
	workspaces = append(workspaces, w)
	return w, nil
}

// InitDefault creates the first workspace and adopts the RootTree that
// already exists (created by tree.SetRootTree at startup).
// The name is the base name of the first directory, or "default".
func InitDefault(dirs []string) (*Workspace, error) {
	rt := tree.CurrentRootTree()
	if rt == nil {
		return nil, errors.New("workspace: no current root tree")
	}

	name := "default"
	if len(dirs) > 0 {
		if abs, err := filepath.Abs(dirs[0]); err == nil {
			name = filepath.Base(abs)
		}
	}
	w, err := New(name, dirs)
	if err != nil {
		return nil, err
	}
	rt.Name = "main"
	if err := w.Attach(rt); err != nil {
		return nil, err
	}
	return w, nil
}

// Attach makes w the owner of rt and registers rt in tree.RootTrees.
// RootTree names must be unique within a workspace.
func (w *Workspace) Attach(rt *tree.RootTreeStruct) error {
	if rt == nil {
		return errors.New("workspace: nil root tree")
	}
	if rt.Owner != nil && rt.Owner != any(w) {
		return errors.New("workspace: root tree already belongs to another workspace")
	}
	if slices.Contains(w.RootTrees, rt) {
		return nil
	}
	if w.findRootTree(rt.Name) != nil {
		return fmt.Errorf("workspace: root tree %q already exists in %q", rt.Name, w.Name)
	}
	rt.Owner = w
	tree.AddRootTree(rt)
	w.RootTrees = append(w.RootTrees, rt)
	return nil
}

// NewRootTree creates a RootTree with one window and attaches it.
// It is not displayed until Switch is called.
func (w *Workspace) NewRootTree(name string, lt tree.LeafType) (*tree.RootTreeStruct, error) {
	if w.findRootTree(name) != nil {
		return nil, fmt.Errorf("workspace: root tree %q already exists in %q", name, w.Name)
	}
	root := tree.NewRootTree(lt)
	rt := &tree.RootTreeStruct{Name: name, Root: root, Active: root}
	if err := w.Attach(rt); err != nil {
		return nil, err
	}
	return rt, nil
}

// RemoveRootTree removes rt from tree.RootTrees and from w.
// The RootTree being displayed cannot be removed.
func (w *Workspace) RemoveRootTree(rt *tree.RootTreeStruct) error {
	i := slices.Index(w.RootTrees, rt)
	if i < 0 {
		return errors.New("workspace: root tree does not belong to this workspace")
	}
	if err := tree.RemoveRootTree(rt); err != nil {
		return err
	}
	w.RootTrees = slices.Delete(w.RootTrees, i, i+1)
	rt.Owner = nil
	return nil
}

// Close removes all RootTrees of w from tree.RootTrees and unregisters w.
// If a RootTree of w is being displayed, it first switches to a RootTree
// of another workspace. The last workspace cannot be closed.
// Buffers stay open (editorleaf.BufferSets is shared).
func (w *Workspace) Close() error {
	i := slices.Index(workspaces, w)
	if i < 0 {
		return errors.New("workspace: not registered")
	}

	if Of(tree.CurrentRootTree()) == w {
		var next *tree.RootTreeStruct
		for _, rt := range tree.RootTrees {
			if Of(rt) != w {
				next = rt
				break
			}
		}
		if next == nil {
			return errors.New("workspace: cannot close the last workspace")
		}
		if err := Switch(Entry{Workspace: Of(next), RootTree: next}); err != nil {
			return err
		}
	}

	for _, rt := range slices.Clone(w.RootTrees) {
		if err := w.RemoveRootTree(rt); err != nil {
			return err
		}
	}
	for _, fn := range closeHooks {
		fn(w)
	}
	workspaces = slices.Delete(workspaces, i, i+1)
	return nil
}

func (w *Workspace) findRootTree(name string) *tree.RootTreeStruct {
	for _, rt := range w.RootTrees {
		if rt.Name == name {
			return rt
		}
	}
	return nil
}

// Entries returns every (workspace, RootTree) pair in menu order.
func Entries() []Entry {
	var es []Entry
	for _, w := range workspaces {
		for _, rt := range w.RootTrees {
			es = append(es, Entry{Workspace: w, RootTree: rt})
		}
	}
	return es
}

// Switch displays e.RootTree and lays out the screen again.
func Switch(e Entry) error {
	if err := tree.SwitchRootTree(e.RootTree); err != nil {
		return err
	}
	overlay.OverlayManager().Layout(screen.Get().Rect)
	return nil
}
