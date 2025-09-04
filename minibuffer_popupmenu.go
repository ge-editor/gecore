package gecore

import (
	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/kill_buffer"
	"github.com/ge-editor/gecore/screen"
	"github.com/ge-editor/gecore/verb"

	"github.com/ge-editor/utils"
)

func NewMiniBufferPopupmenu(message string, prefix string, echo bool) *MiniBufferPopupmenu {
	prefixWidth := 0
	for _, ch := range prefix {
		prefixWidth += utils.RuneWidth(ch)
	}

	mb := &MiniBufferPopupmenu{
		MiniBuffer: NewMiniBuffer(message, prefix, echo),
		Popupmenu:  NewPopupmenu(utils.Rect{0, 0, 10, 10}, []string{}, 0),
		Screen:     screen.Get(),
		histories:  []string{},
		// KeyPointer: KeyMapper(),
	}
	return mb
}

type MiniBufferPopupmenu struct {
	*MiniBuffer
	*Popupmenu
	showPopupmenu bool
	*screen.Screen
	histories []string
}

func (m *MiniBufferPopupmenu) IsShowPopupmenu() bool {
	return m.showPopupmenu
}

func (m *MiniBufferPopupmenu) ShowPopupmenu(b bool) {
	m.showPopupmenu = b
}

func (m *MiniBufferPopupmenu) Draw() {
	m.MiniBuffer.Draw()
	if m.Popupmenu == nil {
		// The position where the Popup menu is displayed is based on the minibuffer cursor position.
		m.Popupmenu = NewPopupmenu(utils.Rect{X: m.CX, Y: m.CY, Width: 32, Height: 10}, m.histories, 0)
	}
	if m.showPopupmenu {
		m.Popupmenu.Draw()
	}
}

func (m *MiniBufferPopupmenu) Event(eKey *tcell.EventKey) *tcell.EventKey {
	str := string(m.MiniBuffer.String())
	verb.PP("MiniBufferPopupmenu Event %v", str)

	switch eKey.Key() {
	case tcell.KeyEscape:
		m.showPopupmenu = false
	case tcell.KeyEnter:
		if !m.showPopupmenu {
			break
		}
		index, s := m.Item()
		if index >= 0 {
			// If the caller assigned an action to the Enter key, an error may occur on the caller.
			//m.histories = utils.MoveElement(m.histories, index, true)
			//m.Popupmenu.Set(m.histories, 0)
			str = s
			m.MiniBuffer.Set(str, len(str))
		}
	case tcell.KeyCtrlY:
		// Yank from kill buffer
		s := string(kill_buffer.KillBuffer.GetLast())
		m.MiniBuffer.Set(s, len(s))
	case tcell.KeyTAB: // Popup search history
		m.showPopupmenu = !m.showPopupmenu
		if m.showPopupmenu {
			m.setBeFilteredHistoriesToPopupMenu(str)
		}
		/*
			case tcell.KeyCtrlN, tcell.KeyDown, tcell.KeyCtrlP, tcell.KeyUp:
				if m.showPopupmenu {
					m.popupmenu.Event(eKey)
				} else {
					m.minibuffer.Event(eKey)
				}
		*/
	default:
		// m.minibuffer.Event(eKey)
		if m.showPopupmenu {
			m.Popupmenu.Event(eKey)
		} else {
			m.MiniBuffer.Event(eKey)
			str = string(m.MiniBuffer.String())
			if str == "" {
				break
			}
			m.setBeFilteredHistoriesToPopupMenu(str)
		}
	}
	return eKey
}

func (m *MiniBufferPopupmenu) setBeFilteredHistoriesToPopupMenu(str string) {
	items := []string{}
	for _, h := range m.histories {
		if utils.ContainsAllCharacters(h, str) {
			items = append(items, h)
		}
	}
	m.Popupmenu.Set(items, 0)
}
