package openclaw

import (
	"strings"
	"testing"
)

func TestSanitizeReasonHandlesEmpty(t *testing.T) {
	r := sanitizeReason("")

	if r == "" {
		t.Fatal("empty input must produce a non-empty fallback reason")
	}
}

func TestSanitizeReasonNormalizesNewlines(t *testing.T) {
	in := "line one\nline two\nline three"
	r := sanitizeReason(in)

	if strings.Contains(r, "\n") {
		t.Fatalf("newlines not normalized: %q", r)
	}
	if !strings.Contains(r, "line one") || !strings.Contains(r, "line three") {
		t.Fatalf("content lost: %q", r)
	}
}

func TestSanitizeReasonCapsLength(t *testing.T) {
	huge := strings.Repeat("a", 2000)
	r := sanitizeReason(huge)

	if len(r) > 400 {
		t.Fatalf("reason not capped: %d chars", len(r))
	}
}
