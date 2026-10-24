package excommand

import (
	"fmt"
	"slices"
	"strings"
)

// Prompt is the state machine behind the ESC x prompt. It knows nothing about
// the minibuffer or the popup menu: the UI feeds it the input text (Update)
// and the Enter key (Enter), shows the items Update returns, and carries out
// the Action Enter returns. The only state is what the input text implies; a
// selection is never remembered, it is written into the input.
//
// States
//
//	Command phase    the input is not "<known command> ...": the popup lists
//	                 command names (candy).
//	Argument phase   the input is "<known command> <argument>": the popup lists
//	                 the argument candidates of that command (argItems; empty
//	                 for commands without any).
//
// Transitions
//
//	Update(text)   decides the phase from the text alone, in both directions.
//	Enter, command phase
//	  a candidate is selected and the input is not exactly its name
//	                                            -> ActionFill  (its name, plus a
//	                                               space if it takes an argument)
//	  otherwise the input is taken as a command name (the selected candidate's
//	  name, or the trimmed input if nothing is selected):
//	    not a command name                      -> ActionError
//	    takes an argument                       -> ActionFill  ("name ")
//	    otherwise                               -> ActionRun
//	Enter, argument phase                       -> ActionRun   (selected argument
//	                                               substituted into the input)
//
// After ActionFill the UI calls Update with the new text, which enters the
// argument phase for a command that takes an argument. ActionRun clears the
// argument state and records the input in the history (see SetHistory); the
// UI then closes the prompt.
//
// History (optional, SetHistory)
//
//	HistoryPrev / HistoryNext walk the inputs that were run before, newest
//	first, limited to commands that exist in the current context. While a
//	walk is in progress the popup lists those inputs (Update returns them,
//	PopupIndex is the one shown), Enter on another popup item writes that
//	item into the input (ActionFill), and HistoryDelete removes an item from
//	the history. The walk
//	starts at the first HistoryPrev (the text then is kept as the draft and
//	comes back when HistoryNext passes the newest entry) and ends as soon as
//	Update sees a text other than the one the walk produced, i.e. when the
//	user edits it.
type Prompt struct {
	commands func() []ExCommand

	// Command phase.
	candy []Candidate

	// Argument phase.
	argCmd   ExCommand
	argItems []string

	// History.
	history *History
	hopts   HistoryOptions
	nav     historyNav
}

// HistoryOptions connects a Prompt to the editor's context.
type HistoryOptions struct {
	// Applicable returns the owners (see HistoryEntry) whose commands exist
	// in the current context: "" for the global commands plus the leaf types
	// from the focused leaf toward its parents. Only their entries are
	// offered by HistoryPrev / HistoryNext.
	Applicable func() []string

	// OwnerOf returns the owner of the command with this name, as
	// commands() resolves it: the leaf type that provides it, or "" if it
	// is global.
	OwnerOf func(name string) string

	// OnError, if set, receives errors from saving the history. The input is
	// kept in memory anyway.
	OnError func(error)
}

// historyNav is the state of a walk through the history.
type historyNav struct {
	active bool
	items  []string // snapshot taken when the walk started
	index  int      // position in items
	draft  string   // text before the walk
	shown  string   // text the walk put into the input
}

// NewPrompt creates a Prompt. commands returns the ordered command list (see
// Find); it is called on every Update and Enter, so it always reflects the
// current state of the editor.
func NewPrompt(commands func() []ExCommand) *Prompt {
	return &Prompt{commands: commands}
}

// SetHistory enables the history: ActionRun records the input in h, and
// HistoryPrev / HistoryNext walk through it.
func (p *Prompt) SetHistory(h *History, opts HistoryOptions) {
	p.history = h
	p.hopts = opts
	p.nav = historyNav{}
}

// HistoryPrev returns the next older input of the history for the input text
// text. ok is false if there is none (no history, nothing applicable, or the
// oldest entry is already shown); the input then stays as it is.
func (p *Prompt) HistoryPrev(text string) (prev string, ok bool) {
	if p.history == nil {
		return "", false
	}
	if !p.nav.active {
		var owners []string
		if p.hopts.Applicable != nil {
			owners = p.hopts.Applicable()
		}
		items := p.history.Inputs(func(owner string) bool {
			return slices.Contains(owners, owner)
		})
		if len(items) == 0 {
			return "", false
		}
		p.nav = historyNav{active: true, items: items, index: -1, draft: text}
	}
	if p.nav.index+1 >= len(p.nav.items) {
		return "", false
	}
	p.nav.index++
	p.nav.shown = p.nav.items[p.nav.index]
	return p.nav.shown, true
}

// HistoryNext returns the next newer input, or the draft from before the walk
// after the newest one. ok is false if no walk is in progress.
func (p *Prompt) HistoryNext() (next string, ok bool) {
	if !p.nav.active {
		return "", false
	}
	p.nav.index--
	if p.nav.index < 0 {
		draft := p.nav.draft
		p.nav = historyNav{}
		return draft, true
	}
	p.nav.shown = p.nav.items[p.nav.index]
	return p.nav.shown, true
}

// record writes input, which is about to be run, to the history and ends any
// walk.
func (p *Prompt) record(input string) {
	p.nav = historyNav{}
	if p.history == nil {
		return
	}
	name, _, _ := strings.Cut(strings.TrimSpace(input), " ")
	owner := ""
	if p.hopts.OwnerOf != nil {
		owner = p.hopts.OwnerOf(name)
	}
	if err := p.history.Add(owner, input); err != nil && p.hopts.OnError != nil {
		p.hopts.OnError(err)
	}
}

// HistoryWalking reports whether a walk through the history is in progress,
// i.e. whether the popup lists history inputs.
func (p *Prompt) HistoryWalking() bool {
	return p.nav.active
}

// PopupIndex is the popup item that corresponds to the input during a walk
// through the history, or -1 when the popup lists commands or arguments (keep
// the popup's own selection then).
func (p *Prompt) PopupIndex() int {
	if !p.nav.active {
		return -1
	}
	return p.nav.index
}

// HistoryDelete removes the item at index of the history popup from the
// history (from every applicable owner; see HistoryOptions.Applicable) and
// returns the input to show next: the entry that took its place, the previous
// one if it was the last, or the draft if none is left. ok is false when no
// walk is in progress or index is out of range.
func (p *Prompt) HistoryDelete(index int) (text string, ok bool) {
	if !p.nav.active || index < 0 || index >= len(p.nav.items) {
		return "", false
	}

	input := p.nav.items[index]
	var owners []string
	if p.hopts.Applicable != nil {
		owners = p.hopts.Applicable()
	}
	err := p.history.Remove(input, func(owner string) bool {
		return slices.Contains(owners, owner)
	})
	if err != nil && p.hopts.OnError != nil {
		p.hopts.OnError(err)
	}

	p.nav.items = slices.Delete(p.nav.items, index, index+1)
	if len(p.nav.items) == 0 {
		draft := p.nav.draft
		p.nav = historyNav{}
		return draft, true
	}
	p.nav.index = min(index, len(p.nav.items)-1)
	p.nav.shown = p.nav.items[p.nav.index]
	return p.nav.shown, true
}

// Update recomputes the phase and the popup items for the input text. During
// a walk through the history the items are the history inputs.
func (p *Prompt) Update(text string) []string {
	if p.nav.active && text != p.nav.shown {
		p.nav = historyNav{} // the user edited the text: the walk is over
	}

	items := p.complete(text)
	if p.nav.active {
		return append([]string{}, p.nav.items...)
	}
	return items
}

// complete decides the phase for text and returns the command or argument
// candidates for the popup.
func (p *Prompt) complete(text string) []string {
	if name, rest, found := strings.Cut(text, " "); found {
		if c := Find(p.commands(), name); c != nil {
			p.candy = nil
			p.argCmd = c
			p.argItems = nil
			if ac, ok := c.(ArgCompleter); ok {
				rest = strings.TrimLeft(rest, " ")
				p.argItems = ac.CompleteArgs(rest)
				if len(p.argItems) == 1 && strings.EqualFold(p.argItems[0], rest) {
					// Exactly typed: hide the popup so Enter runs the input as is.
					p.argItems = nil
				}
			}
			// Commands without argument candidates just show an empty popup;
			// Enter runs them with the raw input.
			return append([]string{}, p.argItems...)
		}
	}
	p.argCmd, p.argItems = nil, nil

	p.candy = Complete(p.commands(), text)

	items := []string{}
	for _, c := range p.candy {
		items = append(items, c.Value+" ("+c.Command.Description()+")")
	}
	if len(items) == 1 && text == p.candy[0].Value {
		items = []string{}
	}
	return items
}

// ActionKind tells the UI what to do after Enter.
type ActionKind int

const (
	// ActionFill: replace the input with Action.Text, put the cursor at its
	// end, then call Update with it and show the items.
	ActionFill ActionKind = iota
	// ActionRun: execute Action.Command with Action.Input as its only
	// argument, then close the prompt.
	ActionRun
	// ActionError: show Action.Message.
	ActionError
)

// Action is the result of Enter.
type Action struct {
	Kind ActionKind

	Text string // ActionFill

	Command      ExCommand // ActionRun
	Input        string    // ActionRun: the raw input, command name included
	FromArgument bool      // ActionRun: run from the argument phase

	Message string // ActionError
}

// Enter handles the Enter key. text is the current input; index is the
// selected popup item (-1 if none, or if the popup is empty).
func (p *Prompt) Enter(text string, index int) Action {
	if p.nav.active {
		// The popup lists history inputs, not candidates.
		if index >= 0 && index < len(p.nav.items) && index != p.nav.index {
			// Another item was selected: take it over; Enter again runs it.
			return Action{Kind: ActionFill, Text: p.nav.items[index]}
		}
		index = -1
	}
	if p.argCmd != nil {
		input := text
		if index >= 0 && index < len(p.argItems) {
			name, _, _ := strings.Cut(text, " ")
			input = name + " " + p.argItems[index]
		}
		c := p.argCmd
		p.argCmd, p.argItems = nil, nil
		p.record(input)
		return Action{Kind: ActionRun, Command: c, Input: input, FromArgument: true}
	}

	var name string
	var c ExCommand
	if index >= 0 && index < len(p.candy) {
		name, c = p.candy[index].Value, p.candy[index].Command
		if name != text {
			return Action{Kind: ActionFill, Text: name + argSeparator(c)}
		}
	} else {
		name = strings.TrimSpace(text)
		c = Find(p.commands(), name)
		if c == nil {
			return Action{
				Kind:    ActionError,
				Message: fmt.Sprintf("Out of []Candidate index: %d, candy: %v", index, p.candy),
			}
		}
	}

	if _, ok := c.(ArgCompleter); ok {
		// Continue in the argument phase: "language " lists the languages.
		return Action{Kind: ActionFill, Text: name + " "}
	}
	p.record(text)
	return Action{Kind: ActionRun, Command: c, Input: text}
}

func argSeparator(c ExCommand) string {
	if _, ok := c.(ArgCompleter); ok {
		return " "
	}
	return ""
}
