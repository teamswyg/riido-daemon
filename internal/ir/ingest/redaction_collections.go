package ingest

import (
	"fmt"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func redactStringMap(in map[string]string, path string) (map[string]string, redactionSummary) {
	out := make(map[string]string, len(in))
	var summary redactionSummary
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		redacted, patternIDs := policy.RedactSecretPatterns(in[key], policy.SecretRedactionMarker)
		out[key] = redacted
		summary.add(joinPath(path, key), patternIDs)
	}
	return out, summary
}

func redactAnySlice(in []any, path string) ([]any, redactionSummary) {
	out := make([]any, len(in))
	var summary redactionSummary
	for idx, item := range in {
		redacted, child := redactValue(item, fmt.Sprintf("%s.%d", path, idx))
		out[idx] = redacted
		summary.merge(child)
	}
	return out, summary
}

func redactStringSlice(in []string, path string) ([]string, redactionSummary) {
	out := make([]string, len(in))
	var summary redactionSummary
	for idx, item := range in {
		redacted, patternIDs := policy.RedactSecretPatterns(item, policy.SecretRedactionMarker)
		out[idx] = redacted
		summary.add(fmt.Sprintf("%s.%d", path, idx), patternIDs)
	}
	return out, summary
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
