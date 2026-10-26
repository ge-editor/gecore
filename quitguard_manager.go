// gecore/quitguard_manager.go
//
// ge/key.go
//   ge/mode/mode.go
//     ge/modes/quitting_mode.go
//       gecore/quitguard_manager.go (manage quitguard list)
//         editorleaf/editorleaf.go  (register to quitguard manager)
//           editorleaf/quitguard.go

package gecore

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/ge-editor/gelog"
	"github.com/ge-editor/keychord"
)

type GuardEventStatus int

const (
	EventWait GuardEventStatus = iota
	EventGaveUp
	EventReadyToClose
)

type GuardType int

const (
	GuardWaitResolved       GuardType = iota // 同期で1つずつ処理（UI応答など）
	GuardWaitResolvedSignal                  // 非同期で全体待機（バックグラウンド処理など）
)

type QuitGuardResolver interface {
	Priority() int

	Name() string
	Keys() *keychord.RootNode

	WillEnter()
	WillExit()
	Draw() bool
}

type QuitGuardManagerStruct struct {
	waitResolved       []QuitGuardResolver
	waitResolvedSignal []QuitGuardResolver
	waitResolvedIndex  int

	timeout time.Duration
	quitApp chan struct{}

	mu           sync.Mutex
	isTryClosing bool

	cancelCtx  context.Context
	cancelFunc context.CancelFunc

	isInAsyncPhase bool
}

var (
	QuitGuardManager *QuitGuardManagerStruct = &QuitGuardManagerStruct{
		timeout: 10 * time.Second,
	}
)

func InitQuitGuardManager(quit chan struct{}) {
	QuitGuardManager.quitApp = quit
}

func (q *QuitGuardManagerStruct) Reset(result GuardEventStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.waitResolvedIndex = 0
}

func (q *QuitGuardManagerStruct) CurrentQuitGuardResolver() QuitGuardResolver {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.waitResolvedIndex >= len(q.waitResolved) {
		gelog.Info("CurrentCloseGuard", "nil")
		return nil
	}
	return q.waitResolved[q.waitResolvedIndex]
}

// HandleNextResolver advances to the next resolver, or exits the app when
// none is left.
//
// mu is released before ExitApp and WillEnter are called: sync.Mutex is not
// reentrant, so a resolver whose WillEnter calls back into the manager
// (Register, Unregister, CurrentQuitGuardResolver, ...) would deadlock.
func (q *QuitGuardManagerStruct) HandleNextResolver() {
	q.mu.Lock()
	q.waitResolvedIndex++

	if q.waitResolvedIndex >= len(q.waitResolved) {
		q.mu.Unlock()
		q.ExitApp()
		return
	}

	next := q.waitResolved[q.waitResolvedIndex]
	q.mu.Unlock()

	next.WillEnter()
}

func (q *QuitGuardManagerStruct) Register(c QuitGuardResolver, t GuardType) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isTryClosing {
		return
	}

	switch t {
	case GuardWaitResolved:
		if !slices.Contains(q.waitResolved, c) {
			q.waitResolved = append(q.waitResolved, c)
		}
	case GuardWaitResolvedSignal:
		if !slices.Contains(q.waitResolvedSignal, c) {
			q.waitResolvedSignal = append(q.waitResolvedSignal, c)
		}
	}
}

func (q *QuitGuardManagerStruct) Unregister(c QuitGuardResolver) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isTryClosing {
		return
	}

	remove := func(s []QuitGuardResolver) []QuitGuardResolver {
		for i, v := range s {
			if v == c {
				return append(s[:i], s[i+1:]...)
			}
		}
		return s
	}

	q.waitResolved = remove(q.waitResolved)
	q.waitResolvedSignal = remove(q.waitResolvedSignal)
}

// ExitApp closes quitApp exactly once. The caller must not hold mu.
func (q *QuitGuardManagerStruct) ExitApp() {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case <-q.quitApp:
		return
	default:
		close(q.quitApp)
	}
}
