package toolargs

import (
	"strings"
	"testing"
)

func TestToolArgsFromPairsRedactsSensitiveKeys(t *testing.T) {
	args := FromPairs("command", "go test ./...", "api_token", "secret-value")

	if args["command"] != "go test ./..." {
		t.Fatalf("command arg = %q", args["command"])
	}
	if args["api_token"] != RedactedValue {
		t.Fatalf("api token must be redacted: %+v", args)
	}
}

func TestToolArgsFromValueRedactsSensitiveValuePatterns(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"aws access key", "AKIA" + strings.Repeat("A", 16)},
		{"gcp api key", "AIza" + strings.Repeat("a", 35)},
		{"github token", "ghp_" + strings.Repeat("a", 20)},
		{"gitlab token", "glpat-" + strings.Repeat("b", 20)},
		{"openai key", "sk-" + strings.Repeat("c", 24)},
		{"anthropic key", "sk-ant-" + strings.Repeat("d", 24)},
		{"jwt", "eyJ" + strings.Repeat("e", 8) + "." + strings.Repeat("f", 12) + "." + strings.Repeat("g", 12)},
		{"pem private key", "-----BEGIN " + "PRIVATE KEY-----\nabc\n-----END " + "PRIVATE KEY-----"},
		{"basic auth url", "https://user:pass@example.com/path"},
		{"env token", "RIIDO_TOKEN=" + strings.Repeat("h", 12)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := FromValue(map[string]any{"note": tc.value})
			if args["note"] != RedactedValue {
				t.Fatalf("secret-looking value must be redacted: %+v", args)
			}
			if !HasRedactedValue(args) {
				t.Fatalf("args must report redacted value: %+v", args)
			}
		})
	}
}

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
