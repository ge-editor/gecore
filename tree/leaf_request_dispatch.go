package tree

import "github.com/ge-editor/gelog"

// OpenerFlag describes which leaf or leaves a request should target. The
// current DispatchLeafRequest implementation does not inspect this value yet.

type OpenerFlag int

const (
	ActiveLeaf OpenerFlag = iota
	InactiveLeaf
	FirstLeaf
	ActiveAndFirstLeaf
	ActiveOrFirstLeaf
)

// DispatchLeafRequest routes data through the LeafType registered as destination. The
// selected LeafType decides how to handle behavior and data. At present the
// callback is not returned to the caller, and behavior and flag are not used
// by this dispatcher. Calls run on the UI goroutine.
func DispatchLeafRequest(destination, behavior string, flag OpenerFlag, data any, f func(bool, error)) {

	lt, ok := LeafTypes.GetLeafTypeByReginsterName(destination)
	if !ok {
		return
	}

	lt.DispatchLeafRequest(destination, behavior, flag, data, func(handled bool, err error) {
		if err != nil {
			gelog.Error(err.Error())
		}
	})
}
