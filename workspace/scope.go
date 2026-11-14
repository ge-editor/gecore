package workspace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ge-editor/gecore/pathscope"
	"github.com/ge-editor/gelog"
)

// settingsKeyPathScope is the key of the PathScope setting in the settings
// object of the workspace file.
const settingsKeyPathScope = "pathScope"

// Scope returns the configured kind of the PathScope for feature f: the entry
// of f, then the workspace default, then the builtin default of f.
func (w *Workspace) Scope(f pathscope.Feature) pathscope.Kind {
	return w.scope.For(f)
}

// ScopeSettings returns a copy of the PathScope setting of w.
func (w *Workspace) ScopeSettings() pathscope.Settings {
	return w.scope.Clone()
}

// SetScope sets the PathScope of feature f. An empty f sets the default of
// the workspace; an empty k removes the entry (the feature inherits the
// default again). The change is saved by the next auto-save, because it is
// part of the serialization.
func (w *Workspace) SetScope(f pathscope.Feature, k pathscope.Kind) error {
	if k != "" && !k.Valid() {
		return fmt.Errorf("workspace: invalid scope %q (%s)", k, pathscope.KindNames())
	}
	if f != "" && !f.Known() {
		return fmt.Errorf("workspace: unknown scope feature %q", f)
	}
	w.scope.Set(f, k)
	return nil
}

// NewMatcher snapshots the workspaces that kind stands for.
//
//   - All: nothing is restricted.
//   - Workspaces: every open workspace except the default one (IsDefault).
//   - Current: cur.
//
// A workspace brings its ID and its directories (see pathscope.Matcher). A
// workspace without directories is matched by ID only. Like the rest of the
// package it must be called from the UI goroutine; the Matcher it returns does
// not refer to the workspaces any more.
func NewMatcher(cur *Workspace, kind pathscope.Kind) *pathscope.Matcher {
	switch kind {
	case pathscope.Current:
		if cur == nil {
			return pathscope.New(kind, nil, nil)
		}
		return pathscope.New(kind, []string{cur.ID}, cur.Directories)
	case pathscope.Workspaces:
		var ids, dirs []string
		for _, w := range workspaces {
			if w.IsDefault() {
				continue
			}
			ids = append(ids, w.ID)
			dirs = append(dirs, w.Directories...)
		}
		return pathscope.New(kind, ids, dirs)
	}
	return pathscope.New(pathscope.All, nil, nil)
}

// NewSession starts the temporary scope of one mode for feature f. It starts
// with the setting of w. The Matcher is built from the current workspace
// whenever the kind changes.
func (w *Workspace) NewSession(f pathscope.Feature) *pathscope.Session {
	return pathscope.NewSession(w.Scope(f), func(k pathscope.Kind) *pathscope.Matcher {
		return NewMatcher(Current(), k)
	})
}

// marshalSettings is the settings object of the file: the keys this version
// does not know, then the PathScope when something is configured. Without
// either it is {}, exactly what files had before the PathScope existed. The
// keys are sorted, so the same settings always give the same bytes (the
// auto-save compares them).
func (w *Workspace) marshalSettings() json.RawMessage {
	m := make(map[string]json.RawMessage, len(w.extraSettings)+1)
	for k, v := range w.extraSettings {
		m[k] = v
	}
	if !w.scope.IsZero() {
		if b, err := json.Marshal(w.scope); err == nil {
			m[settingsKeyPathScope] = b
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

// loadSettings takes the settings object of a loaded file.
func (w *Workspace) loadSettings(raw json.RawMessage) {
	w.scope, w.extraSettings = parseSettings(raw)
	for _, v := range w.scope.Invalid() {
		gelog.Warn("workspace: unknown pathScope value, using the default", "workspace", w.Name, "value", v)
	}
}

// parseSettings splits a settings object into the PathScope and the other
// keys. A pathScope entry that cannot be read is kept with the other keys, so
// that saving does not destroy it.
func parseSettings(raw json.RawMessage) (pathscope.Settings, map[string]json.RawMessage) {
	var scope pathscope.Settings
	if len(bytes.TrimSpace(raw)) == 0 {
		return scope, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		gelog.Warn("workspace: cannot read settings", "error", err)
		return scope, nil
	}
	extra := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		if k == settingsKeyPathScope {
			if err := json.Unmarshal(v, &scope); err != nil {
				gelog.Warn("workspace: cannot read pathScope", "error", err)
				extra[k] = v
			}
			continue
		}
		extra[k] = v
	}
	if len(extra) == 0 {
		extra = nil
	}
	return scope, extra
}

// readFileSettings reads the settings of a workspace file. ok is false when
// the file cannot be read.
func readFileSettings(path string) (scope pathscope.Settings, extra map[string]json.RawMessage, ok bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return scope, nil, false
	}
	var f struct {
		Settings json.RawMessage `json:"settings"`
	}
	if json.Unmarshal(b, &f) != nil {
		return scope, nil, false
	}
	scope, extra = parseSettings(f.Settings)
	return scope, extra, true
}
