package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ge-editor/gecore/tree"
)

// File is the on-disk .ge-workspace representation.
type File struct {
	Version int `json:"version"`
	// ID is the UUID of the workspace. A file without a valid one gets a new
	// ID when it is loaded.
	ID        string                      `json:"id,omitempty"`
	Name      string                      `json:"name"`
	Folders   []Folder                    `json:"folders"`
	RootTrees map[string]*tree.LayoutNode `json:"rootTrees"`
	// Order lists the RootTree names in menu order (a JSON object has no
	// order). Names missing from it follow in alphabetical order.
	Order []string `json:"order,omitempty"`
	// Current is the RootTree that was displayed when the file was saved. It
	// is only written when that RootTree belongs to this workspace.
	Current  string          `json:"current,omitempty"`
	Settings json.RawMessage `json:"settings,omitempty"`
}

type Folder struct {
	Path string `json:"path"`
}

// MarshalFile snapshots a workspace. Folder paths are written relative to the
// workspace file when possible, and RootTrees are serialized recursively.
func MarshalFile(w *Workspace, filePath string) ([]byte, error) {
	if w == nil {
		return nil, fmt.Errorf("workspace: nil workspace")
	}
	base, err := filepath.Abs(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}
	f := File{
		Version:   1,
		ID:        w.ID,
		Name:      w.Name,
		RootTrees: make(map[string]*tree.LayoutNode, len(w.RootTrees)),
		Settings:  json.RawMessage(`{}`),
	}
	for _, dir := range w.Directories {
		p := dir
		if abs, e := filepath.Abs(dir); e == nil {
			if rel, e := filepath.Rel(base, abs); e == nil {
				p = rel
			}
		}
		f.Folders = append(f.Folders, Folder{Path: filepath.ToSlash(p)})
	}
	if cur := tree.CurrentRootTree(); cur != nil && Of(cur) == w {
		f.Current = cur.Name
	}
	for _, rt := range w.RootTrees {
		f.Order = append(f.Order, rt.Name)
		if rt.Root == nil {
			return nil, fmt.Errorf("workspace: RootTree %q has no layout", rt.Name)
		}
		layout, e := rt.Root.MarshalLayout()
		if e != nil {
			return nil, fmt.Errorf("workspace: serialize RootTree %q: %w", rt.Name, e)
		}
		f.RootTrees[rt.Name] = layout
	}
	return json.MarshalIndent(f, "", "  ")
}

// SaveFile writes a workspace atomically as a .ge-workspace JSON file.
func SaveFile(w *Workspace, path string) error {
	b, err := MarshalFile(w, path)
	if err != nil {
		return err
	}
	if err = writeFileAtomic(path, b); err != nil {
		return err
	}
	if abs, err := filepath.Abs(path); err == nil {
		w.setFilePath(abs)
	}
	return nil
}

// writeFileAtomic writes b to a temporary file in the same directory, syncs
// it and renames it over path, so a crash never leaves a half-written file.
func writeFileAtomic(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ge-workspace-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err = tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// LoadFile reads a workspace file, registers the workspace and its RootTrees,
// then switches to its first RootTree. Paths in folders resolve relative to the
// file. On failure, all partially registered state is rolled back.
func LoadFile(path string) (*Workspace, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("workspace: parse %s: %w", path, err)
	}
	if f.Version != 1 {
		return nil, fmt.Errorf("workspace: unsupported file version %d", f.Version)
	}
	if f.Name == "" {
		return nil, fmt.Errorf("workspace: missing name")
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(f.Folders))
	for _, folder := range f.Folders {
		p := folder.Path
		if p == "" {
			return nil, fmt.Errorf("workspace: folder path is empty")
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(base, filepath.FromSlash(p))
		}
		abs, e := filepath.Abs(p)
		if e != nil {
			return nil, e
		}
		dirs = append(dirs, abs)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	w, err := newWorkspace(f.Name, dirs, abs, f.ID)
	if err != nil {
		return nil, err
	}
	names := orderedNames(f)
	rollback := func(cause error) (*Workspace, error) {
		for _, rt := range w.RootTrees {
			_ = tree.RemoveRootTree(rt)
			rt.Owner = nil
		}
		w.RootTrees = nil
		for _, fn := range closeHooks {
			fn(w)
		}
		for i, x := range workspaces {
			if x == w {
				workspaces = append(workspaces[:i], workspaces[i+1:]...)
				break
			}
		}
		return nil, cause
	}
	for _, name := range names {
		layout := f.RootTrees[name]
		if layout == nil {
			return rollback(fmt.Errorf("workspace: RootTree %q has no layout", name))
		}
		root, e := tree.UnmarshalLayout(layout)
		if e != nil {
			return rollback(fmt.Errorf("workspace: restore RootTree %q: %w", name, e))
		}
		rt := &tree.RootTreeStruct{Name: name, Root: root}
		if e = w.Attach(rt); e != nil {
			return rollback(e)
		}
		rt.Active = root.FirstLeaf()
	}
	if len(w.RootTrees) == 0 {
		if _, err = w.NewRootTree("main", tree.LeafTypes.GetDefaultLeafType()); err != nil {
			return rollback(err)
		}
	}
	show := w.RootTrees[0]
	if f.Current != "" {
		if rt := w.findRootTree(f.Current); rt != nil {
			show = rt
		}
	}
	if err := Switch(Entry{Workspace: w, RootTree: show}); err != nil {
		return rollback(err)
	}
	return w, nil
}

// orderedNames returns the RootTree names of f: first those listed in
// f.Order (ignoring unknown and duplicate names), then the others sorted.
func orderedNames(f File) []string {
	names := make([]string, 0, len(f.RootTrees))
	seen := make(map[string]bool, len(f.RootTrees))
	for _, name := range f.Order {
		if _, ok := f.RootTrees[name]; ok && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	var rest []string
	for name := range f.RootTrees {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	return append(names, rest...)
}
