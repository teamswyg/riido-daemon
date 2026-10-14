package toolargs

import (
	"strings"
	"testing"
)

func TestToolArgsFromValueKeepsSafeValues(t *testing.T) {
	args := FromValue(map[string]any{
		"command": "go test ./...",
		"note":    "sample=banana",
	})

	if args["command"] != "go test ./..." {
		t.Fatalf("safe command should not be redacted: %+v", args)
	}
	if args["note"] != "sample=banana" {
		t.Fatalf("safe note should not be redacted: %+v", args)
	}
	if HasRedactedValue(args) {
		t.Fatalf("safe args must not report redacted value: %+v", args)
	}
}

func TestToolArgsFromValueFlattensAndBoundsProviderInput(t *testing.T) {
	args := FromValue(map[string]any{
		"path": ".git/config",
		"headers": map[string]any{
			"authorization": "Bearer raw",
		},
		"items": []any{"a", "b"},
	})

	if args["path"] != ".git/config" {
		t.Fatalf("path arg = %q", args["path"])
	}
	if args["headers.authorization"] != RedactedValue {
		t.Fatalf("authorization must be redacted: %+v", args)
	}
	if args["items.1"] != "b" {
		t.Fatalf("list arg = %+v", args)
	}
}

func TestToolArgsFromValueTruncatesLongStrings(t *testing.T) {
	args := FromValue(map[string]any{"command": strings.Repeat("x", maxValueRunes+12)})

	if got := len([]rune(args["command"])); got != maxValueRunes {
		t.Fatalf("command length = %d, want %d", got, maxValueRunes)
	}
}
