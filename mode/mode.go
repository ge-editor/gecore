package mode

import (
	"github.com/ge-editor/keychord"
)

// Mode は「キー定義＋描画＋ライフサイクル」を持つ編集モード
//
// - Name()        : UI 表示用のモード名
// - Keys()        : このモード専用のキー定義（keychord RootNode）
// - WillEnter()   : モードに入る直前に呼ばれる
// - WillExit()    : モードを抜ける直前に呼ばれる
// - Draw()        : モード固有の描画処理
// - CurrentMode() : 現在の Mode のポインタを返す
type Mode interface {
	Name() string
	Keys() *keychord.RootNode

	WillEnter()
	WillExit()
	Draw()

	// 現在の Mode のポインタを返す
	CurrentMode() Mode
}

// Manager はモードをスタックとして管理する
//
// - stack  : 現在有効なモードのスタック（最後が最上位）
// - global : 常に有効なグローバルキー定義
// - active : 現在入力を受け付ける RootNode
type Manager struct {
	stack  []*Mode
	global *keychord.RootNode
	active *keychord.RootNode
}

// 注意 Mode には
// この様はものを追加してはいけない
// type modeManagerArray []*Manager

// NewManager は ModeManager を初期化する
//
// global は「常に有効なキー定義」で、
// モードスタックが空のときに使用される
// func NewManager(global *keychord.RootNode) *Manager {
func NewManager(global *keychord.RootNode) *Manager {
	m := &Manager{
		global: global,
		active: global,
	}
	return m
}

// ActiveKeys は現在有効なキー定義を返す
//
// イベントループは常にこれを参照して Dispatch する
func (m *Manager) ActiveKeys() *keychord.RootNode {
	return m.active
}

// CurrentMode は現在アクティブな Mode を返す。
// スタックが空の場合は nil を返す
func (m *Manager) CurrentMode() Mode {
	if len(m.stack) == 0 {
		return nil
	}
	return *m.stack[len(m.stack)-1]
}

// Push は新しいモードに入る
//
// 1. 既存の最上位モードがあれば WillExit() を呼ぶ
// 2. 新しいモードをスタックに積む
// 3. 新モードの WillEnter() を呼ぶ
// 4. 入力対象を新モードのキー定義に切り替える
func (m *Manager) Push(mode Mode) {
	// 現在の最上位モードを一旦終了させる
	if len(m.stack) > 0 {
		top := (*m.stack[len(m.stack)-1])
		top.WillExit()
	}

	// 新しいモードをスタックに積む
	m.stack = append(m.stack, &mode)

	// 新モード開始
	mode.WillEnter()
	m.active = mode.Keys()
}

// Pop は現在のモードを抜ける
//
// 1. 現在のモードの WillExit() を呼ぶ
// 2. スタックから取り除く
// 3. ひとつ下のモードがあればそれを再開
// 4. なければ global キー定義に戻る
func (m *Manager) Pop() {
	if len(m.stack) == 0 {
		return
	}

	// 現在モードを終了
	top := (*m.stack[len(m.stack)-1])
	top.WillExit()

	// スタックから削除
	m.stack = m.stack[:len(m.stack)-1]

	// ひとつ下のモードに戻るか、グローバルへ
	if len(m.stack) > 0 {
		prev := (*m.stack[len(m.stack)-1])
		prev.WillEnter()
		m.active = prev.Keys()
	} else {
		m.active = m.global
	}
}

// CurrentName は現在のモード名を返す（UI 表示用）
//
// モードスタックが空の場合は空文字列
func (m *Manager) CurrentName() string {
	if len(m.stack) == 0 {
		return ""
	}
	return (*m.stack[len(m.stack)-1]).Name()
}

func (m *Manager) IsInMode() bool {
	return len(m.stack) > 0
}

func (m *Manager) Cancel() {
	// キー途中状態をリセット
	if m.active != nil {
		m.active.Reset()
	}

	// モードが積まれていれば Pop
	if len(m.stack) > 0 {
		m.Pop()
	}
}

func (m *Manager) CancelAll() {
	for len(m.stack) > 0 {
		m.Cancel()
	}
}
