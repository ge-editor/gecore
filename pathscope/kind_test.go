package pathscope

import "testing"

func TestParseKind(t *testing.T) {
	cases := []struct {
		in      string
		want    Kind
		wantErr bool
	}{
		{"all", All, false},
		{"workspaces", Workspaces, false},
		{"current", Current, false},
		{"a", All, false},
		{"w", Workspaces, false},
		{"c", Current, false},
		{"cur", Current, false},
		{"  Current ", Current, false},
		{"WORK", Workspaces, false},
		{"", "", true},
		{"   ", "", true},
		{"x", "", true},
		{"workspace-x", "", true},
	}
	for _, c := range cases {
		got, err := ParseKind(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseKind(%q) error = %v, wantErr %v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("ParseKind(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestComplete(t *testing.T) {
	cases := []struct {
		in     string
		want   Kind
		wantOK bool
	}{
		{"", "", false}, // every kind matches
		{"a", All, true},
		{"al", All, true},
		{"w", Workspaces, true},
		{"cu", Current, true},
		{"current", Current, true},
		{"x", "", false},
	}
	for _, c := range cases {
		got, ok := Complete(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("Complete(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

func TestKindValid(t *testing.T) {
	for _, k := range Kinds() {
		if !k.Valid() {
			t.Errorf("%q should be valid", k)
		}
	}
	for _, k := range []Kind{"", "none", "work", "works", "ALL"} {
		if k.Valid() {
			t.Errorf("%q should not be valid", k)
		}
	}
}
