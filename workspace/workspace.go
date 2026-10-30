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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ge-editor/gecore/overlay"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/tree"
)

type Workspace struct {
	// ID identifies the workspace for good (a UUID string). Unlike Name and
	// FilePath it never changes, so marks use it to find their workspace.
	ID          string
	Name        string
	Directories []string
	// FilePath is the .ge-workspace file this workspace is saved to. A
	// workspace whose file the user did not choose has the initial value
	// <state dir>/workspaces/<ID>.ge-workspace. See PathDecided.
	FilePath string

	// idRegenerated is true when the ID was missing or invalid in the loaded
	// file and a new one was assigned: the file has to be written again, or
	// the next load would assign yet another ID.
	idRegenerated bool

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

// NameSeparator joins the workspace name and the RootTree name:
// "workspace-name.roottreename". Neither name may contain it.
const NameSeparator = "."

// ValidateName checks a workspace or RootTree name (kind is used in the
// message). A name must not be empty, must be a single line and must not
// contain NameSeparator.
func ValidateName(kind, name string) error {
	switch {
	case strings.TrimSpace(name) == "":
		return fmt.Errorf("workspace: %s name is empty", kind)
	case strings.ContainsAny(name, "\r\n"):
		return fmt.Errorf("workspace: %s name must be a single line", kind)
	case strings.Contains(name, NameSeparator):
		return fmt.Errorf("workspace: %s name %q must not contain %q", kind, name, NameSeparator)
	}
	return nil
}

// QualifiedName returns "workspace-name.roottreename".
func QualifiedName(w *Workspace, rt *tree.RootTreeStruct) string {
	return w.Name + NameSeparator + rt.Name
}

// String returns the qualified name of the entry.
func (e Entry) String() string { return QualifiedName(e.Workspace, e.RootTree) }

// EnsureExt appends FileExt when path does not end with it.
func EnsureExt(path string) string {
	if strings.HasSuffix(strings.ToLower(path), FileExt) {
		return path
	}
	return path + FileExt
}

// nameConflict reports whether another workspace than self makes name
// unusable for a workspace saved to filePath. Names may repeat only when
// both workspaces have a file and the files differ.
func nameConflict(name, filePath string, self *Workspace) bool {
	for _, x := range workspaces {
		if x == self || x.Name != name {
			continue
		}
		// A file that was not chosen (the initial value) counts as no file.
		if !x.PathDecided() || filePath == "" || isInitialPath(filePath) || x.FilePath == filePath {
			return true
		}
	}
	return false
}

// New creates and registers a workspace that has no file yet. Its name must
// be valid (see ValidateName) and not used by another workspace.
func New(name string, dirs []string) (*Workspace, error) {
	return newWorkspace(name, dirs, "", "")
}

// newWorkspace is New for a workspace whose file and ID are known (LoadFile).
// A name that another workspace already has is accepted when both files
// differ. An ID that is missing, invalid, or already used by another workspace
// (a copied file) is replaced by a new one.
func newWorkspace(name string, dirs []string, filePath, id string) (*Workspace, error) {
	if err := ValidateName("workspace", name); err != nil {
		return nil, err
	}
	if nameConflict(name, filePath, nil) {
		return nil, fmt.Errorf("workspace: %q already exists", name)
	}
	regenerated := false
	if !validID(id) || idTaken(id, nil) {
		id = newID()
		regenerated = filePath != "" // loaded from a file that lacks a usable ID
	}
	if filePath == "" {
		// No file chosen: use the initial value.
		if p, err := initialFilePath(id); err == nil {
			filePath = p
		}
	}
	w := &Workspace{ID: id, Name: name, Directories: slices.Clone(dirs), idRegenerated: regenerated}
	workspaces = append(workspaces, w)
	w.setFilePath(filePath)
	return w, nil
}

// Rename changes the name of w. See ValidateName and nameConflict for the
// rules. The file does not change: the initial file is named after the ID.
func (w *Workspace) Rename(name string) error {
	if err := ValidateName("workspace", name); err != nil {
		return err
	}
	if name == w.Name {
		return nil
	}
	if nameConflict(name, w.FilePath, w) {
		return fmt.Errorf("workspace: %q already exists", name)
	}
	w.Name = name
	return nil
}

// RenameRootTree changes the name of rt, a RootTree of w. Names must be valid
// and unique within w.
func (w *Workspace) RenameRootTree(rt *tree.RootTreeStruct, name string) error {
	if !slices.Contains(w.RootTrees, rt) {
		return errors.New("workspace: root tree does not belong to this workspace")
	}
	if err := ValidateName("RootTree", name); err != nil {
		return err
	}
	if name == rt.Name {
		return nil
	}
	if w.findRootTree(name) != nil {
		return fmt.Errorf("workspace: root tree %q already exists in %q", name, w.Name)
	}
	rt.Name = name
	return nil
}

// PathDecided reports whether the user chose the file of w (a workspace file
// that was loaded, or a file given to SetFile). It is false for a workspace
// that has no file, or only the initial file under InitialDir.
func (w *Workspace) PathDecided() bool {
	return w.FilePath != "" && !isInitialPath(w.FilePath)
}

// SetFile makes path (".ge-workspace" is appended when missing) the file of
// w. The file is written by the next save. The workspace of another file may
// keep the same name; the same file may not be used by two workspaces of one
// name.
func (w *Workspace) SetFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("workspace: empty file name")
	}
	abs, err := filepath.Abs(EnsureExt(path))
	if err != nil {
		return err
	}
	if nameConflict(w.Name, abs, w) {
		return fmt.Errorf("workspace: another workspace named %q already uses %s", w.Name, abs)
	}
	if isInitialPath(w.FilePath) && w.FilePath != abs {
		forgetSnapshot(w.FilePath)
		_ = os.Remove(w.FilePath) // the initial file is replaced by the chosen one
	}
	forgetSnapshot(abs) // write it at the next save
	w.setFilePath(abs)  // also updates the marks of w (OnFilePathChanged)
	return nil
}

// Save writes w now: to its file, or to the fallback file in the state
// directory when it has none. Call it from the UI goroutine.
func (w *Workspace) Save() error {
	path, err := autoSavePath(w)
	if err != nil {
		return err
	}
	data, err := MarshalFile(w, path)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(path, data); err != nil {
		return err
	}
	recordSnapshot(path, data)
	return nil
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
			name = strings.ReplaceAll(filepath.Base(abs), NameSeparator, "_")
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
	if err := ValidateName("RootTree", rt.Name); err != nil {
		return err
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
