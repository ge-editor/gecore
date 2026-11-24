package pathscope

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Feature names a place where a scope applies. Add a constant, an entry of
// features and a builtinDefault when a feature adopts PathScope.
type Feature string

const (
	// FeatureMark is the Mark list.
	FeatureMark Feature = "mark"
	// FeatureRecent is the list of recently opened files (ESC x recent-open).
	FeatureRecent Feature = "recent"
	// FeatureSearch is the full text search of the explorer: which
	// directories are searched. Current is the directories of the current
	// workspace; Workspaces and All are those of every open workspace (a
	// search has no "everything" beyond the workspaces).
	FeatureSearch Feature = "search"
)

// defaultKey is the key of Settings.Default in the JSON object. It cannot be
// used as a feature name.
const defaultKey = "default"

var features = []Feature{FeatureMark, FeatureRecent, FeatureSearch}

// builtinDefault keeps today's behavior when nothing is configured.
var builtinDefault = map[Feature]Kind{
	FeatureMark:   All,
	FeatureRecent: All,
	FeatureSearch: Current,
}

// Features returns the features that can have a scope.
func Features() []Feature { return append([]Feature(nil), features...) }

// Known reports whether f is a feature of this version.
func (f Feature) Known() bool {
	for _, v := range features {
		if f == v {
			return true
		}
	}
	return false
}

// Settings is the persisted scope setting of a workspace
// (settings.pathScope in the .ge-workspace file).
//
//	{"default": "current", "mark": "all"}
//
// Values and feature names this version does not know are kept as they are,
// so that saving a file never drops what a newer ge wrote.
type Settings struct {
	// Default applies to every feature without its own entry.
	Default Kind
	// ByFeature holds the per feature entries.
	ByFeature map[Feature]Kind
}

// IsZero reports whether nothing is configured.
func (s Settings) IsZero() bool { return s.Default == "" && len(s.ByFeature) == 0 }

// For returns the kind for f: the entry of f, then Default, then the builtin
// default of f (All when it has none). Values that are not valid are skipped.
func (s Settings) For(f Feature) Kind {
	if k := s.ByFeature[f]; k.Valid() {
		return k
	}
	if s.Default.Valid() {
		return s.Default
	}
	if k, ok := builtinDefault[f]; ok {
		return k
	}
	return All
}

// Set sets the kind of f. An empty f sets Default; an empty k removes the
// entry.
func (s *Settings) Set(f Feature, k Kind) {
	if f == "" {
		s.Default = k
		return
	}
	if k == "" {
		delete(s.ByFeature, f)
		return
	}
	if s.ByFeature == nil {
		s.ByFeature = map[Feature]Kind{}
	}
	s.ByFeature[f] = k
}

// Clone returns a copy that shares nothing with s.
func (s Settings) Clone() Settings {
	c := Settings{Default: s.Default}
	if len(s.ByFeature) > 0 {
		c.ByFeature = make(map[Feature]Kind, len(s.ByFeature))
		for f, k := range s.ByFeature {
			c.ByFeature[f] = k
		}
	}
	return c
}

// Invalid lists the stored values that are not valid kinds, as "key=value"
// (sorted). It is meant for a warning when a file is loaded.
func (s Settings) Invalid() []string {
	var bad []string
	if s.Default != "" && !s.Default.Valid() {
		bad = append(bad, fmt.Sprintf("%s=%s", defaultKey, s.Default))
	}
	for f, k := range s.ByFeature {
		if !k.Valid() {
			bad = append(bad, fmt.Sprintf("%s=%s", f, k))
		}
	}
	sort.Strings(bad)
	return bad
}

// MarshalJSON writes the object with sorted keys, so that the same settings
// always give the same bytes (the auto-save compares them).
func (s Settings) MarshalJSON() ([]byte, error) {
	m := make(map[string]string, len(s.ByFeature)+1)
	for f, k := range s.ByFeature {
		if f == "" || string(f) == defaultKey {
			continue
		}
		m[string(f)] = string(k)
	}
	if s.Default != "" {
		m[defaultKey] = string(s.Default)
	}
	return json.Marshal(m)
}

// UnmarshalJSON reads the object. Entries whose value is not a non-empty
// string are ignored.
func (s *Settings) UnmarshalJSON(b []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*s = Settings{}
	for key, raw := range m {
		var v string
		if key == "" || json.Unmarshal(raw, &v) != nil || v == "" {
			continue
		}
		if key == defaultKey {
			s.Default = Kind(v)
		} else {
			s.Set(Feature(key), Kind(v))
		}
	}
	return nil
}
