package textutil

import "testing"

func TestDefault(t *testing.T) {
	if got := Default("  value  ", "fallback"); got != "value" {
		t.Fatalf("Default() = %q, want %q", got, "value")
	}
	if got := Default(" \t ", "fallback"); got != "fallback" {
		t.Fatalf("Default() = %q, want %q", got, "fallback")
	}
}

func TestFirstNonEmptyPreservesFormatting(t *testing.T) {
	if got := FirstNonEmpty("", "  value  ", "fallback"); got != "  value  " {
		t.Fatalf("FirstNonEmpty() = %q, want original formatting", got)
	}
}

func TestFirstNonEmptyTrimmed(t *testing.T) {
	if got := FirstNonEmptyTrimmed("", "  value  ", "fallback"); got != "value" {
		t.Fatalf("FirstNonEmptyTrimmed() = %q, want %q", got, "value")
	}
	if got := FirstNonEmptyTrimmed(" \t ", ""); got != "" {
		t.Fatalf("FirstNonEmptyTrimmed() = %q, want empty string", got)
	}
}
