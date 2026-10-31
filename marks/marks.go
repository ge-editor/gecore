package marks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ge-editor/gecore/state"
	"github.com/ge-editor/gelog"
	"github.com/google/uuid"
)

var Marks *MarksStruct

func init() {
	Marks = NewMarks()
	// 必要に応じて自動保存を有効化
	Marks.StartAutoSave(5 * time.Second)
}

// --------------
// Mark
// --------------

// Mark はすべての leaf の Mark が実装する共通インターフェース
type Mark interface {
	SetID(string)
	GetID() string
	GetStoredMark() (StoredMark, error)
}

// --------------
// StoredMark
// --------------

// 永続化用の汎用構造体
type StoredMark struct {
	ID                string          `json:"id"`
	Label             string          `json:"label"`
	WorkspaceID       string          `json:"workspaceId,omitempty"` // Workspace.ID (UUID)
	WorkspaceName     string          `json:"workspaceName,omitempty"`
	RootTreeName      string          `json:"rootTreeName,omitempty"`
	WorkspaceFilePath string          `json:"workspaceFilePath,omitempty"`
	ExtraData         json.RawMessage `json:"extraData,omitempty"` // leaf ごとの固有データを保持
}

// StoredMark also implements Mark so persisted records can remain in memory.
func (m *StoredMark) SetID(id string)                    { m.ID = id }
func (m *StoredMark) GetID() string                      { return m.ID }
func (m *StoredMark) GetStoredMark() (StoredMark, error) { return *m, nil }

// WorkspaceRef is implemented by marks that refer to a workspace. Marks.
// UpdateWorkspaceFilePath uses it to keep WorkspaceFilePath up to date.
type WorkspaceRef interface {
	GetWorkspaceID() string
	SetWorkspaceID(string)
	GetWorkspaceFilePath() string
	SetWorkspaceFilePath(string)
}

func (m *StoredMark) GetWorkspaceID() string          { return m.WorkspaceID }
func (m *StoredMark) SetWorkspaceID(id string)        { m.WorkspaceID = id }
func (m *StoredMark) GetWorkspaceFilePath() string    { return m.WorkspaceFilePath }
func (m *StoredMark) SetWorkspaceFilePath(p string)   { m.WorkspaceFilePath = p }

// --------------
// MarksStruct
// --------------

func NewMarks() *MarksStruct {
	m := &MarksStruct{
		marks: make([]Mark, 0),
	}
	storedMarks, err := marksLoad()
	if err != nil {
		gelog.Error(err.Error())
		return m
	}

	for i := range storedMarks {
		m.marks = append(m.marks, &storedMarks[i])
	}

	return m
}

type MarksStruct struct {
	mu sync.RWMutex

	marks   []Mark // marks.json に保存するのはこれだけ
	deleted []string

	dirty bool
	stop  chan struct{}
	done  chan struct{}
}

func (m *MarksStruct) StartAutoSave(interval time.Duration) {
	m.stop = make(chan struct{})
	m.done = make(chan struct{})

	go func() {
		defer close(m.done)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = m.saveIfDirty()

			case <-m.stop:
				_ = m.saveIfDirty()
				return
			}
		}
	}()
}

func (m *MarksStruct) StopAutoSave() {
	close(m.stop)
	<-m.done
}

func (m *MarksStruct) Remove(target Mark) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, v := range m.marks {
		if v.GetID() != target.GetID() {
			continue
		}

		m.deleted = append(m.deleted, v.GetID())

		copy(m.marks[i:], m.marks[i+1:])
		m.marks[len(m.marks)-1] = nil
		m.marks = m.marks[:len(m.marks)-1]

		m.dirty = true
		return true
	}

	return false
}

// Mark に UUID を設定した後、Marks に追加する
func (m *MarksStruct) Add(newMark Mark) Mark {
	m.mu.Lock()
	defer m.mu.Unlock()

	if newMark.GetID() == "" {
		newMark.SetID(uuid.NewString())
	}

	// 送り返す
	if updater, ok := newMark.(interface{ UpdateExtraData() }); ok {
		updater.UpdateExtraData()
	}

	m.marks = append([]Mark{newMark}, m.marks...)
	m.dirty = true

	return newMark
}

// UpdateWorkspaceFilePath sets WorkspaceFilePath to path in every mark of the
// workspace with this ID, and returns how many marks changed. Marks saved
// before IDs existed (no WorkspaceID) whose WorkspaceFilePath is path adopt
// the ID.
func (m *MarksStruct) UpdateWorkspaceFilePath(id, path string) int {
	if id == "" || path == "" {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	changed := 0
	for _, mk := range m.marks {
		ref, ok := mk.(WorkspaceRef)
		if !ok {
			continue
		}
		adopted := false
		switch cur := ref.GetWorkspaceID(); {
		case cur == id:
		case cur == "" && ref.GetWorkspaceFilePath() != "" &&
			filepath.Clean(ref.GetWorkspaceFilePath()) == filepath.Clean(path):
			ref.SetWorkspaceID(id)
			adopted = true
		default:
			continue
		}
		if ref.GetWorkspaceFilePath() != path {
			ref.SetWorkspaceFilePath(path)
			adopted = true
		}
		if adopted {
			changed++
			m.dirty = true
		}
	}
	return changed
}

// Update lets fn rewrite the list of marks while the lock is held, so the
// autosave never sees a half-edited mark. fn gets a copy of the list and
// returns the new list and whether anything changed. Marks that are missing
// from the new list are recorded as deleted. An entry may be replaced by
// another Mark with the same ID (e.g. a StoredMark by the editor's mark).
func (m *MarksStruct) Update(fn func(list []Mark) ([]Mark, bool)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := make([]Mark, len(m.marks))
	copy(old, m.marks)

	list, changed := fn(old)
	if !changed {
		return
	}

	kept := make(map[string]struct{}, len(list))
	for _, v := range list {
		kept[v.GetID()] = struct{}{}
	}
	for _, v := range m.marks {
		if _, ok := kept[v.GetID()]; !ok {
			m.deleted = append(m.deleted, v.GetID())
		}
	}

	m.marks = list
	m.dirty = true
}

func (m *MarksStruct) Items() []Mark {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Mark, len(m.marks))
	copy(result, m.marks)
	return result
}

// ----------------

func tmpFilePath() (string, error) {
	dir, err := state.StateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "marks.tmp"), nil
}

// ~/.local/state/ge/marks.json (linux)
func marksFilePath() (string, error) {
	dir, err := state.StateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "marks.json"), nil
}

// ファイルから読み込む際は StoredMark のスライスとしてデシリアライズする
func marksLoad() ([]StoredMark, error) {
	path, err := marksFilePath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []StoredMark{}, nil
		}
		return []StoredMark{}, err
	}

	var s []StoredMark
	if err := json.Unmarshal(b, &s); err != nil {
		return []StoredMark{}, err
	}

	return s, nil
}

// Save is []StoredMark
func (m *MarksStruct) saveIfDirty() error {
	m.mu.Lock()
	if !m.dirty {
		m.mu.Unlock()
		return nil
	}
	// ロックを一時的に解除して読み込み・保存処理を行う
	m.mu.Unlock()

	storedMarks, err := marksLoad()
	if err != nil {
		gelog.Error(err.Error())
	}

	tmp, err := tmpFilePath()
	if err != nil {
		gelog.Error(err.Error())
		return err
	}

	path, err := marksFilePath()
	if err != nil {
		gelog.Error(err.Error())
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		gelog.Error(err.Error())
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	deletedMap := make(map[string]bool)
	for _, id := range m.deleted {
		deletedMap[id] = true
	}

	currentMap := make(map[string]Mark)
	for _, mk := range m.marks {
		currentMap[mk.GetID()] = mk
	}

	var savedMarks []StoredMark

	// 1. 既存ファイルにあったもので、削除されておらず、かつメモリ側で保持されているものを統合
	for _, mk := range storedMarks {
		if deletedMap[mk.ID] {
			continue
		}
		if updated, exists := currentMap[mk.ID]; exists {
			stored, err := updated.GetStoredMark()
			if err != nil {
				return err
			}
			savedMarks = append(savedMarks, stored)
			delete(currentMap, mk.ID)
		} else {
			savedMarks = append(savedMarks, mk)
		}
	}

	// 2. メモリ側に新規追加されたものを追加
	for _, mk := range m.marks {
		newStored, err := mk.GetStoredMark()
		if err != nil {
			return err
		}
		id := newStored.ID

		// currentMap に残っている（＝ファイル側にまだ存在しない新規追加分）ものを処理
		if _, exists := currentMap[id]; exists {
			savedMarks = append([]StoredMark{newStored}, savedMarks...)
			delete(currentMap, id)
		}
	}

	b, err := json.MarshalIndent(savedMarks, "", "  ")
	if err != nil {
		gelog.Error(err.Error())
		return err
	}

	err = os.WriteFile(tmp, b, 0o644)
	if err != nil {
		gelog.Error(err.Error())
		return err
	}

	err = os.Rename(tmp, path)
	if err != nil {
		gelog.Error(err.Error())
		return err
	}

	// 保存成功後にフラグをクリア
	m.dirty = false
	m.deleted = nil

	return nil
}
