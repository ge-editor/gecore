package tree

import "github.com/ge-editor/gelog"

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

type OpenerFlag int

const (
	ActiveLeaf OpenerFlag = iota
	InactiveLeaf
	FirstLeaf
	ActiveAndFirstLeaf
	ActiveOrFirstLeaf
)

func Opener(destination, behavior string, flag OpenerFlag, data any, f func(bool, error)) {

	lt, ok := LeafTypes.GetLeafTypeByReginsterName(destination)
	if !ok {
		return
	}

	lt.Opener(destination, behavior, flag, data, func(handled bool, err error) {
		if err != nil {
			gelog.Error(err.Error())
		}
	})
}
