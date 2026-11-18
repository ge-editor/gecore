package gecore

import (
	"slices"
	"testing"

	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

func TestMacroModeUsage(t *testing.T) {
	Echo.ClearUsage()
	mm := mode.NewManager(keychord.NewRootNode())

	mm.Push(NewMacroMode(mm, nil, nil))
	if got := Echo.Usage(); !slices.Equal(got, macroUsage()) {
		t.Fatalf("after Push: %v", got)
	}
	if !slices.Contains(macroUsage(), "e:repeat") || !slices.Contains(macroUsage(), "C-g:quit") {
		t.Fatalf("macro usage: %v", macroUsage())
	}

	mm.Cancel()
	if got := Echo.Usage(); len(got) != 0 {
		t.Fatalf("usage stays after the mode is left: %v", got)
	}
}
