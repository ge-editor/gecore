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
	Version   int                         `json:"version"`
	Name      string                      `json:"name"`
	Folders   []Folder                    `json:"folders"`
	RootTrees map[string]*tree.LayoutNode `json:"rootTrees"`
	Settings  json.RawMessage             `json:"settings,omitempty"`
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
	for _, rt := range w.RootTrees {
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
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
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
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if abs, err := filepath.Abs(path); err == nil {
		w.FilePath = abs
	}
	return nil
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
	w, err := New(f.Name, dirs)
	if err != nil {
		return nil, err
	}
	if abs, e := filepath.Abs(path); e == nil {
		w.FilePath = abs
	}
	names := make([]string, 0, len(f.RootTrees))
	for name := range f.RootTrees {
		names = append(names, name)
	}
	sort.Strings(names)
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
	if err := Switch(Entry{Workspace: w, RootTree: w.RootTrees[0]}); err != nil {
		return rollback(err)
	}
	return w, nil
}
