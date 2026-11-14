package modes

import (
	"slices"
	"testing"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

func TestLeafOpModeUsageFollowsTheModeStack(t *testing.T) {
	gecore.Echo.ClearUsage()
	mm := mode.NewManager(keychord.NewRootNode())

	mm.Push(NewLeafOpMode(mm, "abc", nil, false))
	if got := gecore.Echo.Usage(); !slices.Equal(got, leafOpUsage(false)) {
		t.Fatalf("after Push: %v", got)
	}

	// The mode that switches the window replaces the usage, and the usage
	// of the first one returns when it is left.
	mm.Push(NewLeafOpMode(mm, "abc", nil, true))
	if got := gecore.Echo.Usage(); !slices.Equal(got, leafOpUsage(true)) {
		t.Fatalf("after the second Push: %v", got)
	}
	mm.Cancel()
	if got := gecore.Echo.Usage(); !slices.Equal(got, leafOpUsage(false)) {
		t.Fatalf("after the second mode is left: %v", got)
	}

	mm.Cancel()
	if got := gecore.Echo.Usage(); len(got) != 0 {
		t.Fatalf("usage stays after the mode is left: %v", got)
	}
}
