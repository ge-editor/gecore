package excommand

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func all(string) bool { return true }

func TestHistoryAddDedupesAndOrdersNewestFirst(t *testing.T) {
	h, _ := LoadHistory("", 10)
	h.Add("", "quit")
	h.Add("leaf", "language Go")
	h.Add("", "quit") // moves to the front

	if got := h.Inputs(all); !slices.Equal(got, []string{"quit", "language Go"}) {
		t.Fatalf("Inputs = %v", got)
	}
	h.Add("", "   ") // blank is ignored
	if got := h.Inputs(all); len(got) != 2 {
		t.Fatalf("blank input was recorded: %v", got)
	}
}

func TestHistoryInputsFilterByOwner(t *testing.T) {
	h, _ := LoadHistory("", 10)
	h.Add("", "quit")
	h.Add("editor", "language Go")
	h.Add("other", "foo")

	got := h.Inputs(func(o string) bool { return o == "" || o == "editor" })
	if !slices.Equal(got, []string{"language Go", "quit"}) {
		t.Fatalf("Inputs = %v", got)
	}
}

func TestHistoryLimit(t *testing.T) {
	h, _ := LoadHistory("", 2)
	h.Add("", "a")
	h.Add("", "b")
	h.Add("", "c")
	if got := h.Inputs(all); !slices.Equal(got, []string{"c", "b"}) {
		t.Fatalf("Inputs = %v", got)
	}
}

func TestHistoryPersistsAcrossLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "excommand_history.json")

	h, err := LoadHistory(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Add("editor", "language Go"); err != nil {
		t.Fatal(err)
	}
	if err := h.Add("", "quit"); err != nil {
		t.Fatal(err)
	}

	h2, err := LoadHistory(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := h2.Inputs(all); !slices.Equal(got, []string{"quit", "language Go"}) {
		t.Fatalf("reloaded Inputs = %v", got)
	}
	// The owner is kept.
	if got := h2.Inputs(func(o string) bool { return o == "editor" }); !slices.Equal(got, []string{"language Go"}) {
		t.Fatalf("reloaded Inputs for owner editor = %v", got)
	}
}

func TestHistoryMergesEntriesOfOtherInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.json")

	a, _ := LoadHistory(path, 10)
	b, _ := LoadHistory(path, 10)

	a.Add("", "from a")
	b.Add("", "from b") // must not erase "from a"

	c, _ := LoadHistory(path, 10)
	if got := c.Inputs(all); !slices.Equal(got, []string{"from b", "from a"}) {
		t.Fatalf("merged Inputs = %v", got)
	}
}

func TestHistoryCorruptedFileStartsEmptyAndIsReplaced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := LoadHistory(path, 10)
	if err == nil {
		t.Fatal("expected an error for a corrupted file")
	}
	if err := h.Add("", "quit"); err != nil {
		t.Fatal(err)
	}
	h2, err := LoadHistory(path, 10)
	if err != nil || !slices.Equal(h2.Inputs(all), []string{"quit"}) {
		t.Fatalf("after Add: %v, %v", h2.Inputs(all), err)
	}
}

func TestHistoryRemoveOnlyApprovedOwnersAndPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.json")
	h, _ := LoadHistory(path, 10)
	h.Add("editor", "language Go")
	h.Add("other", "language Go") // same input, another leaf type
	h.Add("", "quit")

	if err := h.Remove("language Go", func(o string) bool { return o == "editor" }); err != nil {
		t.Fatal(err)
	}

	h2, _ := LoadHistory(path, 10)
	if got := h2.Inputs(func(o string) bool { return o == "editor" }); len(got) != 0 {
		t.Fatalf("editor entry still there: %v", got)
	}
	if got := h2.Inputs(func(o string) bool { return o == "other" }); !slices.Equal(got, []string{"language Go"}) {
		t.Fatalf("the other owner's entry must stay: %v", got)
	}
	if got := h2.Inputs(all); len(got) != 2 {
		t.Fatalf("Inputs = %v", got)
	}
}
