package remote

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const dialTimeout = 500 * time.Millisecond

// SocketDir is the directory that holds one socket per running ge:
// $XDG_RUNTIME_DIR/ge, or <tmp>/ge-<uid> when that is not set.
func SocketDir() string {
	if d := os.Getenv("XDG_RUNTIME_DIR"); d != "" {
		return filepath.Join(d, "ge")
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("ge-%d", os.Getuid()))
}

// SocketPath is the socket of the ge process pid. The PID is part of the
// file name, so no separate PID file is needed.
func SocketPath(pid int) string {
	return filepath.Join(SocketDir(), "ge-"+strconv.Itoa(pid)+".sock")
}

func parseSocketName(name string) (int, bool) {
	s, ok := strings.CutPrefix(name, "ge-")
	if !ok {
		return 0, false
	}
	s, ok = strings.CutSuffix(s, ".sock")
	if !ok {
		return 0, false
	}
	pid, err := strconv.Atoi(s)
	if err != nil || pid < 1 {
		return 0, false
	}
	return pid, true
}

// LiveInstances returns the PIDs of the ge processes that accept
// connections, largest PID first. Sockets left behind by a process that
// died (connection refused) are removed and not reported.
func LiveInstances() []int {
	dir := SocketDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, ok := parseSocketName(e.Name())
		if !ok {
			continue
		}
		path := filepath.Join(dir, e.Name())
		conn, err := net.DialTimeout("unix", path, dialTimeout)
		if err != nil {
			if errors.Is(err, syscall.ECONNREFUSED) {
				os.Remove(path)
			}
			continue
		}
		conn.Close()
		pids = append(pids, pid)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(pids)))
	return pids
}

// SelectInstance picks the target instance. want > 0 is "--remote=PID" and
// must be alive; otherwise the live instance with the largest PID is used.
// ok is false when there is none.
func SelectInstance(want int) (pid int, ok bool) {
	pids := LiveInstances()
	if want > 0 {
		for _, p := range pids {
			if p == want {
				return p, true
			}
		}
		return 0, false
	}
	if len(pids) == 0 {
		return 0, false
	}
	return pids[0], true
}
