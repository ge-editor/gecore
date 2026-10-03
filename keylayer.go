// gecore/keylayer.go
//
// KeyLayer represents one stage of the key event dispatch pipeline.
//
// The key dispatch pipeline is modeled as an ordered list of layers:
//
//	cancelKeyOnly -> macro -> globalKey -> minibuffer -> mode -> leaf
//
// Each layer has a priority and can be either exclusive or non-exclusive.
// Layers are evaluated in priority order until an event is consumed.
//
// This follows the same design pattern as CancelManager: handlers expose
// a priority and are evaluated from higher priority to lower priority.
// KeyLayer therefore reuses an existing design rather than introducing
// a separate dispatch mechanism.
//
// A layer may use any number of keychord.RootNode instances internally.
// KeyLayerManager is responsible only for ordering the layers and
// determining whether an event may proceed to the next layer.
package gecore

import (
	"sort"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/keychord"
)

// KeyLayer represents one stage of the key dispatch pipeline.
type KeyLayer interface {
	// Name returns the name of the layer for UI display and logging.
	Name() string

	// Active reports whether this layer is currently accepting key events.
	// Inactive layers are skipped without calling Dispatch.
	Active() bool

	// Exclusive reports whether this layer prevents the event from
	// proceeding to subsequent layers.
	//
	// For an exclusive layer, the event is consumed by the layer even when
	// Dispatch returns DispatchNotFound, DispatchPrefix, or
	// DispatchInvalidAfterPrefix.
	//
	// Typical examples are:
	//
	//   - cancelKeyOnly: non-exclusive
	//   - macro: non-exclusive
	//   - globalKey: non-exclusive
	//   - minibuffer: exclusive while active
	//   - mode: exclusive while a mode is active
	//   - leaf: exclusive as the final layer
	Exclusive() bool

	// Priority determines the order in which layers are evaluated.
	// Lower values are evaluated first.
	Priority() int

	// Dispatch processes a key event.
	//
	// The returned string is the current key state for display or logging.
	// It may be passed through directly from keychord.RootNode.Dispatch.
	Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition)
}

// KeyStatusFunc receives the result of a layer dispatch.
//
// Key status reporting is kept outside the layer manager so that UI-related
// side effects remain in the ge package. The keychord RootNode already
// maintains its key state and returns it from Dispatch, so the manager does
// not need to maintain a separate key-state buffer.
type KeyStatusFunc func(layer KeyLayer, status string, res keychord.KeyDispatchTransition)

type KeyLayerManagerStruct struct {
	layers []KeyLayer
}

var keyLayerManager = &KeyLayerManagerStruct{}

func KeyLayerManager() *KeyLayerManagerStruct {
	return keyLayerManager
}

// Register adds a layer to the dispatch pipeline.
//
// Layers are kept in ascending priority order. New layers can therefore
// participate in key dispatch by registering themselves without modifying
// the dispatch loop.
func (m *KeyLayerManagerStruct) Register(l KeyLayer) {
	if l == nil {
		return
	}
	m.layers = append(m.layers, l)
	sort.SliceStable(m.layers, func(i, j int) bool {
		return m.layers[i].Priority() < m.layers[j].Priority()
	})
}

// Layers returns the currently registered layers in dispatch order.
//
// The returned slice is intended for inspection and debugging.
func (m *KeyLayerManagerStruct) Layers() []KeyLayer {
	return m.layers
}

func (m *KeyLayerManagerStruct) Reset() {
	m.layers = nil
}

// Dispatch evaluates the registered layers in priority order.
//
// A layer may consume the event, pass it to the next layer, or prevent
// further dispatch according to its Exclusive setting.
//
// A successful dispatch resets all keychord root nodes. Keeping this
// responsibility here ensures that every successfully executed key
// operation follows the same reset rule.
func (m *KeyLayerManagerStruct) Dispatch(ev tcell.EventKey, onStatus KeyStatusFunc) {
	for _, l := range m.layers {
		if !l.Active() {
			continue
		}

		status, res := l.Dispatch(ev)

		switch res {
		case keychord.DispatchExecuted:
			keychord.ResetAllRootNodes()
			if onStatus != nil {
				onStatus(l, status, res)
			}
			return

		case keychord.DispatchPrefix, keychord.DispatchInvalidAfterPrefix:
			if onStatus != nil {
				onStatus(l, status, res)
			}
			if l.Exclusive() {
				return
			}

		case keychord.DispatchNotFound:
			if l.Exclusive() {
				return
			}
		}
	}
}
