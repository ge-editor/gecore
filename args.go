package gecore

import (
	"io"
	"os"
	"strconv"
	"strings"
)

// Arguments groups
var (
	WorkSpaces []string // directory
	Switches   []string // prefix -
	Files      []string

	// WorkFiles are the workspace files given with "--work FILE" or
	// "--work=FILE" (repeatable). By convention they end in ".ge-workspace",
	// but the extension is not checked. DefaultWork is set by
	// "--default-work": restore the default workspace saved in the state
	// directory. ArgErrors collects option mistakes (for example "--work"
	// without a value); the editor reports them once it can show messages.
	WorkFiles   []string
	DefaultWork bool
	ArgErrors   []string

	// FilePositions[i] is where the cursor starts in Files[i], from
	// "ge +LINE main.go" or "ge +LINE:COLUMN main.go". Same length as Files;
	// the zero value means "not specified". StdinPosition is the same for
	// the stdin buffer ("cat f | ge +LINE", "ge +LINE -").
	FilePositions []FilePosition
	StdinPosition FilePosition

	// Standard input. When stdin is not a terminal ("cat file | ge",
	// "ge < file") and no file is given, or when the argument "-" is given
	// and stdin is not a terminal, init reads it completely into Stdin and
	// sets StdinRead (an empty input still gives an empty buffer).
	// StdinErr is the read error, if any; Stdin then holds what was read.
	// The editor turns it into the first buffer and clears Stdin.
	StdinRead bool
	Stdin     []byte
	StdinErr  error
)

// FilePosition is a cursor position given on the command line. Line and Col
// are 1-based (Col counts characters); 0 means not specified.
type FilePosition struct {
	Line, Col int
}

// parsePosition parses "+LINE" and "+LINE:COLUMN" (both >= 1).
func parsePosition(arg string) (FilePosition, bool) {
	s, ok := strings.CutPrefix(arg, "+")
	if !ok {
		return FilePosition{}, false
	}
	lineStr, colStr, hasCol := strings.Cut(s, ":")
	line, err := strconv.Atoi(lineStr)
	if err != nil || line < 1 {
		return FilePosition{}, false
	}
	pos := FilePosition{Line: line}
	if hasCol {
		col, err := strconv.Atoi(colStr)
		if err != nil || col < 1 {
			return FilePosition{}, false
		}
		pos.Col = col
	}
	return pos, true
}

func init() {
	// Parese arguments
	explicitStdin := false
	var pending FilePosition // "+LINE[:COLUMN]" applies to the next file (or "-")
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--default-work" {
			DefaultWork = true
		} else if arg == "--work" {
			// The next argument is the workspace file, even if it looks like
			// an option or a directory.
			if i+1 < len(os.Args) {
				i++
				WorkFiles = append(WorkFiles, os.Args[i])
			} else {
				ArgErrors = append(ArgErrors, "--work requires a workspace file")
			}
		} else if v, ok := strings.CutPrefix(arg, "--work="); ok {
			if v == "" {
				ArgErrors = append(ArgErrors, "--work= requires a workspace file")
			} else {
				WorkFiles = append(WorkFiles, v)
			}
		} else if pos, ok := parsePosition(arg); ok {
			pending = pos
		} else if arg == "-" {
			explicitStdin = true
			StdinPosition, pending = pending, FilePosition{}
		} else if info, err := os.Stat(arg); err == nil && info.IsDir() {
			WorkSpaces = append(WorkSpaces, arg)
		} else if strings.HasPrefix(arg, "-") {
			Switches = append(Switches, arg)
		} else {
			Files = append(Files, arg)
			FilePositions = append(FilePositions, pending)
			pending = FilePosition{}
		}
	}
	if StdinPosition.Line == 0 {
		StdinPosition = pending // trailing "+LINE": for the stdin buffer
	}

	if (explicitStdin || len(Files) == 0) && stdinIsNotTerminal() {
		StdinRead = true
		Stdin, StdinErr = io.ReadAll(os.Stdin)
	}
}

// stdinIsNotTerminal reports whether stdin is a pipe or a file. A terminal
// (and /dev/null) is a character device, so an interactive start never blocks
// here.
func stdinIsNotTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}
