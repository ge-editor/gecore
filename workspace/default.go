package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/ge-editor/gecore/state"
)

const (
	// FileExt is the conventional extension of a workspace file. It is only a
	// convention: nothing refuses a file that has another extension.
	FileExt = ".ge-workspace"

	// DefaultFileName is the file of the default workspace. It lives in the
	// state directory next to marks.json, for example
	// ~/.local/state/ge/default.ge-workspace on Linux.
	DefaultFileName = "default" + FileExt

	// DefaultAutoSaveInterval is how often the default workspace is checked
	// for changes (marks use the same interval).
	DefaultAutoSaveInterval = 5 * time.Second
)

// DefaultFilePath returns the path of the default workspace file.
func DefaultFilePath() (string, error) {
	dir, err := state.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DefaultFileName), nil
}

// savedMu guards savedSnapshots, which is written from the auto-save
// goroutine (after the UI lock was released).
var (
	savedMu        sync.Mutex
	savedSnapshots = map[string][]byte{} // file path -> content last written (or loaded)
)

// SetDefault makes w the default workspace: its file is DefaultFilePath. The
// file is written at the next check even if w did not change, so a plain
// start replaces the previous session. Marks recorded in w refer to this
// file, so they find the live workspace again.
func SetDefault(w *Workspace) {
	path, err := DefaultFilePath()
	if err != nil || w == nil {
		return
	}
	// The default workspace is replaced at every plain start. Keep the ID of
	// the saved one so that marks of the default workspace still find it.
	if id := readFileID(path); validID(id) && id != w.ID && !idTaken(id, w) {
		w.ID = id
	}
	forgetSnapshot(path)
	w.setFilePath(path)
}

func forgetSnapshot(path string) {
	savedMu.Lock()
	delete(savedSnapshots, path)
	savedMu.Unlock()
}

func recordSnapshot(path string, data []byte) {
	savedMu.Lock()
	savedSnapshots[path] = data
	savedMu.Unlock()
}

// autoSavePath returns the file w is saved to. FilePath is normally set when
// the workspace is created; a workspace that has none (the state directory
// was unavailable then) gets the initial file now.
func autoSavePath(w *Workspace) (string, error) {
	if w.FilePath != "" {
		return w.FilePath, nil
	}
	path, err := initialFilePath(w.ID)
	if err != nil {
		return "", err
	}
	w.setFilePath(path)
	return w.FilePath, nil
}

// MarkSaved records the current state of w as already written, so that a
// workspace just loaded from its file is not rewritten until it changes.
func MarkSaved(w *Workspace) {
	if w == nil || w.FilePath == "" {
		return
	}
	if w.idRegenerated {
		// The loaded file has no usable ID: leave it unrecorded so that the
		// next check writes the new ID.
		return
	}
	data, err := MarshalFile(w, w.FilePath)
	if err != nil {
		return
	}
	recordSnapshot(w.FilePath, data)
}

// PendingSave is a snapshot of one workspace that has to be written.
type PendingSave struct {
	path string
	data []byte
}

// PrepareSaves is the dirty check of every workspace: it serializes each one
// and compares the result with what was written last. Only changed
// workspaces are returned. Cursor moves, splits, opened files and RootTree
// switches all change the serialization, so no code that changes them has to
// set a flag. A workspace that fails to serialize is reported in the error
// and does not stop the others.
//
// Call it with the UI lock held (it reads leaf state). It is cheap; the file
// writes are done by Save.Write, which does not need the lock.
func PrepareSaves() ([]*PendingSave, error) {
	var saves []*PendingSave
	var errs []error
	for _, w := range workspaces {
		path, err := autoSavePath(w)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		data, err := MarshalFile(w, path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", w.Name, err))
			continue
		}
		savedMu.Lock()
		same := bytes.Equal(data, savedSnapshots[path])
		savedMu.Unlock()
		if !same {
			saves = append(saves, &PendingSave{path: path, data: data})
			w.idRegenerated = false // the new ID is written now
		}
	}
	return saves, errors.Join(errs...)
}

// Write writes the snapshot atomically (no UI lock needed). When it fails the snapshot stays
// unrecorded, so the next check tries again.
func (s *PendingSave) Write() error {
	if err := writeFileAtomic(s.path, s.data); err != nil {
		return err
	}
	recordSnapshot(s.path, s.data)
	return nil
}
