package screen

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Cell represents a single character Cell on screen.
type Cell struct {
	Ch    rune // int32
	Style tcell.Style
	Width int // screen cell width
	Class CharClass
}

/*
// Character classification
type CharClass int

const (
	OTHER             CharClass = 1 << iota
	CONTROLCODE                 // ^X
	TAB                         // ^X
	LINEFEED                    // ^X
	DEL                         // ^?
	EOF                         // ^Z
	NUMBER                      //
	ALPHABET                    //
	PROHIBITED                  //
	DECIMAL_SEPARATOR           // comma, dot
	WIDECHAR                    // Zenkaku
	UPPERCASE
	SYMBOL
	SPACE
)

// Requires different processing depending on locale
// an prohibited character,
// Return the character classification
func GetCharClass(ch rune) (cc CharClass) {
	// SPACE!"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~DEL

	// prohibited_characters := "\t !&)*+,-./:;<=>?@]|}~　、。」）！？をんぁぃぅぇぉっゃゅょゎァィゥェォッャュョヮヵヶ"
	prohibited_characters := "\t !&)*+,-./:;<=>?@]|}~　、。」）！？をん"

	if strings.Contains(prohibited_characters, string(ch)) {
		cc |= PROHIBITED
	}
	if ch == ',' || ch == '.' {
		return cc | DECIMAL_SEPARATOR
	}
	if ch == define.DEL {
		return cc | CONTROLCODE | DEL
	}
	if ch == '\t' {
		return cc | CONTROLCODE | TAB
	}
	if ch == ' ' {
		return cc | SPACE
	}
	if ch == '　' {
		return cc | SPACE | WIDECHAR
	}
	if ch == define.LF {
		return cc | CONTROLCODE | LINEFEED
	}
	if ch == define.EOF {
		return cc | CONTROLCODE | EOF
	}
	if ch < 32 {
		return cc | CONTROLCODE
	}

	if ch >= '0' && ch <= '9' {
		return cc | NUMBER
	}
	if ch >= 'A' && ch <= 'Z' {
		return cc | ALPHABET | UPPERCASE
	}
	if ch >= 'a' && ch <= 'z' {
		return cc | ALPHABET
	}
	if ch < 128 {
		return cc | SYMBOL
	}

	if ch >= '０' && ch <= '９' { // zenkaku number
		return cc | NUMBER | WIDECHAR
	}
	if ch >= 'Ａ' && ch <= 'Ｚ' { // zenkaku alpha
		return cc | ALPHABET | WIDECHAR | UPPERCASE
	}
	if ch >= 'ａ' && ch <= 'ｚ' { // zenkaku alpha
		return cc | ALPHABET | WIDECHAR
	}
	if unicode.Is(unicode.Hiragana, ch) {
		return cc
	}
	if unicode.Is(unicode.Katakana, ch) {
		return cc
	}
	return cc | OTHER
}
*/

// Character classification
type CharClass int

const (
	OTHER             CharClass = 1 << iota
	CONTROLCODE                 // ^X
	TAB                         // ^X
	LINEFEED                    // ^X
	DEL                         // ^?
	EOF                         // ^Z
	NUMBER                      //
	ALPHABET                    //
	PROHIBITED                  //
	DECIMAL_SEPARATOR           // comma, dot
	WIDECHAR                    // Zenkaku
	UPPERCASE
	SYMBOL
	SPACE
	HIRAGANA // ひらがな
	KATAKANA // カタカナ
)

// Return the character classification
func GetCharClass(ch rune) (cc CharClass) {
	prohibitedCharacters := "\t !&)*+,-./:;<=>?@]|}~　、。」）！？をん"
	if strings.ContainsRune(prohibitedCharacters, ch) {
		cc |= PROHIBITED
	}

	switch {
	case ch == ',' || ch == '.':
		cc |= DECIMAL_SEPARATOR
	case ch == '\t':
		cc |= CONTROLCODE | TAB
	case ch == ' ':
		cc |= SPACE
	case ch == '　':
		cc |= SPACE | WIDECHAR
	case ch == 127: // DEL character
		cc |= CONTROLCODE | DEL
	case ch == '\n':
		cc |= CONTROLCODE | LINEFEED
	case ch == 26: // EOF (typically Ctrl+Z)
		cc |= CONTROLCODE | EOF
	case ch < 32:
		cc |= CONTROLCODE
	case unicode.IsDigit(ch):
		cc |= NUMBER
	case unicode.IsLetter(ch):
		cc |= ALPHABET
		if unicode.IsUpper(ch) {
			cc |= UPPERCASE
		}
	case unicode.IsSymbol(ch) || unicode.IsPunct(ch):
		cc |= SYMBOL
	case unicode.Is(unicode.Hiragana, ch):
		cc |= HIRAGANA | WIDECHAR
	case unicode.Is(unicode.Katakana, ch):
		cc |= KATAKANA | WIDECHAR
	case ch >= '０' && ch <= '９': // 全角数字
		cc |= NUMBER | WIDECHAR
	case ch >= 'Ａ' && ch <= 'Ｚ': // 全角アルファベット（大文字）
		cc |= ALPHABET | WIDECHAR | UPPERCASE
	case ch >= 'ａ' && ch <= 'ｚ': // 全角アルファベット（小文字）
		cc |= ALPHABET | WIDECHAR
	case ch >= 0x4E00 && ch <= 0x9FFF: // CJK統合漢字の範囲を使用（全角の一部として）
		cc |= WIDECHAR
	default:
		cc |= OTHER
	}

	return cc
}
