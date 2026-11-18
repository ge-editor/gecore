package modes

import (
	"slices"
	"testing"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

func TestRedoModeUsageFollowsTheModeStack(t *testing.T) {
	gecore.Echo.ClearUsage()
	mm := mode.NewManager(keychord.NewRootNode())

	mm.Push(NewLeafOpMode(mm, "abc", nil, false))
	mm.Push(NewRedoMode(mm, nil))
	if got := gecore.Echo.Usage(); !slices.Equal(got, redoUsage()) {
		t.Fatalf("after Push: %v", got)
	}
	if !slices.Contains(redoUsage(), "C-g:quit") {
		t.Fatalf("no way out in the usage: %v", redoUsage())
	}

	// The usage of the mode below returns when the mode is left.
	mm.Cancel()
	if got := gecore.Echo.Usage(); !slices.Equal(got, leafOpUsage(false)) {
		t.Fatalf("after the redo mode is left: %v", got)
	}

	mm.Cancel()
	if got := gecore.Echo.Usage(); len(got) != 0 {
		t.Fatalf("usage stays after the mode is left: %v", got)
	}
}

// fakeResolver is a QuitGuardResolver that asks nothing.
type fakeResolver struct {
	entered, exited int
	root            *keychord.RootNode
}

func (f *fakeResolver) Priority() int            { return 0 }
func (f *fakeResolver) Name() string             { return "fake" }
func (f *fakeResolver) Keys() *keychord.RootNode { return f.root }
func (f *fakeResolver) WillEnter()               { f.entered++ }
func (f *fakeResolver) WillExit()                { f.exited++ }
func (f *fakeResolver) Draw() bool               { return false }

func TestQuittingModeUsage(t *testing.T) {
	gecore.Echo.ClearUsage()
	mm := mode.NewManager(keychord.NewRootNode())
	res := &fakeResolver{root: keychord.NewRootNode()}

	// NewQuittingMode takes the resolver from the global manager, so the
	// mode is built directly.
	qm := &QuittingMode{ModeManager: mm, RootNode: keychord.NewRootNode(), QuitGuardResolver: res}

	mm.Push(qm)
	if got := gecore.Echo.Usage(); !slices.Equal(got, quittingUsage()) {
		t.Fatalf("after Push: %v", got)
	}
	if res.entered != 1 {
		t.Fatalf("the resolver was entered %d times", res.entered)
	}
	if !slices.Contains(quittingUsage(), "C-g:cancel") {
		t.Fatalf("no way out in the usage: %v", quittingUsage())
	}

	mm.Cancel()
	if got := gecore.Echo.Usage(); len(got) != 0 {
		t.Fatalf("usage stays after the mode is left: %v", got)
	}
	if res.exited != 1 {
		t.Fatalf("the resolver was left %d times", res.exited)
	}
}

// NewQuittingMode returns nil when there is no resolver. The nil mode must be
// safe to leave.
func TestQuittingModeNilIsSafeToLeave(t *testing.T) {
	var qm *QuittingMode
	qm.WillEnter()
	qm.WillExit()
}
