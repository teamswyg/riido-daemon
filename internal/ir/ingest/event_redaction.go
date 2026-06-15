package ingest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func NewUUID7EventID(now time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	ms := uint64(now.UTC().UnixNano() / int64(time.Millisecond))
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = (b[6] & 0x0f) | 0x70
	b[8] = (b[8] & 0x3f) | 0x80
	hexed := make([]byte, 32)
	hex.Encode(hexed, b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexed[0:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32]), nil
}

func copyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	maps.Copy(out, in)
	return out
}

func fsmVersionForEvent(eventType ir.EventType, source int) int {
	if eventType.IsTransition() {
		return source
	}
	return 0
}

type redactionSummary struct {
	patternIDs map[string]struct{}
	fields     map[string]struct{}
}

func (s *redactionSummary) add(path string, patternIDs []string) {
	if len(patternIDs) == 0 {
		return
	}
	if s.patternIDs == nil {
		s.patternIDs = map[string]struct{}{}
	}
	if s.fields == nil {
		s.fields = map[string]struct{}{}
	}
	for _, patternID := range patternIDs {
		s.patternIDs[patternID] = struct{}{}
	}
	if path != "" {
		s.fields[path] = struct{}{}
	}
}

func (s redactionSummary) hasRedaction() bool {
	return len(s.patternIDs) > 0
}

func (s redactionSummary) auditPayload(source ir.CanonicalEvent) map[string]any {
	return map[string]any{
		"category":        "SECRET_LEAK_ATTEMPTED",
		"subject":         strings.Join(sortedKeys(s.patternIDs), ","),
		"severity":        "high",
		"sourceEventID":   source.EventID,
		"sourceEventType": string(source.Type),
		"redactedFields":  sortedKeys(s.fields),
	}
}

func redactDraftPayload(payload, unknown map[string]any) (map[string]any, map[string]any, redactionSummary) {
	var summary redactionSummary
	redactedPayload, payloadSummary := redactMap(payload, "payload")
	redactedUnknown, unknownSummary := redactMap(unknown, "unknown")
	summary.merge(payloadSummary)
	summary.merge(unknownSummary)
	return redactedPayload, redactedUnknown, summary
}

func (s *redactionSummary) merge(other redactionSummary) {
	for patternID := range other.patternIDs {
		s.add("", []string{patternID})
	}
	if len(other.fields) == 0 {
		return
	}
	if s.fields == nil {
		s.fields = map[string]struct{}{}
	}
	for field := range other.fields {
		s.fields[field] = struct{}{}
	}
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
		out := make(map[string]string, len(v))
		var summary redactionSummary
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			redacted, patternIDs := policy.RedactSecretPatterns(v[key], policy.SecretRedactionMarker)
			out[key] = redacted
			summary.add(joinPath(path, key), patternIDs)
		}
		return out, summary
	case []any:
		out := make([]any, len(v))
		var summary redactionSummary
		for idx, item := range v {
			redacted, child := redactValue(item, fmt.Sprintf("%s.%d", path, idx))
			out[idx] = redacted
			summary.merge(child)
		}
		return out, summary
	case []string:
		out := make([]string, len(v))
		var summary redactionSummary
		for idx, item := range v {
			redacted, patternIDs := policy.RedactSecretPatterns(item, policy.SecretRedactionMarker)
			out[idx] = redacted
			summary.add(fmt.Sprintf("%s.%d", path, idx), patternIDs)
		}
		return out, summary
	default:
		return value, redactionSummary{}
	}
}

func joinPath(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
