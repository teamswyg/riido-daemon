//go:build darwin || linux || freebsd || netbsd || openbsd

package cliui

import (
	"os"
	"testing"
)

// /dev/null is a character device but not a terminal; redirecting there must
// keep plain, non-interactive behavior.
func TestDetectTreatsDevNullAsPlain(t *testing.T) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	if term := Detect(devNull, func(string) string { return "" }); term.Interactive || term.Color {
		t.Fatalf("Detect(/dev/null) = %+v", term)
	}
}
