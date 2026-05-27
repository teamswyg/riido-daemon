package policy

import (
	"strings"
	"testing"
)

func TestRedactSecretPatternsRedactsCatalogMatches(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		patternID string
	}{
		{"aws access key", "AKIA" + strings.Repeat("A", 16), "aws-access-key"},
		{"gcp api key", "AIza" + strings.Repeat("a", 35), "gcp-api-key"},
		{"github token", "ghp_" + strings.Repeat("a", 20), "github-token"},
		{"gitlab token", "glpat-" + strings.Repeat("b", 20), "gitlab-token"},
		{"openai key", "sk-" + strings.Repeat("c", 24), "openai-api-key"},
		{"anthropic key", "sk-ant-" + strings.Repeat("d", 24), "anthropic-api-key"},
		{"jwt", "eyJ" + strings.Repeat("e", 8) + "." + strings.Repeat("f", 12) + "." + strings.Repeat("g", 12), "jwt"},
		{"pem private key", strings.Join([]string{"-----BEGIN", "PRIVATE KEY-----\nabc\n-----END PRIVATE KEY-----"}, " "), "pem-private-key"},
		{"basic auth url", "https://user:pass@example.com/path", "basic-auth-url"},
		{"env token", "RIIDO_TOKEN=" + strings.Repeat("h", 12), "env-secret-assignment"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			redacted, ids := RedactSecretPatterns("prefix "+tc.value+" suffix", nil)
			if len(ids) != 1 || ids[0] != tc.patternID {
				t.Fatalf("pattern IDs = %v, want %q", ids, tc.patternID)
			}
			if strings.Contains(redacted, tc.value) {
				t.Fatalf("redacted value still contains raw secret: %q", redacted)
			}
			if !strings.Contains(redacted, SecretRedactionMarker(tc.patternID)) {
				t.Fatalf("redacted value missing marker %q: %q", SecretRedactionMarker(tc.patternID), redacted)
			}
		})
	}
}

func TestRedactSecretPatternsSupportsCustomMarker(t *testing.T) {
	redacted, ids := RedactSecretPatterns("token ghp_"+strings.Repeat("a", 20), func(string) string {
		return "[redacted]"
	})

	if len(ids) != 1 || ids[0] != "github-token" {
		t.Fatalf("pattern IDs = %v", ids)
	}
	if redacted != "token [redacted]" {
		t.Fatalf("redacted value = %q", redacted)
	}
}

func TestContainsSecretPatternLeavesSafeValuesAlone(t *testing.T) {
	for _, value := range []string{"go test ./...", "sample=banana", "https://example.com/path"} {
		if ContainsSecretPattern(value) {
			t.Fatalf("safe value matched secret pattern: %q", value)
		}
	}
}
