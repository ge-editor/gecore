package pathscope

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/ge-editor/utils"
)

// root is one directory of a scope.
type root struct {
	lex   string // absolute and cleaned, as given
	canon string // with symbolic links resolved
	fold  bool   // the file system of the directory ignores case (utils.IsCaseInsensitivePath)
}

// Matcher decides whether an entry is in a scope. It is a snapshot: it does
// not refer to a Workspace, so later changes of the workspaces do not affect
// it, and it may be used from other goroutines (the cache of resolved paths is
// guarded by mu).
type Matcher struct {
	kind  Kind
	all   bool                // Match accepts everything
	ids   map[string]struct{} // workspace IDs of the scope
	roots []root              // directories of those workspaces

	mu    sync.Mutex
	canon map[string]string // path -> path with symbolic links resolved
}

// New builds a Matcher from the IDs and the directories of the workspaces in
// the scope. For All (and for a kind that is not valid) ids and dirs are
// ignored and everything matches.
func New(kind Kind, ids []string, dirs []string) *Matcher {
	m := &Matcher{kind: kind, canon: map[string]string{}}
	if !kind.Valid() || kind == All {
		m.kind = All
		m.all = true
		return m
	}
	m.ids = make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id != "" {
			m.ids[id] = struct{}{}
		}
	}
	seen := map[string]bool{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = filepath.Clean(dir)
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		canon := CanonicalPath(abs)
		m.roots = append(m.roots, root{lex: abs, canon: canon, fold: utils.IsCaseInsensitivePath(canon)})
	}
	return m
}

// Kind returns the kind the Matcher was built for.
func (m *Matcher) Kind() Kind { return m.kind }

// Match reports whether an entry is in the scope: its workspace ID is one of
// the scope's IDs, or its path is under one of the scope's directories.
// workspaceID may be "" for entries that have none; then only the path
// counts. An empty path never matches by path.
func (m *Matcher) Match(workspaceID, path string) bool {
	if m.all {
		return true
	}
	// The ID needs no I/O, so it is checked first.
	if workspaceID != "" {
		if _, ok := m.ids[workspaceID]; ok {
			return true
		}
	}
	if path == "" || len(m.roots) == 0 {
		return false
	}
	p, err := filepath.Abs(path)
	if err != nil {
		p = filepath.Clean(path)
	}
	// Compare the strings first (no I/O). Only when that fails, resolve the
	// symbolic links of the path (cached).
	return m.within(p) || m.within(m.canonical(p))
}

// within reports whether p is under one of the roots (by string comparison).
func (m *Matcher) within(p string) bool {
	for _, r := range m.roots {
		if isUnder(r.lex, p, r.fold) || (r.canon != r.lex && isUnder(r.canon, p, r.fold)) {
			return true
		}
	}
	return false
}

// canonical returns CanonicalPath(p), remembering the result.
func (m *Matcher) canonical(p string) string {
	m.mu.Lock()
	c, ok := m.canon[p]
	m.mu.Unlock()
	if ok {
		return c
	}
	c = CanonicalPath(p) // file system access: outside the lock
	m.mu.Lock()
	m.canon[p] = c
	m.mu.Unlock()
	return c
}

// isUnder reports whether p is dir itself or inside it. It compares whole path
// elements, so /a/foobar is not under /a/foo.
func isUnder(dir, p string, fold bool) bool {
	if fold {
		dir, p = strings.ToLower(dir), strings.ToLower(p)
	}
	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// CanonicalPath returns the absolute path with symbolic links resolved. For a
// path that does not exist, the part that exists is resolved and the rest is
// appended as it is.
func CanonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	abs = filepath.Clean(abs)
	candidate := abs
	var suffix []string
	for {
		if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved)
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
		suffix = append(suffix, filepath.Base(candidate))
		candidate = parent
	}
	return abs
}
