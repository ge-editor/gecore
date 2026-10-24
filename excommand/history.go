package excommand

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// HistoryEntry is one executed ESC x input.
//
// Owner says where the command lives: the real name of the leaf type that
// provided it (e.g. "github.com/ge-editor/editorleaf"), or "" for a global
// command. An entry is offered again only where its owner is available, so a
// command of one leaf type does not show up in the history of another.
type HistoryEntry struct {
	Owner string `json:"owner"`
	Input string `json:"input"` // the raw input, command name included
}

type historyFile struct {
	Version int            `json:"version"`
	Entries []HistoryEntry `json:"entries"`
}

// History is the persistent ESC x history, newest first. It is safe for
// concurrent use. With an empty path it only lives in memory.
type History struct {
	mu      sync.Mutex
	path    string
	max     int
	entries []HistoryEntry
}

// LoadHistory reads the history from path (a missing file is an empty
// history) and keeps at most max entries. The returned History is usable even
// when err != nil (an unreadable or corrupted file starts an empty history,
// which the next Add overwrites).
func LoadHistory(path string, max int) (*History, error) {
	h := &History{path: path, max: max}
	if path == "" {
		return h, nil
	}
	entries, err := readHistory(path)
	if err != nil {
		return h, err
	}
	h.entries = capEntries(entries, max)
	return h, nil
}

// Add records input as the newest entry of owner. An older identical entry
// (same owner and input) is dropped, the oldest entries beyond the limit are
// dropped, and the file is rewritten. Blank input is ignored.
//
// Entries another ge instance wrote in the meantime are merged in, so two
// running editors do not erase each other's history. The in-memory history is
// updated even when the returned error says the file could not be written.
func (h *History) Add(owner, input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = mergeEntries([]HistoryEntry{{Owner: owner, Input: input}}, h.entries, h.max)

	if h.path == "" {
		return nil
	}
	if disk, err := readHistory(h.path); err == nil {
		h.entries = mergeEntries(h.entries, disk, h.max)
	}
	return writeHistory(h.path, h.entries)
}

// Remove deletes the entries with this input whose owner accept approves (nil:
// any owner), also from the file. Entries other ge instances wrote in the
// meantime are merged in first, so only the requested ones disappear.
func (h *History) Remove(input string, accept func(owner string) bool) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries := h.entries
	if h.path != "" {
		if disk, err := readHistory(h.path); err == nil {
			entries = mergeEntries(entries, disk, h.max)
		}
	}
	kept := make([]HistoryEntry, 0, len(entries))
	for _, e := range entries {
		if e.Input == input && (accept == nil || accept(e.Owner)) {
			continue
		}
		kept = append(kept, e)
	}
	h.entries = kept

	if h.path == "" {
		return nil
	}
	return writeHistory(h.path, h.entries)
}

// Inputs returns the inputs of the entries whose owner accept approves,
// newest first, without duplicates.
func (h *History) Inputs(accept func(owner string) bool) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var inputs []string
	seen := map[string]bool{}
	for _, e := range h.entries {
		if accept != nil && !accept(e.Owner) {
			continue
		}
		if seen[e.Input] {
			continue
		}
		seen[e.Input] = true
		inputs = append(inputs, e.Input)
	}
	return inputs
}

// mergeEntries returns front followed by the entries of rest that are not in
// front already, cut to max entries (max <= 0: no limit).
func mergeEntries(front, rest []HistoryEntry, max int) []HistoryEntry {
	merged := make([]HistoryEntry, 0, len(front)+len(rest))
	seen := make(map[HistoryEntry]bool, len(front)+len(rest))
	for _, list := range [][]HistoryEntry{front, rest} {
		for _, e := range list {
			if seen[e] {
				continue
			}
			seen[e] = true
			merged = append(merged, e)
		}
	}
	return capEntries(merged, max)
}

func capEntries(entries []HistoryEntry, max int) []HistoryEntry {
	if max > 0 && len(entries) > max {
		return entries[:max]
	}
	return entries
}

func readHistory(path string) ([]HistoryEntry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var f historyFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return f.Entries, nil
}

// writeHistory replaces the file atomically (write a temporary file, rename).
func writeHistory(path string, entries []HistoryEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(historyFile{Version: 1, Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
