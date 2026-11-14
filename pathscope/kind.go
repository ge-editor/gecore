// Package pathscope decides which entries (Marks, buffers, ...) belong to the
// scope of a workspace: the entries of the current workspace, of every open
// workspace, or of all.
//
// An entry is in a scope when its workspace ID is one of the scope's
// workspaces, or when its path is under a directory of one of them (see
// Matcher). The package has no dependency on the workspace, tree or screen
// packages, so it can be tested without a terminal.
package pathscope

import (
	"fmt"
	"strings"
)

// Kind is a scope value. It is a string so that a value written by a newer ge
// survives a load/save round trip; Valid tells the values this version knows.
type Kind string

const (
	// All does not restrict anything.
	All Kind = "all"
	// Workspaces is every open workspace except the default one.
	Workspaces Kind = "workspaces"
	// Current is the current workspace.
	Current Kind = "current"
)

var kinds = []Kind{All, Workspaces, Current}

// Kinds returns the valid kinds in display order.
func Kinds() []Kind { return append([]Kind(nil), kinds...) }

// Valid reports whether k is a kind this version knows.
func (k Kind) Valid() bool {
	for _, v := range kinds {
		if k == v {
			return true
		}
	}
	return false
}

func (k Kind) String() string { return string(k) }

// KindNames returns the valid kinds as "all, workspaces, current".
func KindNames() string {
	names := make([]string, len(kinds))
	for i, k := range kinds {
		names[i] = string(k)
	}
	return strings.Join(names, ", ")
}

// ParseKind resolves a kind name. The name is case-insensitive and may be
// shortened to a prefix that is unique ("a", "w", "c").
func ParseKind(s string) (Kind, error) {
	name := strings.ToLower(strings.TrimSpace(s))
	if name == "" {
		return "", fmt.Errorf("empty scope name (%s)", KindNames())
	}
	var found []Kind
	for _, k := range kinds {
		if string(k) == name {
			return k, nil
		}
		if strings.HasPrefix(string(k), name) {
			found = append(found, k)
		}
	}
	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return "", fmt.Errorf("unknown scope %q (%s)", strings.TrimSpace(s), KindNames())
	}
	return "", fmt.Errorf("ambiguous scope %q (%s)", strings.TrimSpace(s), KindNames())
}

// Complete returns the only kind that starts with prefix. ok is false when no
// kind or several kinds match (an empty prefix matches all of them).
func Complete(prefix string) (Kind, bool) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	var found []Kind
	for _, k := range kinds {
		if strings.HasPrefix(string(k), prefix) {
			found = append(found, k)
		}
	}
	if len(found) == 1 {
		return found[0], true
	}
	return "", false
}
