package tree

import (
	"fmt"
	"sort"
)

var LeafTypes *LeafTypesStruct // User Views

func NewLeafTypes() *LeafTypesStruct {
	return &LeafTypesStruct{
		names: map[string]int{},
	}
}

type LeafTypeFactory func() LeafType

type leafTypeEntry struct {
	name     string
	priority int
	factory  LeafTypeFactory
}

type LeafTypesStruct struct {
	entries []leafTypeEntry
	names   map[string]int
}

// Register LeafType with priority
// return false if name already exists
func (vs *LeafTypesStruct) Register(name string, f LeafTypeFactory, priority int) error {
	// 重複チェック
	if _, exists := vs.names[name]; exists {
		return fmt.Errorf("LeafType already registered: %s", name)
	}

	vs.entries = append(vs.entries, leafTypeEntry{
		name:     name,
		priority: priority,
		factory:  f,
	})

	// priority 昇順でソート
	sort.SliceStable(vs.entries, func(i, j int) bool {
		return vs.entries[i].priority < vs.entries[j].priority
	})

	// index 再構築
	vs.names = make(map[string]int, len(vs.entries))
	for i, e := range vs.entries {
		vs.names[e.name] = i
	}

	return nil
}

// Return LeafType by name
func (vs *LeafTypesStruct) GetLeafTypeByReginsterName(name string) (LeafType, bool) {
	idx, ok := vs.names[name]
	if !ok {
		return nil, false
	}
	return vs.entries[idx].factory(), true
}

// Return highest-priority LeafType as default
func (vs *LeafTypesStruct) GetDefaultLeafType() LeafType {
	if len(vs.entries) == 0 {
		return nil
	}
	return vs.entries[0].factory()
}
