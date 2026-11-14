package state

import (
	"encoding/json"
	"testing"
	"time"
)

func paths(list []RecentFile) []string {
	out := make([]string, len(list))
	for i, f := range list {
		out[i] = f.Path
	}
	return out
}

func TestRecentOrderAndDedup(t *testing.T) {
	s := newState()
	t0 := time.Unix(1000, 0)
	s.recordRecent("/a", t0)
	s.recordRecent("/b", t0.Add(time.Second))
	s.recordRecent("/c", t0.Add(2*time.Second))
	s.recordRecent("/a", t0.Add(3*time.Second)) // opened again: moves to the front

	got := s.recentFiles()
	want := []string{"/a", "/c", "/b"}
	if len(got) != 3 {
		t.Fatalf("got %v", paths(got))
	}
	for i := range want {
		if got[i].Path != want[i] {
			t.Errorf("got %v, want %v", paths(got), want)
			break
		}
	}
	if got[0].OpenedAt != t0.Add(3*time.Second).Unix() {
		t.Errorf("OpenedAt = %d", got[0].OpenedAt)
	}
}

func TestRecentLimit(t *testing.T) {
	s := newState()
	for i := 0; i < recentLimit+50; i++ {
		s.recordRecent("/f"+string(rune('a'+i%26))+string(rune('a'+i/26)), time.Unix(int64(i), 0))
	}
	if n := len(s.recentFiles()); n != recentLimit {
		t.Errorf("len = %d, want %d", n, recentLimit)
	}
}

func TestRecentLegacyEntriesComeLast(t *testing.T) {
	s := newState()
	// Cursor entries of files opened before the list existed.
	s.Items["editor:/old/b"] = json.RawMessage(`{}`)
	s.Items["editor:/old/a"] = json.RawMessage(`{}`)
	s.Items["editor:/new"] = json.RawMessage(`{}`)
	s.Items["unrelated"] = json.RawMessage(`{}`)

	s.recordRecent("/new", time.Unix(1, 0))

	got := s.recentFiles()
	want := []string{"/new", "/old/a", "/old/b"} // listed first, then by path
	if len(got) != len(want) {
		t.Fatalf("got %v", paths(got))
	}
	for i := range want {
		if got[i].Path != want[i] {
			t.Fatalf("got %v, want %v", paths(got), want)
		}
	}
	if got[1].OpenedAt != 0 {
		t.Errorf("a legacy entry has no time, got %d", got[1].OpenedAt)
	}
}

func TestRecentEmptyPathIsIgnored(t *testing.T) {
	s := newState()
	s.recordRecent("", time.Now())
	if len(s.recentFiles()) != 0 {
		t.Error("an empty path must not be recorded")
	}
}
