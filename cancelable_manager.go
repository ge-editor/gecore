// gecore/cancelable_manager.go
package gecore

import (
	"slices"
	"sort"
)

type Cancelable interface {
	Cancel()        // Ctrl-G
	IsActive() bool // 今キャンセル対象か
	Priority() int  // 小さいほど優先（0 が最優先）
}

type CancelManagerStruct struct {
	list []Cancelable
}

var cancelManager = &CancelManagerStruct{}

func CancelManager() *CancelManagerStruct {
	return cancelManager
}

// Manager を Cancelable として登録する
// Cancelable = Manager
// Session = 状態
// Popupmenu / Editorleaf = 部品
// Cancel / Close / Dispatch 集約点は CancelManager
func (m *CancelManagerStruct) Register(c Cancelable) {
	if c == nil {
		return
	}

	/*
		// 重複登録防止
		for _, v := range m.list {
			if v == c {
				return
			}
		}
	*/
	// 重複登録防止
	if slices.Contains(m.list, c) {
		return
	}

	m.list = append(m.list, c)
}

func (m *CancelManagerStruct) Unregister(c Cancelable) {
	for i, v := range m.list {
		if v == c {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return
		}
	}
}

func (m *CancelManagerStruct) CancelTop() bool {
	if len(m.list) == 0 {
		return false
	}

	alive := make([]Cancelable, 0, len(m.list))

	// ① 掃除しながら有効なものを集める
	for _, c := range m.list {
		if c == nil {
			continue
		}
		if c.IsActive() {
			alive = append(alive, c)
		}
	}

	// list を掃除済みに更新
	m.list = alive

	if len(alive) == 0 {
		return false
	}

	// ② 優先度でソート
	sort.SliceStable(alive, func(i, j int) bool {
		return alive[i].Priority() < alive[j].Priority()
	})

	// ③ 最優先だけ Cancel
	alive[0].Cancel()

	// ④ Cancel 後にも掃除（Cancel 内で active=false になる想定）
	clean := alive[:0]
	for _, c := range alive {
		if c.IsActive() {
			clean = append(clean, c)
		}
	}
	m.list = clean

	return true
}

/*
func (m *CancelManagerStruct) CancelAll_1() {
	// Priority 昇順でソート（小さいものから順に Cancel）
	sort.SliceStable(m.list, func(i, j int) bool {
		return m.list[i].Priority() < m.list[j].Priority()
	})

	for _, it := range m.list {
		if it != nil && it.IsActive() {
			it.Cancel()
			// Priority で残すかどうか判定する方法も考えられる
		}
	}

	// シングルトンがあるので削除してはいけない
	// m.list = m.list[:0]
}
*/

func (m *CancelManagerStruct) CancelAll() {
	// Priority 昇順でソート（小さいものから順に Cancel）
	sort.SliceStable(m.list, func(i, j int) bool {
		return m.list[i].Priority() < m.list[j].Priority()
	})

	// 新しいスライスに残す要素だけ集める
	//newList := m.list[:0]
	newList := []Cancelable{}

	for _, it := range m.list {
		if it != nil && it.IsActive() {
			it.Cancel()
			// Priority で残すかどうか判定する方法も考えられる
			newList = append(newList, it) // Cancel できたら残す
		}
	}

	m.list = newList
	// シングルトンがあるので削除してはいけない
	// m.list = m.list[:0]
}

/*
func HandleEvent(ev tcell.EventKey, km *keychord.RootNode) {
	if ev.Key() == tcell.KeyCtrlG {
		if input.Cancel().CancelTop() {
			return
		}
	}

	km.Dispatch(ev)
}
*/
