package toolargs

import (
	"strings"
	"testing"
)

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
