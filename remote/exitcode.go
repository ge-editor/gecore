// Package remote implements the ge-remote command (the client of a running
// ge) and the small RPC layer it talks to: one Unix domain socket per ge
// process, served by ge itself (see Listen).
//
// This package depends on the standard library only, so that ge-remote
// does not pull in the terminal, the editor or cgo.
package remote

// ExitCode is the process exit status of ge-remote.
// The values are part of the CLI contract (scripts and AI agents test them):
// do not renumber.
type ExitCode int

const (
	ExitSuccess              ExitCode = iota // 0
	ExitError                                // 1 general error
	ExitInvalidArgument                      // 2 bad argument or option
	ExitInstanceNotFound                     // 3 no (matching) ge instance
	ExitWorkspaceNotFound                    // 4
	ExitRootTreeNotFound                     // 5
	ExitBufferNotFound                       // 6 no target buffer, or the leaf has none
	ExitUnsupportedOperation                 // 7
	ExitConnectionError                      // 8 connect / communication failure
	ExitOperationFailed                      // 9 ge accepted the request but failed
	ExitTimeout                              // 10 no response in time
)
