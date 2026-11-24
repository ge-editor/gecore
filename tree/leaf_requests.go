package tree

// This file contains payload types accepted by LeafType.DispatchLeafRequest. Keeping the
// request types in gecore lets leaf packages exchange requests without
// importing one another.

// ----------------------------------
// OpenFile
// ----------------------------------

// OpenRequest asks an editor leaf to open a file.
type OpenRequest struct {
	// Path of the file to open.
	Path string
	// Line and Col place the cursor. Both are 1-based; Col counts bytes
	// within the line, as editbuffer.RowsPos.ColIndex does.
	// The zero value means "not specified": keep the position the receiver
	// would use anyway (the top, or the position remembered for the file).
	// A receiver that has no notion of a position ignores both.
	Line, Col int
}

// ----------------------------------
// Replace
// ----------------------------------

// ReplaceRequest asks an editor leaf to replace text in several files (a
// "replace in files" from a search result list).
//
// The receiver checks every edit against the text it has (an open buffer, or
// the file on disk) and refuses the whole request when the text is not what
// Old says: the search result is then out of date. It shows what will change
// and applies it only after the user confirms. The files are edited as
// buffers and are not saved.
type ReplaceRequest struct {
	// Title is the prompt of the confirmation.
	Title string
	Files []ReplaceFile
}

// ReplaceFile is the replacements of one file.
type ReplaceFile struct {
	Path  string
	Edits []ReplaceEdit
}

// ReplaceEdit replaces Old, which starts at Line and Col, with New.
// Line and Col are 1-based and Col counts bytes, as in OpenRequest.
// Old does not contain a line end.
type ReplaceEdit struct {
	Line, Col int
	Old, New  string
}
