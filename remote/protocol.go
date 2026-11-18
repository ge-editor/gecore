package remote

// Operation names carried in Request.Op.
const (
	OpList          = "list"           // describe the instance (workspaces, root trees)
	OpOpen          = "open"           // open Files
	OpSwitch        = "switch"         // switch (or create) Work / RootTree only
	OpGetBuffer     = "get-buffer"     // return the buffer contents in Response.Data
	OpNewBuffer     = "new-buffer"     // new buffer from Request.Data
	OpReplaceBuffer = "replace-buffer" // replace a whole buffer with Request.Data
	OpInsert        = "insert"         // insert Request.Data at the cursor
)

// FileSpec is one file argument. Path is absolute. Line and Col are 1-based;
// 0 means not specified ("+120:8 file.go").
type FileSpec struct {
	Path string `json:"path"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

// Request is sent by the CLI. Work and RootTree select (and, when missing,
// create) the target environment; empty means "the current one".
type Request struct {
	Op       string     `json:"op"`
	Cwd      string     `json:"cwd,omitempty"` // the CLI's working directory (for a new workspace)
	Work     string     `json:"work,omitempty"`
	RootTree string     `json:"root_tree,omitempty"`
	Files    []FileSpec `json:"files,omitempty"`
	Data     []byte     `json:"data,omitempty"` // base64 in JSON
}

// WorkspaceInfo describes one workspace of an instance.
type WorkspaceInfo struct {
	Name      string   `json:"name"`
	RootTrees []string `json:"root_trees"`
}

// InstanceInfo is the answer to OpList.
type InstanceInfo struct {
	PID        int             `json:"pid"`
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

// Response is sent by ge once the operation has finished (not merely been
// received): the CLI exits 0 only after OK is true.
// When OK is false, Code says why (ExitSuccess is treated as
// ExitOperationFailed) and Error is a message for stderr.
type Response struct {
	OK       bool          `json:"ok"`
	Code     ExitCode      `json:"code,omitempty"`
	Error    string        `json:"error,omitempty"`
	Data     []byte        `json:"data,omitempty"`
	Instance *InstanceInfo `json:"instance,omitempty"`
}
