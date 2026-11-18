package remote

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestOperation(t *testing.T) {
	cases := []struct {
		name string
		o    Options
		want string
		bad  bool
	}{
		{"file", Options{Files: []FileSpec{{Path: "/a"}}}, OpOpen, false},
		{"switch work", Options{Work: "w"}, OpSwitch, false},
		{"switch root tree", Options{RootTree: "main"}, OpSwitch, false},
		{"nothing", Options{}, "", true},
		{"get", Options{GetBuffer: true}, OpGetBuffer, false},
		{"get two files", Options{GetBuffer: true, Files: []FileSpec{{Path: "/a"}, {Path: "/b"}}}, "", true},
		{"new with file", Options{NewBufferStdin: true, Files: []FileSpec{{Path: "/a"}}}, "", true},
		{"two ops", Options{GetBuffer: true, InsertStdin: true}, "", true},
		{"list", Options{ListInstances: true}, OpList, false},
		{"list with work", Options{ListInstances: true, Work: "w"}, "", true},
	}
	for _, c := range cases {
		got, err := c.o.operation()
		if (err != nil) != c.bad || got != c.want {
			t.Errorf("%s: got (%q, %v), want (%q, bad=%v)", c.name, got, err, c.want, c.bad)
		}
	}
}

func TestParseSocketName(t *testing.T) {
	if pid, ok := parseSocketName("ge-12345.sock"); !ok || pid != 12345 {
		t.Errorf("got %d %v", pid, ok)
	}
	for _, n := range []string{"ge-.sock", "ge-0.sock", "ge-x.sock", "other.sock", "ge-1"} {
		if _, ok := parseSocketName(n); ok {
			t.Errorf("%q accepted", n)
		}
	}
}

func TestExitCodeValues(t *testing.T) {
	// The numbers are a CLI contract.
	if ExitInstanceNotFound != 3 || ExitWorkspaceNotFound != 4 || ExitBufferNotFound != 6 ||
		ExitConnectionError != 8 || ExitTimeout != 10 {
		t.Fatal("exit code numbering changed")
	}
}

func serve(t *testing.T, h HandlerFunc) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	s, err := Listen(os.Getpid(), h)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
}

func TestRunGetBuffer(t *testing.T) {
	serve(t, func(r Request) Response {
		if r.Op != OpGetBuffer {
			return Response{Code: ExitUnsupportedOperation, Error: "unsupported"}
		}
		return Response{OK: true, Data: []byte("package main\n")}
	})
	var out, errb bytes.Buffer
	code := Run(Options{Enabled: true, GetBuffer: true}, strings.NewReader(""), &out, &errb)
	if code != ExitSuccess || out.String() != "package main\n" || errb.Len() != 0 {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errb.String())
	}
}

func TestRunErrorGoesToStderr(t *testing.T) {
	serve(t, func(r Request) Response {
		return Response{Code: ExitBufferNotFound, Error: "active leaf has no buffer"}
	})
	var out, errb bytes.Buffer
	code := Run(Options{Enabled: true, GetBuffer: true}, strings.NewReader(""), &out, &errb)
	if code != ExitBufferNotFound || out.Len() != 0 || !strings.Contains(errb.String(), "no buffer") {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errb.String())
	}
}

func TestRunSendsStdin(t *testing.T) {
	var got Request
	serve(t, func(r Request) Response { got = r; return Response{OK: true} })
	var out, errb bytes.Buffer
	code := Run(Options{Enabled: true, InsertStdin: true}, strings.NewReader("x := 1\n"), &out, &errb)
	if code != ExitSuccess || got.Op != OpInsert || string(got.Data) != "x := 1\n" {
		t.Fatalf("code=%d req=%+v err=%q", code, got, errb.String())
	}
}

func TestRunNoInstance(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	var out, errb bytes.Buffer
	if code := Run(Options{Enabled: true, GetBuffer: true}, strings.NewReader(""), &out, &errb); code != ExitInstanceNotFound {
		t.Fatalf("code=%d", code)
	}
}

func TestSelectInstanceLargestPID(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	// Only this process has a live socket; the others are stale files.
	s, err := Listen(os.Getpid(), HandlerFunc(func(Request) Response { return Response{OK: true} }))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if pid, ok := SelectInstance(0); !ok || pid != os.Getpid() {
		t.Fatalf("got %d %v", pid, ok)
	}
	if _, ok := SelectInstance(os.Getpid() + 1); ok {
		t.Fatal("a PID without a live instance must not be selected")
	}
}

func TestParseArgs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	o := ParseArgs([]string{"--remote=42", "--work", "proj", "--root-tree=main", "+120:8", "a.go", "b.go"})
	if len(o.Errors) != 0 || o.PID != 42 || o.Work != "proj" || o.RootTree != "main" {
		t.Fatalf("%+v", o)
	}
	if len(o.Files) != 2 || o.Files[0].Line != 120 || o.Files[0].Col != 8 || o.Files[1].Line != 0 {
		t.Fatalf("files: %+v", o.Files)
	}
	if !strings.HasSuffix(o.Files[0].Path, "a.go") || o.Files[0].Path[0] != '/' {
		t.Fatalf("path not absolute: %q", o.Files[0].Path)
	}

	for _, bad := range [][]string{
		{"--bogus"},
		{"--remote=0"},
		{"--work"},
		{"+3"},
		{dir},
	} {
		if o := ParseArgs(bad); len(o.Errors) == 0 {
			t.Errorf("%v: no error", bad)
		}
	}
	if o := ParseArgs([]string{"--get-buffer"}); !o.GetBuffer || len(o.Errors) != 0 {
		t.Errorf("%+v", o)
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run(ParseArgs([]string{"--help"}), strings.NewReader(""), &out, &errb); code != ExitSuccess || !strings.Contains(out.String(), "usage:") {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}
