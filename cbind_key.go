package gecore

import (
	"errors"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// This code comes from the following package
// https://gitlab.com/tslocum/cbind/key.go
// modified.

// Modifier labels
const (
	LabelCtrl  = "ctrl"
	LabelAlt   = "alt"
	LabelMeta  = "meta"
	LabelShift = "shift"
)

// ErrInvalidKeyEvent is the error returned when encoding or decoding a key event fails.
var ErrInvalidKeyEvent = errors.New("invalid key event")

// UnifyEnterKeys is a flag that determines whether or not KPEnter (keypad
// enter) key events are interpreted as Enter key events. When enabled, Ctrl+J
// key events are also interpreted as Enter key events.
var UnifyEnterKeys = true

var fullKeyNames = map[string]string{
	"backspace2": "Backspace",
	"pgup":       "PageUp",
	"pgdn":       "PageDown",
	"esc":        "Escape",
}

var ctrlKeys = map[rune]tcell.Key{
	' ':  tcell.KeyCtrlSpace,
	'a':  tcell.KeyCtrlA,
	'b':  tcell.KeyCtrlB,
	'c':  tcell.KeyCtrlC,
	'd':  tcell.KeyCtrlD,
	'e':  tcell.KeyCtrlE,
	'f':  tcell.KeyCtrlF,
	'g':  tcell.KeyCtrlG,
	'h':  tcell.KeyCtrlH,
	'i':  tcell.KeyCtrlI,
	'j':  tcell.KeyCtrlJ,
	'k':  tcell.KeyCtrlK,
	'l':  tcell.KeyCtrlL,
	'm':  tcell.KeyCtrlM,
	'n':  tcell.KeyCtrlN,
	'o':  tcell.KeyCtrlO,
	'p':  tcell.KeyCtrlP,
	'q':  tcell.KeyCtrlQ,
	'r':  tcell.KeyCtrlR,
	's':  tcell.KeyCtrlS,
	't':  tcell.KeyCtrlT,
	'u':  tcell.KeyCtrlU,
	'v':  tcell.KeyCtrlV,
	'w':  tcell.KeyCtrlW,
	'x':  tcell.KeyCtrlX,
	'y':  tcell.KeyCtrlY,
	'z':  tcell.KeyCtrlZ,
	'\\': tcell.KeyCtrlBackslash,
	']':  tcell.KeyCtrlRightSq,
	'^':  tcell.KeyCtrlCarat,
	'_':  tcell.KeyCtrlUnderscore,
}

// Decode decodes a string as a key or combination of keys.
func Decode(s string) (mod tcell.ModMask, key tcell.Key, ch rune, err error) {
	if len(s) == 0 {
		return 0, 0, 0, ErrInvalidKeyEvent
	}

	// Special case for plus rune decoding
	if s[len(s)-1:] == "+" {
		key = tcell.KeyRune
		ch = '+'

		if len(s) == 1 {
			return mod, key, ch, nil
		} else if len(s) == 2 {
			return 0, 0, 0, ErrInvalidKeyEvent
		} else {
			s = s[:len(s)-2]
		}
	}

	split := strings.Split(s, "+")
DECODEPIECE:
	for _, piece := range split {
		// Decode modifiers
		pieceLower := strings.ToLower(piece)
		switch pieceLower {
		case LabelCtrl:
			mod |= tcell.ModCtrl
			continue
		case LabelAlt:
			mod |= tcell.ModAlt
			continue
		case LabelMeta:
			mod |= tcell.ModMeta
			continue
		case LabelShift:
			mod |= tcell.ModShift
			continue
		}

		// Decode key
		for shortKey, fullKey := range fullKeyNames {
			if pieceLower == strings.ToLower(fullKey) {
				pieceLower = shortKey
				break
			}
		}
		switch pieceLower {
		case "backspace":
			key = tcell.KeyBackspace2
			continue
		case "space", "spacebar":
			key = tcell.KeyRune
			ch = ' '
			continue
		}
		for k, keyName := range tcell.KeyNames {
			if pieceLower == strings.ToLower(strings.ReplaceAll(keyName, "-", "+")) {
				key = k
				if key < 0x80 {
					ch = rune(k)
				}
				continue DECODEPIECE
			}
		}

		// Decode rune
		if len(piece) > 1 {
			return 0, 0, 0, ErrInvalidKeyEvent
		}

		key = tcell.KeyRune
		ch = rune(piece[0])
	}

	if mod&tcell.ModCtrl != 0 {
		k, ok := ctrlKeys[unicode.ToLower(ch)]
		if ok {
			key = k
			if UnifyEnterKeys && key == ctrlKeys['j'] {
				// key = tcell.KeyEnter
				// verb.PP("ctrl j 1")
				key = tcell.KeyCtrlJ
			} else if key < 0x80 {
				ch = rune(key)
			}
		}
	}

	switch key {
	case tcell.KeyEsc:
		ch = 0
	case tcell.KeyBS:
		// Enable Ctrl+H
		mod = 0
	case tcell.KeyCtrlJ:
		// Enable Ctrl+J
		mod = tcell.ModCtrl
		//ch = 'j'
		// ch = 0
	}
	return mod, key, ch, nil
}
