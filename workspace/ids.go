package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/ge-editor/gecore/state"
)

// initialDirName is the directory (inside the state directory) of the files
// of workspaces whose file the user did not choose.
const initialDirName = "workspaces"

// newID returns a new workspace ID.
func newID() string { return uuid.NewString() }

// validID reports whether id is a UUID in its canonical form
// (lower case, with hyphens).
func validID(id string) bool {
	u, err := uuid.Parse(id)
	return err == nil && u.String() == id
}

// idTaken reports whether a workspace other than self already has id.
func idTaken(id string, self *Workspace) bool {
	for _, x := range workspaces {
		if x != self && x.ID == id {
			return true
		}
	}
	return false
}

// InitialDir is the directory of the initial (not chosen) workspace files:
// <state dir>/workspaces.
func InitialDir() (string, error) {
	dir, err := state.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, initialDirName), nil
}

// initialFilePath is the initial FilePath of the workspace with this ID. The
// ID is used (not the name) so that renaming never moves the file.
func initialFilePath(id string) (string, error) {
	dir, err := InitialDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+FileExt), nil
}

// isInitialPath reports whether path is inside InitialDir.
func isInitialPath(path string) bool {
	if path == "" {
		return false
	}
	dir, err := InitialDir()
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(dir, abs)
	return err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// readFileID returns the ID stored in a workspace file, or "".
func readFileID(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var f struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(b, &f) != nil {
		return ""
	}
	return f.ID
}

var pathHooks []func(*Workspace)

// OnFilePathChanged registers fn to be called whenever the FilePath of a
// workspace is set or changed (creation, loading, SetFile, SaveFile,
// SetDefault). For example, to update the marks that refer to the workspace.
func OnFilePathChanged(fn func(*Workspace)) {
	pathHooks = append(pathHooks, fn)
}

// setFilePath is the only place that assigns FilePath after creation.
func (w *Workspace) setFilePath(path string) {
	w.FilePath = path
	if path == "" {
		return
	}
	for _, fn := range pathHooks {
		fn(w)
	}
}

// FindByID returns the workspace with this ID, or nil.
func FindByID(id string) *Workspace {
	if id == "" {
		return nil
	}
	for _, w := range workspaces {
		if w.ID == id {
			return w
		}
	}
	return nil
}

// IsDefault reports whether w is saved to the default workspace file.
func (w *Workspace) IsDefault() bool {
	p, err := DefaultFilePath()
	return err == nil && w.FilePath != "" && filepath.Clean(w.FilePath) == filepath.Clean(p)
}
