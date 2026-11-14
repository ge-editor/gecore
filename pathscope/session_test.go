package pathscope

import "testing"

func TestSession(t *testing.T) {
	var asked []Kind
	resolve := func(k Kind) *Matcher {
		asked = append(asked, k)
		return New(k, []string{"id-" + string(k)}, nil)
	}
	s := NewSession(Current, resolve)
	if s.Kind() != Current || !s.Matcher().Match("id-current", "") {
		t.Fatal("the session must start with the given kind")
	}
	s.Set(Workspaces)
	if s.Kind() != Workspaces || !s.Matcher().Match("id-workspaces", "") || s.Matcher().Match("id-current", "") {
		t.Fatal("Set must build the Matcher of the new kind")
	}
	s.Set("bogus") // ignored
	if s.Kind() != Workspaces {
		t.Fatal("an invalid kind must be ignored")
	}
	s.Set(All)
	if !s.Matcher().Match("x", "/y") {
		t.Fatal("All must match everything")
	}
	if len(asked) != 3 {
		t.Errorf("resolver calls = %v", asked)
	}
}

func TestSessionWithoutResolver(t *testing.T) {
	s := NewSession("bogus", nil)
	if s.Kind() != All || !s.Matcher().Match("", "/x") {
		t.Error("an invalid initial kind must start as All")
	}
	s.Set(Current)
	if s.Kind() != Current || s.Matcher() == nil || s.Matcher().Match("id", "/x") {
		t.Error("without a resolver the scope has no workspaces")
	}
}
