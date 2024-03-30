package gecore

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/define"
	"github.com/ge-editor/gecore/screen"

	"github.com/ge-editor/utils"

	"github.com/ge-editor/theme"
)

// There is a bug in cursor movement

const (
	threshold = 5 // cursor left or right side
)

var eventKey = KeyMapper()

// Create and initialize a new MiniBuffer instance.
//
// parameters:
//   - message: Message to display in the minibuffer.
//   - prefix: String displayed as prefix.
//   - echo: echo flag. Specifies whether to echo the contents of the minibuffer.
//
// Return value: initialized MiniBuffer instance.
func NewMiniBuffer(message string, prefix string, echo bool) *MiniBuffer {
	prefixWidth := 0
	for _, ch := range prefix {
		prefixWidth += utils.RuneWidth(ch)
	}

	mb := &MiniBuffer{
		screen:      screen.Get(),
		buffer:      []screen.Cell{},
		prefix:      prefix,
		prefixWidth: prefixWidth,
		echo:        echo,
		KeyPointer:  KeyMapper(),
	}
	mb.toBuffer(message)
	mb.bufferIndex = len(message) // Move cursor to end
	return mb
}

// bufferIndex is buffer[bufferIndex]
// Matches the edit cursor position
// In echo mode, buffer[bufferIndex] becomes the drawing start position drawStartIndex
type MiniBuffer struct {
	echo bool

	screen      *screen.Screen
	buffer      []screen.Cell
	bufferIndex int
	cx          int

	prefix      string
	prefixWidth int

	drawStartIndex, drawEndIndex int

	*KeyPointer
}

// Convert the given string to a buffer for MiniBuffer
func (mb *MiniBuffer) toBuffer(str string) {
	mb.buffer = mb.buffer[:0] // empty?
	for _, ch := range str {
		//utils.RuneWidth(ch)
		//ch, size := utf8.DecodeRune(m.raw[i:])
		mb.buffer = append(mb.buffer, screen.Cell{Ch: ch, Width: utils.RuneWidth(ch), Style: theme.ColorDefault})
		//i += size
	}
	mb.buffer = append(mb.buffer, screen.Cell{Ch: '~', Width: 1, Style: theme.ColorSpecialChar})
}

// Draw MiniBuffer and
// update MiniBuffer.drawStartIndex, MiniBuffer.drawEndIndex
func (mb *MiniBuffer) Draw() {
	y := mb.screen.Height - 1
	x := 0

	mb.drawStartIndex, mb.drawEndIndex = mb.screen.DrawMiniBuffer(x, y, mb.screen.Width-x, threshold, mb.buffer, mb.bufferIndex, theme.ColorDefault, mb.prefix, mb.prefixWidth, theme.ColorDefault, mb.echo)
}

// Return editing data as string
func (mb *MiniBuffer) String() string {
	var s string
	for _, a := range mb.buffer[:len(mb.buffer)-1] { // remove EOF mark
		s += string(a.Ch)
	}
	return s
}

func (mb *MiniBuffer) Index() int {
	return mb.bufferIndex
}

func (mb *MiniBuffer) Set(message string, index int) {
	mb.toBuffer(message)
	if index > len(mb.buffer)-1 {
		index = len(mb.buffer) - 1
	}
	mb.bufferIndex = index
}

//-------------------

func (mb *MiniBuffer) CursorForward() {
	if mb.echo {
		if mb.drawEndIndex >= len(mb.buffer)-1 {
			return
		}

	} else if mb.bufferIndex >= len(mb.buffer)-1 {
		return
	}
	mb.bufferIndex++
}

func (mb *MiniBuffer) CursorBackward() {
	if mb.echo {
		if mb.drawStartIndex <= 0 {
			return
		}
	} else if mb.bufferIndex <= 0 {
		return
	}
	mb.bufferIndex--
}

func (mb *MiniBuffer) CursorHome() {
	mb.bufferIndex = 0
}

func (mb *MiniBuffer) CursorEnd() {
	mb.bufferIndex = len(mb.buffer) - 1
}

func (mb *MiniBuffer) InsertRune(ch rune) {
	after := append(make([]screen.Cell, 1), mb.buffer[mb.bufferIndex:]...)
	after[0] = screen.Cell{Ch: ch, Width: utils.RuneWidth(ch), Style: theme.ColorDefault}
	mb.buffer = append(mb.buffer[:mb.bufferIndex], after...)
	mb.bufferIndex++
}

func (mb *MiniBuffer) DeleteRuneBackward() {
	index := mb.bufferIndex - 1
	if index < 0 {
		return
	}
	if index >= len(mb.buffer)-1 {
		return
	} else {
		mb.buffer = append(mb.buffer[:index], mb.buffer[index+1:]...)
	}
	mb.bufferIndex--
}

func (mb *MiniBuffer) DeleteRune() {
	index := mb.bufferIndex
	if index < 0 {
		return
	}
	if index >= len(mb.buffer)-1 {
		return
	} else {
		mb.buffer = append(mb.buffer[:index], mb.buffer[index+1:]...)
	}
}

// Truncate: Remove characters before the cursor position.
func (mb *MiniBuffer) Truncate() {
	index := mb.bufferIndex
	mb.buffer = mb.buffer[index:]
	mb.bufferIndex = 0
}

// Cutoff: Remove characters after the cursor position.
func (mb *MiniBuffer) Cutoff() {
	index := mb.bufferIndex
	if mb.bufferIndex <= 1 { // end of line mark
		return
	}
	mb.buffer = mb.buffer[:index]
	mb.bufferIndex--
}

func (mb *MiniBuffer) Event(tev *tcell.EventKey) *tcell.EventKey {
	// e.Execute(tev, false)

	switch tev.Key() {
	case tcell.KeyCtrlF, tcell.KeyRight:
		mb.CursorForward()
	case tcell.KeyCtrlB, tcell.KeyLeft:
		mb.CursorBackward()
	default:
		if !mb.echo {
			switch tev.Key() {
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				mb.DeleteRuneBackward()
			case tcell.KeyDelete, tcell.KeyCtrlD:
				mb.DeleteRune()
			case tcell.KeyCtrlE:
				mb.CursorEnd()
			// mac delete-key is this
			case tcell.KeyCtrlA:
				mb.CursorHome()
			case tcell.KeyCtrlU:
				mb.Truncate()
			case tcell.KeyCtrlK:
				mb.Cutoff()
			default:
				if tev.Rune() < 32 || tev.Rune() == define.DEL {
					return tev
				}
				mb.InsertRune(tev.Rune())
			}
		} // if !e.echo
	}
	return tev
}
