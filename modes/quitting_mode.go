// gecore/modes/quitting_mode.go
//
// ge/key.go
//   gecore/mode/mode.go
//     gecore/modes/quitting_mode.go
//       gecore/quitguard_manager.go (manage quitguard list)
//         editorleaf/editorleaf.go  (register to quitguard manager)
//           editorleaf/quitguard.go

package modes

import (
	"context"

	"github.com/ge-editor/gecore"
	"github.com/ge-editor/gecore/mode"
	"github.com/ge-editor/keychord"
)

// implements mode.Mode interface
// in ge/mode/mode.ge
type QuittingMode struct {
	ModeManager       *mode.Manager
	RootNode          *keychord.RootNode
	QuitGuardResolver gecore.QuitGuardResolver
}

func NewQuittingMode(mm *mode.Manager, key func(*keychord.RootNode, *QuittingMode)) *QuittingMode {
	resolver := gecore.QuitGuardManager.CurrentQuitGuardResolver()
	// gelog.Info("CurrentQuitGuardResolver", resolver)
	if resolver == nil {
		return nil
	}

	m := &QuittingMode{
		ModeManager:       mm,
		RootNode:          keychord.NewRootNode(),
		QuitGuardResolver: resolver,
	}

	if key != nil {
		key(m.RootNode, m)
	}

	return m
}

func (qm *QuittingMode) Name() string {
	return "QuittingMode"
	//return m.CloseGuard.Name()
}

func (qm *QuittingMode) Keys() *keychord.RootNode {
	if qm == nil {
		return nil
	}
	if qm.RootNode == nil {
		return nil
	}
	return qm.RootNode
}

func (qm *QuittingMode) WillEnter() {
	if qm == nil {
		return
	}
	if qm.QuitGuardResolver == nil {
		return
	}
	// The usage first: the resolver may finish at once (nothing to ask)
	// and quit the app.
	gecore.Echo.SetUsage(quittingUsage()...)
	qm.QuitGuardResolver.WillEnter()
}

func (qm *QuittingMode) WillExit() {
	if qm == nil {
		return
	}
	gecore.Echo.ClearUsage()
	if qm.QuitGuardResolver != nil {
		qm.QuitGuardResolver.WillExit()
	}
}

// quittingUsage is what the echo line shows in the mode. The keys of each
// question (y/n for a modified buffer, a file name for a workspace) are in
// the prompt of the minibuffer; what is common to all of them is the way out.
func quittingUsage() []string {
	return []string{
		"Quitting",
		"answer the prompt",
		"C-g:cancel",
	}
}

func (qm *QuittingMode) Draw(ctx context.Context) {
	// 呼び出し不要
	// qm.QuitGuardResolver.Draw()
}

// return Mode interface
func (qm *QuittingMode) CurrentMode() mode.Mode {
	return qm
}
