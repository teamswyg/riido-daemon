package cliui

import (
	"strings"
	"testing"
)

func TestHTMLTextKeepsBlocksAndDropsTags(t *testing.T) {
	in := "<p>첫 <strong>문단</strong> &amp; a &lt; b</p><p></p><ul><li>하나</li><li>둘</li></ul>a < b<br/>끝<!-- note -->"
	want := "첫 문단 & a < b\n\n• 하나\n• 둘\na < b\n끝"
	if got := HTMLText(in); got != want {
		t.Fatalf("HTMLText = %q, want %q", got, want)
	}
	if got := HTMLText("plain <unclosed"); got != "plain <unclosed" {
		t.Fatalf("unclosed tag = %q", got)
	}
}

func TestWrapRespectsCellWidth(t *testing.T) {
	for _, tc := range []struct {
		in    string
		width int
	}{
		{"짧은 줄", 20},
		{"가나다라마바사아자차카타파하 가나다라마바사", 10},
		{"word another-word yet-another-very-long-word-without-spaces end", 12},
		{"가", 1},
	} {
		lines := Wrap(tc.in, tc.width)
		if strings.Join(lines, "") == "" {
			t.Fatalf("%q lost its text", tc.in)
		}
		for _, line := range lines {
			if DisplayWidth(line) > max(tc.width, 2) {
				t.Fatalf("%q width %d: line %q is %d cells", tc.in, tc.width, line, DisplayWidth(line))
			}
		}
		if got := strings.ReplaceAll(strings.Join(lines, ""), " ", ""); got != strings.ReplaceAll(tc.in, " ", "") {
			t.Fatalf("%q rewrapped to %q", tc.in, lines)
		}
	}
}
