package excommand

import (
	"context"
	"slices"
	"testing"
)

type promptCmd struct {
	name string
	args func(string) []string
}

func (c *promptCmd) Name() string        { return c.name }
func (c *promptCmd) Aliases() []string   { return nil }
func (c *promptCmd) Description() string { return "d" }
func (c *promptCmd) Execute(context.Context, []string) (bool, error) {
	return true, nil
}

type promptArgCmd struct{ *promptCmd }

func (c *promptArgCmd) CompleteArgs(input string) []string { return c.args(input) }

func newTestPrompt() (*Prompt, *promptCmd, *promptArgCmd) {
	plain := &promptCmd{name: "quit"}
	lang := &promptArgCmd{&promptCmd{name: "language", args: func(input string) []string {
		all := []string{"Go", "Markdown"}
		var result []string
		for _, item := range all {
			if input == "" || len(item) >= len(input) && item[:len(input)] == input {
				result = append(result, item)
			}
		}
		return result
	}}}
	p := NewPrompt(func() []ExCommand { return []ExCommand{lang, plain} })
	return p, plain, lang
}

func TestPromptCommandPhaseAndSelection(t *testing.T) {
	p, plain, _ := newTestPrompt()
	if got := p.Update("qu"); !slices.Equal(got, []string{"quit (d)"}) {
		t.Fatalf("Update(qu) = %v", got)
	}
	a := p.Enter("qu", 0)
	if a.Kind != ActionFill || a.Text != "quit" {
		t.Fatalf("Enter(candidate) = %+v", a)
	}
	p.Update("foo")
	if a := p.Enter("foo", -1); a.Kind != ActionError {
		t.Fatalf("Enter(foo) = %+v, want ActionError", a)
	}
	p.Update("quit")
	if a := p.Enter("quit", -1); a.Kind != ActionRun || a.Command != ExCommand(plain) {
		t.Fatalf("Enter(quit) = %+v", a)
	}
}

func TestPromptArgumentPhase(t *testing.T) {
	p, _, lang := newTestPrompt()
	if got := p.Update("lang"); !slices.Equal(got, []string{"language (d)"}) {
		t.Fatalf("Update(lang) = %v", got)
	}
	if a := p.Enter("lang", 0); a.Kind != ActionFill || a.Text != "language " {
		t.Fatalf("Enter(language candidate) = %+v", a)
	}
	if got := p.Update("language M"); !slices.Equal(got, []string{"Markdown"}) {
		t.Fatalf("Update(language M) = %v", got)
	}
	a := p.Enter("language M", 0)
	if a.Kind != ActionRun || a.Command != ExCommand(lang) || a.Input != "language Markdown" || !a.FromArgument {
		t.Fatalf("Enter(argument) = %+v", a)
	}
	if got := p.Update("language Go"); len(got) != 0 {
		t.Fatalf("Update(exact argument) = %v, want empty", got)
	}
	if a := p.Enter("language Go", -1); a.Kind != ActionRun || a.Input != "language Go" {
		t.Fatalf("Enter(exact argument) = %+v", a)
	}
}

func TestPromptExactNameAmongLongerCandidates(t *testing.T) {
	quit := &promptCmd{name: "quit"}
	quitAll := &promptCmd{name: "quit-all"}
	p := NewPrompt(func() []ExCommand { return []ExCommand{quit, quitAll} })
	if got := p.Update("quit"); len(got) != 2 {
		t.Fatalf("Update(quit) = %v, want both candidates", got)
	}
	index := -1
	for i, candidate := range p.candy {
		if candidate.Value == "quit" {
			index = i
		}
	}
	if a := p.Enter("quit", index); a.Kind != ActionRun || a.Command != ExCommand(quit) {
		t.Fatalf("Enter(exact name) = %+v", a)
	}
	for i, candidate := range p.candy {
		if candidate.Value == "quit-all" {
			index = i
		}
	}
	if a := p.Enter("quit", index); a.Kind != ActionFill || a.Text != "quit-all" {
		t.Fatalf("Enter(longer candidate) = %+v", a)
	}
}

func newHistoryPrompt() (*Prompt, *History) {
	p, _, _ := newTestPrompt()
	h, _ := LoadHistory("", 10)
	p.SetHistory(h, HistoryOptions{
		Applicable: func() []string { return []string{"", "editor"} },
		OwnerOf: func(name string) string {
			if name == "language" {
				return "editor"
			}
			return ""
		},
	})
	return p, h
}

func TestPromptRecordsRunInputsWithTheirOwner(t *testing.T) {
	p, h := newHistoryPrompt()
	p.Update("quit")
	p.Enter("quit", -1)
	p.Update("language Go")
	p.Enter("language Go", -1)
	if got := h.Inputs(func(owner string) bool { return owner == "" }); !slices.Equal(got, []string{"quit"}) {
		t.Fatalf("global history = %v", got)
	}
	if got := h.Inputs(func(owner string) bool { return owner == "editor" }); !slices.Equal(got, []string{"language Go"}) {
		t.Fatalf("editor history = %v", got)
	}
}

func TestPromptHistoryWalk(t *testing.T) {
	p, h := newHistoryPrompt()
	h.Add("", "quit")
	h.Add("editor", "language Go")
	h.Add("other", "foo")

	text, ok := p.HistoryPrev("draft")
	if !ok || text != "language Go" {
		t.Fatalf("Prev = %q, %v", text, ok)
	}
	p.Update(text)
	text, ok = p.HistoryPrev(text)
	if !ok || text != "quit" {
		t.Fatalf("Prev = %q, %v", text, ok)
	}
	p.Update(text)
	if _, ok = p.HistoryPrev(text); ok {
		t.Fatal("Prev past oldest entry should fail")
	}
	text, ok = p.HistoryNext()
	if !ok || text != "language Go" {
		t.Fatalf("Next = %q, %v", text, ok)
	}
	p.Update(text)
	text, ok = p.HistoryNext()
	if !ok || text != "draft" {
		t.Fatalf("Next to draft = %q, %v", text, ok)
	}
	p.Update(text)
	if _, ok = p.HistoryNext(); ok {
		t.Fatal("Next without active walk should fail")
	}
}

func TestPromptEditingEndsHistoryWalk(t *testing.T) {
	p, h := newHistoryPrompt()
	h.Add("", "quit")
	h.Add("editor", "language Go")
	text, _ := p.HistoryPrev("")
	p.Update(text)
	p.HistoryPrev(text)
	p.Update("language G")
	text, ok := p.HistoryPrev("language G")
	if !ok || text != "language Go" {
		t.Fatalf("Prev after editing = %q, %v", text, ok)
	}
	p.Update(text)
	text, ok = p.HistoryNext()
	if !ok || text != "language G" {
		t.Fatalf("Next = %q, %v", text, ok)
	}
}

func TestPromptHistoryWithoutHistory(t *testing.T) {
	p, _, _ := newTestPrompt()
	if _, ok := p.HistoryPrev(""); ok {
		t.Fatal("Prev without history should fail")
	}
	if _, ok := p.HistoryNext(); ok {
		t.Fatal("Next without history should fail")
	}
}

func TestPromptHistoryWalkListsHistoryInPopup(t *testing.T) {
	p, h := newHistoryPrompt()
	h.Add("", "quit")
	h.Add("editor", "language Go")

	if p.HistoryWalking() || p.PopupIndex() != -1 {
		t.Fatal("no history walk should be active")
	}
	text, _ := p.HistoryPrev("")
	items := p.Update(text)
	if !p.HistoryWalking() || !slices.Equal(items, []string{"language Go", "quit"}) || p.PopupIndex() != 0 {
		t.Fatalf("items = %v, index = %d", items, p.PopupIndex())
	}

	text, _ = p.HistoryPrev(text)
	p.Update(text)
	if p.PopupIndex() != 1 {
		t.Fatalf("PopupIndex = %d, want 1", p.PopupIndex())
	}
	items = p.Update("qu")
	if p.HistoryWalking() || !slices.Equal(items, []string{"quit (d)"}) {
		t.Fatalf("after editing: items = %v", items)
	}
}

func TestPromptEnterDuringHistoryWalk(t *testing.T) {
	p, h := newHistoryPrompt()
	h.Add("", "quit")
	h.Add("editor", "language Go")
	text, _ := p.HistoryPrev("")
	p.Update(text)

	a := p.Enter(text, 1)
	if a.Kind != ActionFill || a.Text != "quit" {
		t.Fatalf("Enter on another history item = %+v", a)
	}
	a = p.Enter(text, 0)
	if a.Kind != ActionRun || a.Input != "language Go" || !a.FromArgument {
		t.Fatalf("Enter on shown history item = %+v", a)
	}
	if p.HistoryWalking() {
		t.Fatal("running an input should end the history walk")
	}
}

func TestPromptHistoryDelete(t *testing.T) {
	p, h := newHistoryPrompt()
	h.Add("", "a")
	h.Add("", "b")
	h.Add("", "c")
	h.Add("other", "d")
	if _, ok := p.HistoryDelete(0); ok {
		t.Fatal("delete without a history walk should fail")
	}

	text, _ := p.HistoryPrev("draft")
	p.Update(text)
	text, _ = p.HistoryPrev(text)
	p.Update(text)
	text, ok := p.HistoryDelete(1)
	if !ok || text != "a" || p.PopupIndex() != 1 {
		t.Fatalf("delete b = %q, %v, index %d", text, ok, p.PopupIndex())
	}
	if got := h.Inputs(all); !slices.Equal(got, []string{"d", "c", "a"}) {
		t.Fatalf("history = %v", got)
	}
	text, ok = p.HistoryDelete(1)
	if !ok || text != "c" || p.PopupIndex() != 0 {
		t.Fatalf("delete a = %q, %v, index %d", text, ok, p.PopupIndex())
	}
	text, ok = p.HistoryDelete(0)
	if !ok || text != "draft" || p.HistoryWalking() {
		t.Fatalf("delete c = %q, %v, walking %v", text, ok, p.HistoryWalking())
	}
	if got := h.Inputs(all); !slices.Equal(got, []string{"d"}) {
		t.Fatalf("history = %v, want only other owner's entry", got)
	}
}
