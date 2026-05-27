package gecore

import (
	"context"
	"sync"
)

type EventCancelManager struct {
	mu     sync.Mutex
	parent context.Context
	ctxMap map[string]context.Context
	cancel map[string]context.CancelFunc
}

func NewEventCancelManager(parent context.Context) *EventCancelManager {
	return &EventCancelManager{
		parent: parent,
		ctxMap: make(map[string]context.Context),
		cancel: make(map[string]context.CancelFunc),
	}
}

/*
	 func (m *EventCancelManager) Get(key string) context.Context {
		m.mu.Lock()
		defer m.mu.Unlock()

		ctx, ok := m.ctxMap[key]
		if !ok {
			ctx, cancel := context.WithCancel(m.parent)
			m.ctxMap[key] = ctx
			m.cancel[key] = cancel
		}

		return ctx
	}
*/
/* func (m *EventCancelManager) Get(key string) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.ctxMap[key]
	if !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(m.parent)
		m.ctxMap[key] = ctx
		m.cancel[key] = cancel
	}

	return ctx
}
*/

func (m *EventCancelManager) Get(key string) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.ctxMap[key]
	if !ok {
		return nil
		/*
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(m.parent)
			m.ctxMap[key] = ctx
			m.cancel[key] = cancel
		*/
	}

	return ctx
}

func (m *EventCancelManager) Rotate(key string) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	// ここで前世代を止める
	if cancelFunc := m.cancel[key]; cancelFunc != nil {
		cancelFunc()
	}

	ctx, cancel := context.WithCancel(m.parent)

	m.ctxMap[key] = ctx
	m.cancel[key] = cancel

	return ctx
}

func (m *EventCancelManager) Cancel(key string) context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancelFunc := m.cancel[key]; cancelFunc != nil {
		cancelFunc()
	}

	ctx, cancel := context.WithCancel(m.parent)
	m.ctxMap[key] = ctx
	m.cancel[key] = cancel

	return ctx
}

func (m *EventCancelManager) IsCanceled(ctx context.Context, key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, ok := m.ctxMap[key]
	if !ok {
		return true // すでに消えている
	}

	// 現在の世代じゃない
	if current != ctx {
		return true
	}

	// 現世代でも cancel 済み
	select {
	case <-ctx.Done():
		return true
	default:
	}

	return false
}

func (m *EventCancelManager) CancelAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, cancel := range m.cancel {
		cancel()
		delete(m.cancel, key)
		delete(m.ctxMap, key)
	}
}
