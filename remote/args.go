package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Usage is the help text of ge-remote.
const Usage = `ge-remote - control a running ge

usage:
  ge-remote [options] [+LINE[:COL]] [file ...]
  ge-remote --list-instances
  ge-remote --get-buffer [file]
  ge-remote --new-buffer-stdin
  ge-remote --replace-buffer-stdin [file]
  ge-remote --insert-stdin [file]

options:
  --remote=PID              target the ge with this process ID
                            (default: the running ge with the largest PID)
  --list-instances          list running ge instances, workspaces and RootTrees
  --work NAME               use this workspace (created if missing)
  --root-tree NAME          use this RootTree (created if missing)
  --get-buffer              write the buffer to standard output
  --new-buffer-stdin        make a new buffer from standard input
  --replace-buffer-stdin    replace the whole buffer with standard input
  --insert-stdin            insert standard input at the cursor
  -h, --help                show this help

Files are opened in ge, at +LINE[:COL] when given. With only --work and/or
--root-tree, ge switches to it. Without a file, a buffer operation works on
the active leaf's buffer of the target workspace and RootTree.

Buffer data goes to standard output; messages go to standard error.

exit status:
   0 success                 6 buffer not found
   1 error                   7 unsupported operation
   2 invalid argument        8 connection error
   3 instance not found      9 operation failed in ge
   4 workspace not found    10 timeout
   5 RootTree not found

examples:
  ge-remote +120:8 main.go
  ge-remote --work myproj --root-tree main main.go
  ge-remote --get-buffer > out.go
  cat gen.go | ge-remote --replace-buffer-stdin
`

// parsePosition parses "+LINE" and "+LINE:COLUMN" (both >= 1).
func parsePosition(arg string) (line, col int, ok bool) {
	s, ok := strings.CutPrefix(arg, "+")
	if !ok {
		return 0, 0, false
	}
	lineStr, colStr, hasCol := strings.Cut(s, ":")
	line, err := strconv.Atoi(lineStr)
	if err != nil || line < 1 {
		return 0, 0, false
	}
	if hasCol {
		col, err = strconv.Atoi(colStr)
		if err != nil || col < 1 {
			return 0, 0, false
		}
	}
	return line, col, true
}

// ParseArgs parses the command line of ge-remote; args does not include the
// program name. "--work" and "--root-tree" take a workspace / RootTree
// name (not a workspace file), and files are made absolute.
// Mistakes are collected in Options.Errors.
func ParseArgs(args []string) Options {
	o := Options{Enabled: true}
	bad := func(msg string) { o.Errors = append(o.Errors, msg) }
	var pendLine, pendCol int // "+LINE[:COL]" applies to the next file

	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, val, hasVal := strings.Cut(arg, "=")
		switch name {
		case "-h", "--help":
			o.Help = true
		case "--remote":
			if hasVal {
				pid, err := strconv.Atoi(val)
				if err != nil || pid < 1 {
					bad("--remote=PID needs a positive process ID: " + arg)
				} else {
					o.PID = pid
				}
			}
		case "--list-instances":
			o.ListInstances = true
		case "--get-buffer":
			o.GetBuffer = true
		case "--new-buffer-stdin":
			o.NewBufferStdin = true
		case "--replace-buffer-stdin":
			o.ReplaceBufferStdin = true
		case "--insert-stdin":
			o.InsertStdin = true
		case "--work", "--root-tree":
			if !hasVal {
				if i+1 >= len(args) {
					bad(name + " requires a name")
					continue
				}
				i++
				val = args[i]
			}
			if val == "" {
				bad(name + " requires a name")
				continue
			}
			if name == "--work" {
				o.Work = val
			} else {
				o.RootTree = val
			}
		default:
			if line, col, ok := parsePosition(arg); ok {
				pendLine, pendCol = line, col
				continue
			}
			if strings.HasPrefix(arg, "-") {
				bad("unknown option: " + arg)
				continue
			}
			if info, err := os.Stat(arg); err == nil && info.IsDir() {
				bad("a directory cannot be opened remotely: " + arg)
				continue
			}
			abs, err := filepath.Abs(arg)
			if err != nil {
				bad(err.Error())
				continue
			}
			o.Files = append(o.Files, FileSpec{Path: abs, Line: pendLine, Col: pendCol})
			pendLine, pendCol = 0, 0
		}
	}
	if pendLine != 0 {
		bad(fmt.Sprintf("+%d needs a file after it", pendLine))
	}
	return o
}
