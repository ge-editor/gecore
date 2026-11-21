package lang

import (
	"bytes"
	"sort"
)

// CommentStyle is how a language writes a comment that covers whole lines.
type CommentStyle struct {
	// Line is the string that starts a comment that ends with the row:
	// "//", "#". Empty if the language has none.
	Line string

	// Block is used when Line is empty: each line is wrapped in it
	// ("<!--" "-->"), so that the lines stay independent comments.
	Block *BlockDelimiter
}

// CommentStyleProvider is implemented by a Mode that knows how to write a
// comment (optional). A Mode that has SpellDelimiters but no CommentStyle gets
// its style from them (see CommentStyleOf).
type CommentStyleProvider interface {
	CommentStyle() CommentStyle
}

// CommentStyleOf returns the comment style of a mode. ok is false when the
// mode has no way to write a comment (Fundamental mode).
func CommentStyleOf(m Mode) (st CommentStyle, ok bool) {
	if m == nil {
		return CommentStyle{}, false
	}

	if p, isProvider := m.(CommentStyleProvider); isProvider {
		st = p.CommentStyle()
		return st, st.Line != "" || st.Block != nil
	}

	if d := SpellDelimitersOf(m); d != nil {
		for _, lc := range d.LineComments {
			if lc != "" {
				return CommentStyle{Line: lc}, true
			}
		}

		for i := range d.BlockComments {
			if b := d.BlockComments[i]; b.Open != "" && b.Close != "" {
				return CommentStyle{Block: &b}, true
			}
		}
	}

	return CommentStyle{}, false
}

// CommentEdit replaces Del bytes at byte column Col of line Line with Ins.
// Line is an index into the lines given to CommentEdits; Col is in the
// original text of the line.
type CommentEdit struct {
	Line, Col, Del int
	Ins            string
}

// CommentEdits works out the edits that toggle the comments of lines: if
// every line that is not blank is a comment it removes the comment marks,
// otherwise it comments them all. Blank lines are left alone.
//
// The edits are in descending order of (Line, Col), so that they can be
// applied one after the other without the earlier ones moving the later ones.
// commented tells which way it went. ok is false when there is nothing to do
// (no style, or only blank lines).
func CommentEdits(lines [][]byte, st CommentStyle) (edits []CommentEdit, commented, ok bool) {
	if st.Line == "" && st.Block == nil {
		return nil, false, false
	}

	var rows []int

	for i, l := range lines {
		if len(bytes.TrimSpace(l)) > 0 {
			rows = append(rows, i)
		}
	}

	if len(rows) == 0 {
		return nil, false, false
	}

	if st.Line != "" {
		edits, commented = lineCommentEdits(lines, rows, st.Line)
	} else {
		edits, commented = blockCommentEdits(lines, rows, st.Block)
	}

	sort.SliceStable(edits, func(a, b int) bool {
		if edits[a].Line != edits[b].Line {
			return edits[a].Line > edits[b].Line
		}

		return edits[a].Col > edits[b].Col
	})

	return edits, commented, true
}

// indentLen is the length of the white space at the start of l.
func indentLen(l []byte) int {
	i := 0
	for i < len(l) && (l[i] == ' ' || l[i] == '\t') {
		i++
	}

	return i
}

func lineCommentEdits(lines [][]byte, rows []int, marker string) ([]CommentEdit, bool) {
	m := []byte(marker)

	allCommented := true

	for _, r := range rows {
		l := lines[r]
		if !bytes.HasPrefix(l[indentLen(l):], m) {
			allCommented = false
			break
		}
	}

	var edits []CommentEdit

	if allCommented {
		for _, r := range rows {
			l := lines[r]
			i := indentLen(l)

			del := len(m)
			if i+del < len(l) && l[i+del] == ' ' {
				del++ // the space that was put after the mark
			}

			edits = append(edits, CommentEdit{Line: r, Col: i, Del: del})
		}

		return edits, false
	}

	// The mark goes at the indentation that all the lines have, so that the
	// marks line up.
	common := -1

	for _, r := range rows {
		l := lines[r]
		ind := l[:indentLen(l)]

		if common < 0 {
			common = len(ind)
			continue
		}

		n := 0
		for n < common && n < len(ind) && ind[n] == lines[rows[0]][n] {
			n++
		}

		common = n
	}

	for _, r := range rows {
		edits = append(edits, CommentEdit{Line: r, Col: common, Ins: marker + " "})
	}

	return edits, true
}

func blockCommentEdits(lines [][]byte, rows []int, b *BlockDelimiter) ([]CommentEdit, bool) {
	open, closing := []byte(b.Open), []byte(b.Close)

	isComment := func(l []byte) bool {
		t := bytes.TrimSpace(l)

		return len(t) >= len(open)+len(closing) &&
			bytes.HasPrefix(t, open) &&
			bytes.HasSuffix(t, closing)
	}

	allCommented := true

	for _, r := range rows {
		if !isComment(lines[r]) {
			allCommented = false
			break
		}
	}

	var edits []CommentEdit

	for _, r := range rows {
		l := lines[r]
		i := indentLen(l)
		end := len(bytes.TrimRight(l, " \t\r"))

		if !allCommented {
			edits = append(edits,
				CommentEdit{Line: r, Col: i, Ins: b.Open + " "},
				CommentEdit{Line: r, Col: end, Ins: " " + b.Close},
			)

			continue
		}

		startDel := len(open)
		if i+startDel < len(l) && l[i+startDel] == ' ' {
			startDel++
		}

		closeAt := end - len(closing)

		closeDel := len(closing)
		if closeAt-1 >= i+startDel && l[closeAt-1] == ' ' {
			closeAt--
			closeDel++
		}

		edits = append(edits,
			CommentEdit{Line: r, Col: i, Del: startDel},
			CommentEdit{Line: r, Col: closeAt, Del: closeDel},
		)
	}

	return edits, !allCommented
}

// AdjustCol returns where the byte column col of line moves to when the
// edits (as given by CommentEdits, in original coordinates) are applied: an
// edit before it moves it, a cursor inside the removed bytes goes to the
// end of what replaced them.
func AdjustCol(edits []CommentEdit, line, col int) int {
	shift := 0

	for _, e := range edits {
		if e.Line != line {
			continue
		}

		switch {
		case e.Col+e.Del <= col:
			shift += len(e.Ins) - e.Del
		case e.Col < col:
			shift += e.Col - col + len(e.Ins)
		}
	}

	return col + shift
}
