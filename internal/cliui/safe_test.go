package cliui

import (
	"strings"
	"testing"
)

func TestSafeEscapesTerminalControls(t *testing.T) {
	for in, want := range map[string]string{
		"평범한 이름":            "평범한 이름",
		"clear\x1b[2Jme":    `clear\x1b[2Jme`,
		"two\nrows":         `two\nrows`,
		"tab\there":         `tab\there`,
		"c1\u009bx":         `c1\u009bx`,
		"bidi\u202egnp.exe": `bidi\u202egnp.exe`,
		"bad\xffbyte":       `bad\xffbyte`,
	} {
		if got := Safe(in); got != want {
			t.Fatalf("Safe(%q) = %q, want %q", in, got, want)
		}
		if strings.ContainsAny(Safe(in), "\x1b\n\r") {
			t.Fatalf("Safe(%q) kept a control character", in)
		}
	}
}
