package lang

import (
	"reflect"
	"testing"
)

// apply applies edits (descending order) to lines, like the editor does.
func applyCommentEdits(lines []string, edits []CommentEdit) []string {
	out := append([]string(nil), lines...)

	for _, e := range edits {
		l := out[e.Line]
		out[e.Line] = l[:e.Col] + e.Ins + l[e.Col+e.Del:]
	}

	return out
}

func toBytes(lines []string) [][]byte {
	var out [][]byte
	for _, l := range lines {
		out = append(out, []byte(l))
	}

	return out
}

var goStyle = CommentStyle{Line: "//"}
var htmlStyle = CommentStyle{Block: &BlockDelimiter{Open: "<!--", Close: "-->"}}

func TestCommentEdits(t *testing.T) {
	for _, tc := range []struct {
		name          string
		style         CommentStyle
		in, want      []string
		wantCommented bool
	}{
		{
			"line: comment at the common indent, blank lines left",
			goStyle,
			[]string{"\tfoo()", "", "\t\tbar()"},
			[]string{"\t// foo()", "", "\t// \tbar()"},
			true,
		},
		{
			"line: uncomment when all are comments",
			goStyle,
			[]string{"\t// foo()", "", "\t//bar()"},
			[]string{"\tfoo()", "", "\tbar()"},
			false,
		},
		{
			"line: one line that is not a comment makes it comment all",
			goStyle,
			[]string{"// a", "b"},
			[]string{"// // a", "// b"},
			true,
		},
		{
			"line: different indentation lines up at the smaller",
			goStyle,
			[]string{"    a", "  b"},
			[]string{"  //   a", "  // b"},
			true,
		},
		{
			"block: each line is wrapped",
			htmlStyle,
			[]string{"  <p>x</p>", "", "y  "},
			[]string{"  <!-- <p>x</p> -->", "", "<!-- y -->  "}, // the trailing blanks stay
			true,
		},
		{
			"block: unwrapped",
			htmlStyle,
			[]string{"  <!-- <p>x</p> -->", "<!--y-->"},
			[]string{"  <p>x</p>", "y"},
			false,
		},
	} {
		edits, commented, ok := CommentEdits(toBytes(tc.in), tc.style)
		if !ok {
			t.Errorf("%s: not ok", tc.name)
			continue
		}

		if commented != tc.wantCommented {
			t.Errorf("%s: commented = %v", tc.name, commented)
		}

		if got := applyCommentEdits(tc.in, edits); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\n got  %q\n want %q", tc.name, got, tc.want)
		}
	}
}

func TestCommentEditsNothingToDo(t *testing.T) {
	if _, _, ok := CommentEdits(toBytes([]string{"", "  "}), goStyle); ok {
		t.Error("blank lines only: want not ok")
	}

	if _, _, ok := CommentEdits(toBytes([]string{"a"}), CommentStyle{}); ok {
		t.Error("no style: want not ok")
	}
}

// Comment, then uncomment: the text is the same as it was.
func TestCommentEditsRoundTrip(t *testing.T) {
	in := []string{"\tfunc f() {", "\t\treturn 1", "", "\t}"}

	for _, st := range []CommentStyle{goStyle, htmlStyle} {
		edits, _, _ := CommentEdits(toBytes(in), st)
		mid := applyCommentEdits(in, edits)

		edits, commented, _ := CommentEdits(toBytes(mid), st)
		if commented {
			t.Errorf("second toggle should uncomment: %q", mid)
		}

		if got := applyCommentEdits(mid, edits); !reflect.DeepEqual(got, in) {
			t.Errorf("round trip: got %q, want %q", got, in)
		}
	}
}

func TestAdjustCol(t *testing.T) {
	edits, _, _ := CommentEdits(toBytes([]string{"\tfoo()"}), goStyle)
	// "\tfoo()" -> "\t// foo()": the mark goes in at column 1.
	for _, tc := range []struct{ col, want int }{
		{0, 0}, // before the mark
		{1, 4}, // at it: after the inserted text
		{3, 6},
	} {
		if got := AdjustCol(edits, 0, tc.col); got != tc.want {
			t.Errorf("AdjustCol(%d) = %d, want %d", tc.col, got, tc.want)
		}
	}

	// Uncomment: "\t// foo()" -> "\tfoo()"; a cursor inside "// " goes to its start.
	edits, _, _ = CommentEdits(toBytes([]string{"\t// foo()"}), goStyle)
	for _, tc := range []struct{ col, want int }{
		{0, 0}, {2, 1}, {4, 1}, {6, 3},
	} {
		if got := AdjustCol(edits, 0, tc.col); got != tc.want {
			t.Errorf("uncomment AdjustCol(%d) = %d, want %d", tc.col, got, tc.want)
		}
	}

	if got := AdjustCol(edits, 5, 3); got != 3 {
		t.Errorf("other line moved: %d", got)
	}
}
