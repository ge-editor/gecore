package killbuffer

import (
	"github.com/atotto/clipboard"

	"github.com/ge-editor/gelog"
	"github.com/ge-editor/utils"
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

// killBuffer stores killed text as a collection of rows.
// Rows do not contain newline characters.
var KillBuffer = &killBuffer{}

type killBuffer [][][]byte

/* func (kb *killBuffer) PushKillBuffer(buff [][]byte, newline []byte) error {
	*kb = append(*kb, buff)

	joined, _, _ := utils.JoinRows(buff, newline, true)
	err := clipboard.WriteAll(string(joined))
	if err != nil {
		gelog.Error(err.Error())
	}
	return err
} */

func (kb *killBuffer) PushKillBuffer(buff [][]byte, newline []byte) error {
	// Keep an independent copy in the kill buffer.
	saved := make([][]byte, len(buff))
	for i, row := range buff {
		saved[i] = append([]byte(nil), row...)
	}

	*kb = append(*kb, saved)

	joined, _, _ := utils.JoinRows(saved, newline, true)
	err := clipboard.WriteAll(string(joined))
	if err != nil {
		gelog.Error(err.Error())
	}
	return err
}

func (kb *killBuffer) PopKillBuffer() [][]byte {
	l := len(*kb)
	if l == 0 {
		return nil
	}
	buff := (*kb)[l-1]
	*kb = (*kb)[:l-1]
	return buff
}

func (kb *killBuffer) GetLast() [][]byte {
	return kb.Get(len(*kb) - 1)
}

// Get retrieves the element at the specified index in the buffer,
// and then moves that element to the end of the buffer.
func (kb *killBuffer) Get(index int) [][]byte {
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
