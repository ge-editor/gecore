package tree

// 共通で使えそうなリクエストはここに定義しておく

// This file holds the vocabulary leaves use to ask each other for things.
//
// A request is a plain data type defined here, so the sender and the
// receiver do not import each other. A leaf declares what it can handle by
// implementing the matching capability interface (FileOpener, ...). The
// sender never names a leaf or a leaf type: it calls a dispatch function
// (OpenFile, ...) and gecore finds a leaf that implements the capability.
//
// To add an action:
//  1. define a request type,
//  2. define a capability interface with a dedicated method name (one leaf
//     must be able to implement several capabilities, so the method names
//     must differ),
//  3. add a dispatch function that walks Candidates[I]().
//
// All of this runs on the UI goroutine, like every other tree operation.

// ----------------------------------
// OpenFile
// ----------------------------------

// OpenRequest asks for a file to be opened.
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
