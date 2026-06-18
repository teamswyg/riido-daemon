package ingest

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func redactDraftPayload(payload, unknown map[string]any) (map[string]any, map[string]any, redactionSummary) {
	var summary redactionSummary
	redactedPayload, payloadSummary := redactMap(payload, "payload")
	redactedUnknown, unknownSummary := redactMap(unknown, "unknown")
	summary.merge(payloadSummary)
	summary.merge(unknownSummary)
	return redactedPayload, redactedUnknown, summary
}

func redactMap(in map[string]any, prefix string) (map[string]any, redactionSummary) {
	if len(in) == 0 {
		return nil, redactionSummary{}
	}
	out := make(map[string]any, len(in))
	var summary redactionSummary
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value, child := redactValue(in[key], joinPath(prefix, key))
		out[key] = value
		summary.merge(child)
	}
	return out, summary
}

func redactValue(value any, path string) (any, redactionSummary) {
	switch v := value.(type) {
	case string:
		redacted, patternIDs := policy.RedactSecretPatterns(v, policy.SecretRedactionMarker)
		var summary redactionSummary
		summary.add(path, patternIDs)
		return redacted, summary
	case map[string]any:
		return redactMap(v, path)
	case map[string]string:
		return redactStringMap(v, path)
	case []any:
		return redactAnySlice(v, path)
	case []string:
		return redactStringSlice(v, path)
	default:
		return value, redactionSummary{}
	}
}
