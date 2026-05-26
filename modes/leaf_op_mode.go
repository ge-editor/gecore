package modes

import (
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
}

func NewLeafOpMode(modeManager *mode.Manager, viewNames string, key func(*keychord.RootNode, *LeafOpMode)) mode.Mode {
	vm := &LeafOpMode{
		ModeManager: modeManager,
		RootNode:    keychord.NewRootNode(),
		OpMode:      tree.NewOpMode(viewNames),
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
}

func (o *LeafOpMode) WillExit() {

}

func (o *LeafOpMode) Draw() {
	gecore.Echo.AddText("view operations mode")
	o.OpMode.Draw()
}

func (o *LeafOpMode) CurrentMode() mode.Mode {
	return o
}

// ---------------

// 'h'
func (e LeafOpMode) SplitHorizontally() {
	tree.ActiveTreeGet().SplitHorizontally()
}

// 'v'
func (e LeafOpMode) SplitVertically() {
	tree.ActiveTreeGet().SplitVertically()
}

// 'k'
func (e LeafOpMode) Remove() {
	tree.ActiveTreeGet().DeleteWindow()
}

func (e LeafOpMode) InsertTop() {
	tree.ActiveTreeGet().InsertTop()
}

func (e LeafOpMode) InsertRight() {
	tree.ActiveTreeGet().InsertRight()
}

func (e LeafOpMode) InsertBottom() {
	tree.ActiveTreeGet().InsertBottom()
}

func (e LeafOpMode) InsertLeft() {
	tree.ActiveTreeGet().InsertLeft()
}

func (e LeafOpMode) SwitchSplitDirection() {
	tree.ActiveTreeGet().SwitchSplitDirection()
}

// case tcell.KeyCtrlN, tcell.KeyDown:
func (e LeafOpMode) NearestVSplitStepResizeIncrement() {
	node := tree.ActiveTreeGet().NearestVSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlP, tcell.KeyUp:
func (e LeafOpMode) NearestVSplitStepResizeDecrement() {
	node := tree.ActiveTreeGet().NearestVSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

// case tcell.KeyCtrlF, tcell.KeyRight:
func (e LeafOpMode) NearestHSplitStepResizeIncrement() {
	node := tree.ActiveTreeGet().NearestHSplit()
	if node != nil {
		node.StepResize(1)
	}
}

// case tcell.KeyCtrlB, tcell.KeyLeft:
func (e LeafOpMode) NearestHSplitStepResizeDecrement() {
	node := tree.ActiveTreeGet().NearestHSplit()
	if node != nil {
		node.StepResize(-1)
	}
}

func (e LeafOpMode) SelectName(tev tcell.EventKey) bool {
	leaf := e.OpMode.SelectName(tev.Str())
	if leaf != nil {
		tree.ActiveTreeSet(leaf)
		return true
	}
	return false
}
