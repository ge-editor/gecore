// gecore/keylayer.go
//
// KeyLayer はキーイベント処理の「優先度付き多段パイプライン」を
// 明示的なデータとして表現する。
//
// これまで ge/key_dispatcher.go の dispatch() 関数に直列 if/switch として
// ベタ書きされていた
//
//	cancelKeyOnly -> macro -> globalKey -> minibuffer -> mode -> leaf
//
// という6段のキー階層を、優先度 (Priority) を持つレイヤーのリストとして
// 明示化する。設計思想は gecore/cancelable_manager.go の CancelManager と
// 同型 (Priority() を持つハンドラを登録し、上位から順に試す) であり、
// 新しいパターンを持ち込むのではなく既存の設計を転用したもの。
//
// 各レイヤーは内部で好きなだけ keychord.RootNode を使ってよい。
// KeyLayerManager はレイヤー同士の「優先順位」と「排他性」だけを扱う。
package gecore

import (
	"sort"

	"github.com/gdamore/tcell/v3"

	"github.com/ge-editor/keychord"
)

// KeyLayer は、キーディスパッチパイプラインの1段を表す。
type KeyLayer interface {
	// UI 表示やログに使うレイヤー名
	Name() string

	// このレイヤーが今キー入力を受け付ける状態か。
	// false ならこのレイヤーは丸ごとスキップされ、Dispatch は呼ばれない。
	Active() bool

	// Dispatch が DispatchNotFound / DispatchPrefix / DispatchInvalidAfterPrefix
	// を返したとき、後続のレイヤーへイベントを流してよいかどうか。
	//
	// true (排他) の場合、後続レイヤー (リーフの自己挿入を含む) には
	// 一切イベントを渡さない。
	//
	// 例:
	//   - cancelKeyOnly    : false (Ctrl+G 以外は普通に下へ流れてほしい)
	//   - macro            : false (C-x e e e... の連続リプレイを妨げないため)
	//   - globalKey        : false (該当しなければ普通の編集キーとして下へ流す)
	//   - minibuffer       : IsActive() と同じ (アクティブなら排他)
	//   - mode (modeManager): IsInMode() と同じ
	//     (Push されたモード中は他のキー階層/リーフへ絶対漏らさない。
	//     旧実装は Prefix/NotFound を無条件に下へ流していたため、
	//     LeafOpMode 中の未知キーがリーフの自己挿入に抜けるバグがあった)
	//   - leaf             : true (最終防衛ライン。ここで必ず消費される)
	Exclusive() bool

	// 優先度。小さいほど先に試される。
	Priority() int

	// 実際のディスパッチ。string は echo 表示用のキー状態文字列
	// (RootNode.Dispatch がそのまま返すものをそのまま使えばよい)。
	Dispatch(ev tcell.EventKey) (string, keychord.KeyDispatchTransition)
}

// KeyStatusFunc はレイヤーの結果を通知するコールバック。
// 旧 `prefix []string` パッケージグローバル変数が担っていた
// echo 表示用の副作用を、呼び出し側 (ge パッケージ) に一本化して渡すためのもの。
//
// 注意: 各 keychord.RootNode は内部で既に "C-x" のような
// キー状態文字列を蓄積して Dispatch の戻り値として返しているため、
// dispatch() 側で改めてグローバル変数に再集計する必要はない。
type KeyStatusFunc func(layer KeyLayer, status string, res keychord.KeyDispatchTransition)

type KeyLayerManagerStruct struct {
	layers []KeyLayer
}

var keyLayerManager = &KeyLayerManagerStruct{}

func KeyLayerManager() *KeyLayerManagerStruct {
	return keyLayerManager
}

// Register はレイヤーを優先度順(昇順)に挿入する。
// userleaf など将来の拡張は、ここに自分のレイヤーを Register するだけで
// パイプラインに参加できる (ge/key_dispatcher.go 本体を触らずに済む)。
func (m *KeyLayerManagerStruct) Register(l KeyLayer) {
	if l == nil {
		return
	}
	m.layers = append(m.layers, l)
	sort.SliceStable(m.layers, func(i, j int) bool {
		return m.layers[i].Priority() < m.layers[j].Priority()
	})
}

// Layers はデバッグ表示用に現在の登録順序を返す。
func (m *KeyLayerManagerStruct) Layers() []KeyLayer {
	return m.layers
}

func (m *KeyLayerManagerStruct) Reset() {
	m.layers = nil
}

// Dispatch は登録済みレイヤーを優先度順に試し、
// 最初に「イベントを消費した」レイヤーで処理を止める。
//
// 旧 dispatch() 関数の6段 if/switch を置き換えるのがこの1メソッド。
func (m *KeyLayerManagerStruct) Dispatch(ev tcell.EventKey, onStatus KeyStatusFunc) {
	for _, l := range m.layers {
		if !l.Active() {
			continue
		}

		status, res := l.Dispatch(ev)

		switch res {
		case keychord.DispatchExecuted:
			// リセットの責任をここ1箇所に集約する。
			// 旧実装では ResetAllRootNodes() が複数箇所に個別に
			// 書かれており、「実行後は必ずリセットする」という
			// ルールがコード規約としてしか存在しなかった。
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
			// 非排他レイヤーのプレフィックス状態は表示だけして
			// 次のレイヤーも試す (globalKey の途中入力表示と、
			// 下位レイヤーの通常キー処理を両立させる)。

		case keychord.DispatchNotFound:
			if l.Exclusive() {
				// 旧実装のバグ修正点:
				// Push されたモードが未知キーで DispatchNotFound を
				// 返しても、以前はそのままリーフの自己挿入まで
				// 抜けてしまっていた。Exclusive() なレイヤーは
				// NotFound でも下に流さない。
				return
			}
		}
	}
}
