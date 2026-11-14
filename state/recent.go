package state

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The list of recently opened files.
//
// state.json keeps its items in a JSON object, which has no order, and the
// per file cursor entries ("editor:<path>") have no time; so the order of
// recent opening cannot be read from them. The list is kept under its own key,
// newest first. The entries of files that were opened before the list existed
// are still known by their "editor:" key: RecentFiles puts them after the
// list, without a time.

const (
	recentKey    = "recent:files"
	recentLimit  = 300
	legacyPrefix = "editor:"
)

// RecentFile is one file of the list.
type RecentFile struct {
	Path string `json:"path"` // absolute
	// OpenedAt is the time of the last opening, in Unix seconds. It is 0 for a
	// file that is only known by its cursor entry.
	OpenedAt int64 `json:"openedAt"`
}

// RecordRecentFile puts path at the front of the list of recent files. The
// list is saved with the rest of the state (StateSave).
func RecordRecentFile(path string) {
	AppState().recordRecent(path, time.Now())
}

// RecentFiles returns the recent files, the most recently opened first.
func RecentFiles() []RecentFile {
	return AppState().recentFiles()
}

func (s *state) recordRecent(path string, now time.Time) {
	if path == "" {
		return
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.readRecent()
	out := make([]RecentFile, 0, min(len(list)+1, recentLimit))
	out = append(out, RecentFile{Path: path, OpenedAt: now.Unix()})
	for _, f := range list {
		if f.Path == path {
			continue
		}
		if len(out) >= recentLimit {
			break
		}
		out = append(out, f)
	}

	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	s.Items[recentKey] = b
}

func (s *state) recentFiles() []RecentFile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := s.readRecent()

	seen := make(map[string]bool, len(list))
	for _, f := range list {
		seen[f.Path] = true
	}

	var legacy []string
	for key := range s.Items {
		p, ok := strings.CutPrefix(key, legacyPrefix)
		if !ok || p == "" || seen[p] {
			continue
		}
		legacy = append(legacy, p)
	}
	sort.Strings(legacy) // a map has no order; keep the list stable
	for _, p := range legacy {
		list = append(list, RecentFile{Path: p})
	}
	return list
}

// readRecent decodes the list. The caller holds the lock.
func (s *state) readRecent() []RecentFile {
	b, ok := s.Items[recentKey]
	if !ok {
		return nil
	}
	var list []RecentFile
	if err := json.Unmarshal(b, &list); err != nil {
		return nil
	}
	return list
}
