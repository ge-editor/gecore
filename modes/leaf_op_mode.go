package modes

import (
	"context"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/gecore/tree"
	"github.com/ge-editor/keychord"
)

type LeafOpMode struct {
	ModeManager *mode.Manager
	RootNode    *keychord.RootNode
	*tree.OpMode
	SwitchWindow bool
}

func NewLeafOpMode(modeManager *mode.Manager, viewNames string, key func(*keychord.RootNode, *LeafOpMode), switchWindow bool) mode.Mode {
	vm := &LeafOpMode{
		ModeManager:  modeManager,
		RootNode:     keychord.NewRootNode(),
		OpMode:       tree.NewOpMode(viewNames),
		SwitchWindow: switchWindow,
	}

	// ★ ここでキー定義を注入
	if key != nil {
		key(vm.RootNode, vm)
	}

	return vm
}

func (o *LeafOpMode) Name() string {
	return "LeafOpMode"
}

func (o *LeafOpMode) Keys() *keychord.RootNode {
	return o.RootNode
}

func (o *LeafOpMode) WillEnter() {
	o.RootNode.Reset()
	gecore.Echo.SetUsage(leafOpUsage(o.SwitchWindow)...)
}

func (o *LeafOpMode) WillExit() {
	gecore.Echo.ClearUsage()
}

func (o *LeafOpMode) Draw(ctx context.Context) {
	o.OpMode.Draw(o.SwitchWindow)
}

// leafOpUsage is what the echo line shows in the mode: its name and the keys
// (see the bindings in ge/keyset.go). The mode that switches the window is
// left after a window is selected.
func leafOpUsage(switchWindow bool) []string {
	if switchWindow {
		return []string{
			"Switch window",
			"<label>:select",
			"h:hsplit", "v:vsplit", "k:close",
			"C-g:quit",
		}
	}
	return []string{
		"Window ops",
		"<label>:select",
		"h:hsplit", "v:vsplit", "k:close",
		"t/r/b/l:insert at edge", "s:swap direction",
		"Arrows:resize",
		"C-g:quit",
	}
}

func (o *LeafOpMode) CurrentMode() mode.Mode {
	return o
}

// ---------------

// 'h'
func (e LeafOpMode) SplitHorizontally() {
	tree.GetActiveTree().SplitHorizontally()
}

// 'v'
func (e LeafOpMode) SplitVertically() {
	tree.GetActiveTree().SplitVertically()
}

// 'k'
func (e LeafOpMode) Remove() {
	tree.GetActiveTree().DeleteWindow()
}

func (e LeafOpMode) InsertTop() {
	// self parent
	// tree.GetActiveTree().InsertTop(tree.GetActiveTree())

	// root
	tree.GetActiveTree().InsertTop(tree.GetRootTree())
}

func (e LeafOpMode) InsertRight() {
	tree.GetActiveTree().InsertRight(tree.GetRootTree())
}

func (e LeafOpMode) InsertBottom() {
	tree.GetActiveTree().InsertBottom(tree.GetRootTree())
}

func (e LeafOpMode) InsertLeft() {
	tree.GetActiveTree().InsertLeft(tree.GetRootTree())
}

func (e LeafOpMode) SwitchSplitDirection() {
	tree.GetActiveTree().SwitchSplitDirection()
}

// case tcell.KeyCtrlN, tcell.KeyDown:
func (e LeafOpMode) NearestVSplitStepResizeIncrement() {
	node := tree.GetActiveTree().NearestVSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlP, tcell.KeyUp:
func (e LeafOpMode) NearestVSplitStepResizeDecrement() {
	node := tree.GetActiveTree().NearestVSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

// case tcell.KeyCtrlF, tcell.KeyRight:
func (e LeafOpMode) NearestHSplitStepResizeIncrement() {
	node := tree.GetActiveTree().NearestHSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlB, tcell.KeyLeft:
func (e LeafOpMode) NearestHSplitStepResizeDecrement() {
	node := tree.GetActiveTree().NearestHSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

func (e LeafOpMode) SelectName(tev tcell.EventKey) bool {
	modeQuit := func() {
		if e.SwitchWindow {
			e.ModeManager.Cancel()
		}
	}

	leaf := e.OpMode.SelectName(tev.Str())
	if leaf != nil {
		tree.SetActiveTree(leaf)
		modeQuit()
		return true
	}
	modeQuit()
	return false
}
