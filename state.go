package gecore

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	stateOnce sync.Once
	stateInst *state
)

type state struct {
	mu sync.RWMutex

	Version int `json:"version"`

	Items map[string]json.RawMessage `json:"items"`
}

func AppState() *state {
	stateOnce.Do(func() {
		s, err := stateLoad()
		if err != nil {
			s = newState()
		}

		stateInst = s
	})

	return stateInst
}

func newState() *state {
	return &state{
		Version: 1,
		Items:   map[string]json.RawMessage{},
	}
}

func stateLoad() (*state, error) {
	path, err := stateFilePath()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newState(), nil
		}
		return nil, err
	}

	var s state
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}

	if s.Items == nil {
		s.Items = map[string]json.RawMessage{}
	}

	return &s, nil
}

func StateSave() error {
	s := AppState()

	path, err := stateFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	b, err := json.MarshalIndent(
		s,
		"",
		"  ",
	)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o644)
}

//

func stateFilePath() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "state.json"), nil
}

// ~/.config/ge/config.json		# ユーザー設定
// ~/.local/state/ge/state.json	# cursor, history
// ~/.cache/ge/					# 一時cache
func stateDir() (string, error) {
	switch runtime.GOOS {

	case "linux":
		if v := os.Getenv("XDG_STATE_HOME"); v != "" {
			return filepath.Join(v, "ge"), nil
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, ".local", "state", "ge"), nil

	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(
			home,
			"Library",
			"Application Support",
			"ge",
		), nil

	case "windows":
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = os.Getenv("APPDATA")
		}

		return filepath.Join(base, "ge"), nil
	}

	return os.UserConfigDir()
}

// generic-ish API

func (s *state) Save(
	key string,
	v any,
) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.Items[key] = b
	return nil
}

func (s *state) Load(
	key string,
	v any,
) bool {

	s.mu.RLock()
	b, ok := s.Items[key]
	s.mu.RUnlock()

	if !ok {
		return false
	}

	if err := json.Unmarshal(b, v); err != nil {
		return false
	}

	return true
}

/*
Example:

// editor leaf 側
type FileState struct {
	Line    int
	Column  int
	ScrollY int
}

// save:

AppState().Save(
	"editor:"+path,
	FileState{
		Line: line,
		Column: col,
	},
)

// load:

var st FileState

if AppState().Load(
	"editor:"+path,
	&st,
) {
	restore(st)
}
*/
