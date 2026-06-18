package main

import (
	"strings"
	"testing"
)

func TestBridgeDetectIncludesEachProvider(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run([]string{"bridge", "detect"}); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	for _, want := range []string{"claude", "codex", "openclaw", "cursor"} {
		if !strings.Contains(out, `"`+want+`"`) {
			t.Fatalf("detect output missing %s: %s", want, out)
		}
	}
}
