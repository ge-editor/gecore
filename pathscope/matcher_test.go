package pathscope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ge-editor/utils"
)

func mkdirs(t *testing.T, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMatchAll(t *testing.T) {
	m := New(All, nil, nil)
	if !m.Match("", "/anything") || !m.Match("id", "") || !m.Match("", "") {
		t.Error("All must match everything")
	}
	// A kind that is not valid behaves like All.
	if m := New("bogus", []string{"x"}, nil); m.Kind() != All || !m.Match("", "/x") {
		t.Error("an invalid kind must behave like All")
	}
}

func TestMatchByID(t *testing.T) {
	m := New(Current, []string{"id-1"}, nil)
	if !m.Match("id-1", "/nowhere/x.go") {
		t.Error("the ID of the scope must match whatever the path is")
	}
	if m.Match("id-2", "/nowhere/x.go") {
		t.Error("another ID must not match")
	}
	if m.Match("", "/nowhere/x.go") {
		t.Error("no ID and no root must not match")
	}
	if m.Match("id-1x", "") {
		t.Error("IDs are compared as a whole")
	}
}

func TestMatchByPath(t *testing.T) {
	base := t.TempDir()
	foo := filepath.Join(base, "foo")
	foobar := filepath.Join(base, "foobar")
	mkdirs(t, filepath.Join(foo, "sub"), foobar)

	m := New(Current, nil, []string{foo})
	cases := []struct {
		path string
		want bool
	}{
		{foo, true},                                   // the root itself
		{filepath.Join(foo, "a.go"), true},            // inside
		{filepath.Join(foo, "sub", "b.go"), true},     // nested
		{filepath.Join(foo, "new", "dir", "c.go"), true}, // does not exist yet
		{filepath.Join(foobar, "a.go"), false},        // same prefix, other directory
		{filepath.Join(base, "x.go"), false},          // parent
		{"", false},
	}
	for _, c := range cases {
		if got := m.Match("", c.path); got != c.want {
			t.Errorf("Match(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestMatchIDOrPath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "proj")
	mkdirs(t, root)

	m := New(Workspaces, []string{"a"}, []string{root})
	if !m.Match("b", filepath.Join(root, "x.go")) {
		t.Error("another workspace's entry under the root must match by path")
	}
	if !m.Match("a", filepath.Join(base, "elsewhere", "x.go")) {
		t.Error("an entry of the workspace outside its directories must match by ID")
	}
	if m.Match("b", filepath.Join(base, "elsewhere", "x.go")) {
		t.Error("neither the ID nor the path matches")
	}
}

func TestMatchEmptyDirectories(t *testing.T) {
	// A scope without directories uses the ID only.
	m := New(Current, []string{"id"}, nil)
	if !m.Match("id", "/x") || m.Match("other", "/x") {
		t.Error("an empty directory list must fall back to the ID only")
	}
}

func TestMatchRootDoesNotExist(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "nope")
	m := New(Current, nil, []string{root})
	if !m.Match("", filepath.Join(root, "x.go")) {
		t.Error("a missing root must still match lexically")
	}
}

func TestMatchSymlink(t *testing.T) {
	base := t.TempDir()
	realDir := filepath.Join(base, "real")
	mkdirs(t, realDir)
	link := filepath.Join(base, "link")
	if err := os.Symlink(realDir, link); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}

	// The path goes through a link, the root is the real directory.
	m := New(Current, nil, []string{realDir})
	if !m.Match("", filepath.Join(link, "a.go")) {
		t.Error("a path through a symbolic link must match the real root")
	}
	// The root is the link, the path is real.
	m = New(Current, nil, []string{link})
	if !m.Match("", filepath.Join(realDir, "a.go")) {
		t.Error("a real path must match a root that is a symbolic link")
	}
	// Asking twice uses the cache and gives the same answer.
	if !m.Match("", filepath.Join(realDir, "a.go")) {
		t.Error("cached answer differs")
	}
}

func TestMatchNestedAndDuplicateRoots(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "p")
	child := filepath.Join(parent, "c")
	mkdirs(t, child)

	m := New(Workspaces, nil, []string{parent, child, parent, ""})
	if !m.Match("", filepath.Join(child, "x.go")) || !m.Match("", filepath.Join(parent, "y.go")) {
		t.Error("nested roots are a union")
	}
	if len(m.roots) != 2 {
		t.Errorf("duplicate roots must be dropped: %d roots", len(m.roots))
	}
}

func TestIsUnderFold(t *testing.T) {
	if !isUnder("/A/Foo", "/a/foo/x.go", true) {
		t.Error("fold: case must be ignored")
	}
	if isUnder("/A/Foo", "/a/foo/x.go", false) {
		t.Error("no fold: case must matter")
	}
	if isUnder("/a/foo", "/a/foobar/x.go", true) {
		t.Error("whole elements only")
	}
}

func TestMatchCaseFollowsFileSystem(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "Proj")
	mkdirs(t, filepath.Join(root, "sub"))

	m := New(Current, nil, []string{root})
	got := m.Match("", filepath.Join(base, "PROJ", "a.go"))
	if want := utils.IsCaseInsensitivePath(root); got != want {
		t.Errorf("a path with another case: Match = %v, but the file system ignores case = %v", got, want)
	}
	// The same case always matches.
	if !m.Match("", filepath.Join(root, "a.go")) {
		t.Error("the same case must match")
	}
}
