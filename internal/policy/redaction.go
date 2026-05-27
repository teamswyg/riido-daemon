package policy

import "regexp"

const (
	SecretRedactionMarkerPrefix = "[REDACTED:"
	SecretRedactionMarkerSuffix = "]"
)

type secretPattern struct {
	id string
	re *regexp.Regexp
}

var secretRedactionPatterns = []secretPattern{
	{id: "aws-access-key", re: regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{id: "gcp-api-key", re: regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	{id: "github-token", re: regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9_]{16,}|github_pat_[A-Za-z0-9_]{16,})\b`)},
	{id: "gitlab-token", re: regexp.MustCompile(`\bglpat-[A-Za-z0-9_-]{16,}\b`)},
	{id: "anthropic-api-key", re: regexp.MustCompile(`\bsk-ant-[A-Za-z0-9][A-Za-z0-9_-]{15,}\b`)},
	{id: "openai-api-key", re: regexp.MustCompile(`\bsk-[A-Za-z0-9][A-Za-z0-9_-]{15,}\b`)},
	{id: "jwt", re: regexp.MustCompile(`\beyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)},
	{id: "pem-private-key", re: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{id: "basic-auth-url", re: regexp.MustCompile(`(?i)https?://[^/[:space:]:@]+:[^/[:space:]@]+@`)},
	{id: "env-secret-assignment", re: regexp.MustCompile(`(?i)\b(?:[A-Z0-9]+_)*(?:TOKEN|SECRET|KEY)[[:space:]]*=[[:space:]]*[^[:space:]]+`)},
}

// SecretRedactionMarker returns the canonical IR payload marker for a redacted
// secret pattern.
func SecretRedactionMarker(patternID string) string {
	return SecretRedactionMarkerPrefix + patternID + SecretRedactionMarkerSuffix
}

// ContainsSecretPattern reports whether value matches the C7 redaction catalog.
func ContainsSecretPattern(value string) bool {
	for _, pattern := range secretRedactionPatterns {
		if pattern.re.MatchString(value) {
			return true
		}
	}
	return false
}

// RedactSecretPatterns replaces every C7 secret pattern match with a marker and
// returns the unique pattern IDs that were redacted in catalog order.
func RedactSecretPatterns(value string, marker func(patternID string) string) (string, []string) {
	if marker == nil {
		marker = SecretRedactionMarker
	}
	redacted := value
	seen := map[string]struct{}{}
	var patternIDs []string
	for _, pattern := range secretRedactionPatterns {
		if !pattern.re.MatchString(redacted) {
			continue
		}
		redacted = pattern.re.ReplaceAllStringFunc(redacted, func(string) string {
			if _, ok := seen[pattern.id]; !ok {
				seen[pattern.id] = struct{}{}
				patternIDs = append(patternIDs, pattern.id)
			}
			return marker(pattern.id)
		})
	}
	return redacted, patternIDs
}
