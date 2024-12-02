package kill_buffer

import (
	"github.com/atotto/clipboard"

	"github.com/ge-editor/gecore/verb"
)

/*
Clipboard for Go

Provide copying and pasting to the Clipboard for Go.

Build:

$ go get github.com/atotto/clipboard

Platforms:

    OSX
    Windows 7 (probably work on other Windows)
    Linux, Unix (requires 'xclip' or 'xsel' command to be installed)
*/

// ViewLeaf common kill buffer
var KillBuffer = &killBuffer{}

type killBuffer [][]byte

func (kb *killBuffer) PushKillBuffer(buff []byte) error {
	*kb = append(*kb, buff)

	err := clipboard.WriteAll(string(buff))
	if err != nil {
		verb.PP(err.Error())
	}
	return err
}

func (kb *killBuffer) PopKillBuffer() []byte {
	l := len(*kb)
	if l == 0 {
		return nil
	}
	buff := (*kb)[l-1]
	*kb = (*kb)[:l-1]
	return buff
}

func (kb *killBuffer) GetLast() []byte {
	return kb.Get(len(*kb) - 1)
}

// Get retrieves the element at the specified index in the buffer,
// and then moves that element to the end of the buffer.
func (kb *killBuffer) Get(index int) []byte {
	l := len(*kb)
	if index < 0 || index >= l {
		return nil
	}
	// Get the requested element.
	result := (*kb)[index]
	// Remove the element from the original position.
	copy((*kb)[index:], (*kb)[index+1:])
	*kb = (*kb)[:l-1] // Remove the last (redundant) element
	// Append the result to the end of the buffer.
	*kb = append(*kb, result)
	return result
}
