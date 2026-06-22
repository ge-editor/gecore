package tree

import (
	"context"
	"fmt"
	"sort"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gelog"
)

var ECM *gecore.EventCancelManager
var rootCtx context.Context
var rootCancel context.CancelFunc

func init() {
	rootCtx, rootCancel = context.WithCancel(context.Background())
	ECM = gecore.NewEventCancelManager(rootCtx)
}

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
func (vs *LeafTypesStruct) Register(f LeafTypeFactory, priority int, specifyName ...string) error {
	var name string
	if len(specifyName) > 0 {
		name = specifyName[0]
	} else {
		name = f().Name()
	}

	// 重複チェック
	if _, exists := vs.names[name]; exists {
		return fmt.Errorf("LeafType already registered: %s (%s)", name, f().RealName())
	}

	vs.entries = append(vs.entries, leafTypeEntry{
		name:     name,
		priority: priority,
		factory: func() LeafType {
			leaf := f()
			leaf.SetRegisteredName(name)
			leaf.SetCtx(&LeafContext{
				CancelManager: ECM,
			})
			return leaf
		},
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
		gelog.Error("GetLeafTypeByReginsterName", "name", name)
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
