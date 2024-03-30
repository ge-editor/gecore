package screen

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/ge-editor/gecore/define"
)

// Cell represents a single character Cell on screen.
type Cell struct {
	Ch    rune // int32
	Style tcell.Style
	Width int // screen cell width
	Class CharClass
}

// Character classification
type CharClass int16

const (
	OTHER             CharClass = 1 << iota
	CONTROLCODE                 // ^X
	TAB                         // ^X
	LINEFEED                    // ^X
	DEL                         // ^?
	NUMBER                      //
	ALPHABET                    //
	PROHIBITED                  //
	DECIMAL_SEPARATOR           // comma, dot
	WIDECHAR                    // Zenkaku
)

// Requires different processing depending on locale
// an prohibited character,
// Return the character classification
func GetCharClass(ch rune) CharClass {
	// SPACE!"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~DEL

	// prohibited_characters := "\t !&)*+,-./:;<=>?@]|}~　、。」）！？をんぁぃぅぇぉっゃゅょゎァィゥェォッャュョヮヵヶ"
	prohibited_characters := "\t !&)*+,-./:;<=>?@]|}~　、。」）！？をん"

	if ch == ',' || ch == '.' {
		return PROHIBITED | DECIMAL_SEPARATOR
	}
	if strings.Contains(prohibited_characters, string(ch)) {
		return PROHIBITED
	}

	if ch == define.DEL {
		return CONTROLCODE | DEL
	}
	if ch == '\t' {
		return CONTROLCODE | TAB
	}
	if ch == define.LF {
		return CONTROLCODE | LINEFEED
	}
	if ch < 32 {
		return CONTROLCODE
	}

	if ch >= '0' && ch <= '9' {
		return NUMBER
	}
	if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
		return ALPHABET
	}
	if ch >= '０' && ch <= '９' { // zenkaku number
		return NUMBER | WIDECHAR
	}
	if (ch >= 'Ａ' && ch <= 'Ｚ') || (ch >= 'ａ' && ch <= 'ｚ') { // zenkaku alpha
		return ALPHABET | WIDECHAR
	}
	return OTHER
}
