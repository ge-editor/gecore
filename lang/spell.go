package lang

// Delimiters tell where the prose of a language is: in its comments and
// strings. The spell check looks at nothing else of a program, so a language
// mode that has them gets its comments and strings checked, and the code is
// left alone.
//
// A Mode gives them by implementing SpellDelimiters (optional: a Mode without
// it is taken as text, and every word of it is checked).
type Delimiters struct {
	// LineComments are the strings that start a comment which ends with the
	// row: "//", "#".
	LineComments []string

	// BlockComments are the comments between two strings, which may span
	// rows: "/*" "*/".
	BlockComments []BlockDelimiter

	// Strings are the quoted literals. Add the ones that are not text
	// (a character literal) with Code set, so that a quote inside them does
	// not start a string.
	Strings []StringDelimiter
}

// BlockDelimiter is a comment from Open to Close.
type BlockDelimiter struct {
	Open, Close string
}

// StringDelimiter is a quoted literal from Open to Close.
type StringDelimiter struct {
	Open, Close string

	// Escape is the byte that makes the next byte part of the literal
	// ('\\'); 0 if there is none (raw strings). An escape sequence is not
	// prose either: the n of "\n" is not a word.
	Escape byte

	// Multiline: the literal goes on in the next row if it is not closed
	// (a raw string). Otherwise it ends with the row.
	Multiline bool

	// Code: the content is not prose (a character literal such as 'a').
	Code bool
}

// SpellDelimiterProvider is implemented by a Mode that has Delimiters.
type SpellDelimiterProvider interface {
	// SpellDelimiters returns the delimiters, or nil if the whole text is
	// prose.
	SpellDelimiters() *Delimiters
}

// SpellDelimitersOf returns the delimiters of a mode, or nil.
func SpellDelimitersOf(m Mode) *Delimiters {
	if p, ok := m.(SpellDelimiterProvider); ok {
		return p.SpellDelimiters()
	}
	return nil
}
