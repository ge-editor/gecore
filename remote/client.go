package remote

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"text/tabwriter"
	"time"
)

// Options is what the command line asked for. gecore's argument parser fills
// it; Errors collects option mistakes.
type Options struct {
	Enabled bool // always true for ge-remote
	Help    bool // -h / --help
	PID     int  // --remote=PID; 0 = the live instance with the largest PID

	ListInstances      bool
	GetBuffer          bool
	NewBufferStdin     bool
	ReplaceBufferStdin bool
	InsertStdin        bool

	Work     string // --work NAME (means a workspace name, not a file, here)
	RootTree string // --root-tree NAME
	Files    []FileSpec

	Errors []string
}

// callTimeout bounds one request/response exchange.
const callTimeout = 10 * time.Second

// operation decides which request the options describe.
func (o Options) operation() (string, error) {
	var ops []string
	if o.ListInstances {
		ops = append(ops, OpList)
	}
	if o.GetBuffer {
		ops = append(ops, OpGetBuffer)
	}
	if o.NewBufferStdin {
		ops = append(ops, OpNewBuffer)
	}
	if o.ReplaceBufferStdin {
		ops = append(ops, OpReplaceBuffer)
	}
	if o.InsertStdin {
		ops = append(ops, OpInsert)
	}
	if len(ops) > 1 {
		return "", errors.New("--list-instances, --get-buffer, --new-buffer-stdin, --replace-buffer-stdin and --insert-stdin cannot be combined")
	}
	if len(ops) == 1 {
		switch ops[0] {
		case OpList:
			if o.PID != 0 || o.Work != "" || o.RootTree != "" || len(o.Files) > 0 {
				return "", errors.New("--list-instances takes no other arguments")
			}
		case OpNewBuffer:
			if len(o.Files) > 0 {
				return "", errors.New("--new-buffer-stdin takes no file")
			}
		default: // get / replace / insert
			if len(o.Files) > 1 {
				return "", errors.New("at most one file can be given with a buffer operation")
			}
		}
		return ops[0], nil
	}
	switch {
	case len(o.Files) > 0:
		return OpOpen, nil
	case o.Work != "" || o.RootTree != "":
		return OpSwitch, nil
	}
	return "", errors.New("nothing to do: give a file, --work/--root-tree, or an operation")
}

// Run executes ge-remote and returns the exit code. Buffer data goes
// to stdout only; every message goes to stderr, so
// "ge-remote --get-buffer > out.go" is safe.
func Run(o Options, stdin io.Reader, stdout, stderr io.Writer) ExitCode {
	if o.Help && len(o.Errors) == 0 {
		fmt.Fprint(stdout, Usage)
		return ExitSuccess
	}
	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			fmt.Fprintln(stderr, "ge-remote:", e)
		}
		fmt.Fprintln(stderr, "Try 'ge-remote --help' for usage.")
		return ExitInvalidArgument
	}
	op, err := o.operation()
	if err != nil {
		fmt.Fprintln(stderr, "ge-remote:", err)
		fmt.Fprintln(stderr, "Try 'ge-remote --help' for usage.")
		return ExitInvalidArgument
	}
	if op == OpList {
		return listInstances(stdout, stderr)
	}

	pid, ok := SelectInstance(o.PID)
	if !ok {
		if o.PID > 0 {
			fmt.Fprintf(stderr, "ge-remote: no running ge instance with PID %d\n", o.PID)
		} else {
			fmt.Fprintln(stderr, "ge-remote: no running ge instance")
		}
		return ExitInstanceNotFound
	}

	req := Request{Op: op, Work: o.Work, RootTree: o.RootTree, Files: o.Files}
	req.Cwd, _ = os.Getwd()
	switch op {
	case OpNewBuffer, OpReplaceBuffer, OpInsert:
		data, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintln(stderr, "ge-remote: reading standard input:", err)
			return ExitError
		}
		req.Data = data
	}

	resp, err := call(pid, req)
	if err != nil {
		return reportCallError(stderr, err)
	}
	if !resp.OK {
		fmt.Fprintln(stderr, "ge-remote:", errorText(resp))
		if resp.Code == ExitSuccess {
			return ExitOperationFailed
		}
		return resp.Code
	}
	if op == OpGetBuffer {
		if _, err := stdout.Write(resp.Data); err != nil {
			fmt.Fprintln(stderr, "ge-remote: writing output:", err)
			return ExitError
		}
	}
	return ExitSuccess
}

func errorText(r Response) string {
	if r.Error != "" {
		return r.Error
	}
	return "operation failed"
}

func reportCallError(stderr io.Writer, err error) ExitCode {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		fmt.Fprintln(stderr, "ge-remote: timed out waiting for ge")
		return ExitTimeout
	}
	fmt.Fprintln(stderr, "ge-remote: cannot talk to ge:", err)
	return ExitConnectionError
}

// call sends one request to the instance pid and waits for the response.
func call(pid int, req Request) (Response, error) {
	conn, err := net.DialTimeout("unix", SocketPath(pid), dialTimeout)
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(callTimeout))
	if err := encode(conn, req); err != nil {
		return Response{}, err
	}
	var resp Response
	if err := decode(conn, &resp); err != nil {
		return Response{}, err
	}
	return resp, nil
}

func listInstances(stdout, stderr io.Writer) ExitCode {
	pids := LiveInstances()
	if len(pids) == 0 {
		fmt.Fprintln(stderr, "ge-remote: no running ge instance")
		return ExitInstanceNotFound
	}
	tw := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "PID\tWorkspace\tRootTree")
	rows, failed := 0, 0
	for _, pid := range pids {
		resp, err := call(pid, Request{Op: OpList})
		if err != nil || !resp.OK || resp.Instance == nil {
			fmt.Fprintf(stderr, "ge-remote: instance %d did not answer\n", pid)
			failed++
			continue
		}
		for _, w := range resp.Instance.Workspaces {
			for _, rt := range w.RootTrees {
				fmt.Fprintf(tw, "%d\t%s\t%s\n", pid, w.Name, rt)
				rows++
			}
		}
	}
	tw.Flush()
	if rows == 0 && failed > 0 {
		return ExitConnectionError
	}
	return ExitSuccess
}
