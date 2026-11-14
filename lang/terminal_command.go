package lang

// A TerminalCommand is an ESC x command that a mode offers while a buffer
// has that mode. The editor does not run it inside the editor screen: it
// leaves the screen like the "shell" command, runs the steps one after the
// other in the terminal, and returns to the editor afterwards.
//
// A mode only describes the command; running it (and the terminal) is the
// editor's work, so a mode package does not have to depend on the editor.
type TerminalCommand struct {
	// Name is the name typed after ESC x (kebab-case, like the others).
	Name        string
	Description string

	// Steps run in this order. A failed step does not stop the following
	// ones (a tool that cannot be installed must not block the others).
	Steps []TerminalStep

	// PauseOnSuccess waits for Enter before returning to the editor also
	// when every step succeeded, so the output can be read. When a step
	// failed the editor always waits. When false the editor returns at once
	// after success (exit).
	PauseOnSuccess bool
}

// TerminalStep is one program to run: Program with Args, no shell.
type TerminalStep struct {
	Program string
	Args    []string
}

// TerminalCommandProvider is implemented by a Mode that offers
// TerminalCommands. The editor looks for it with a type assertion.
type TerminalCommandProvider interface {
	TerminalCommands() []TerminalCommand
}
