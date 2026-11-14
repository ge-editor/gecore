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
