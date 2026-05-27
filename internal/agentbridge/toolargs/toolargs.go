// Package toolargs owns provider-neutral, redacted tool argument summaries for
// C4 adapter events. It does not own provider raw schemas or C7 policy
// decisions.
package toolargs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

const (
	// RedactedValue is stored when an argument key or value identifies
	// sensitive material. The original value must not be preserved in
	// ToolRef.Args.
	RedactedValue = "[redacted]"

	maxArgs       = 32
	maxDepth      = 4
	maxValueRunes = 256
)

var sensitiveKeyTokens = []string{
	"api_key",
	"apikey",
	"authorization",
	"bearer",
	"credential",
	"password",
	"private_key",
	"secret",
	"token",
}

// FromPairs returns a redacted argument map from alternating key/value strings.
// Empty keys are ignored. An odd trailing value is ignored.
func FromPairs(pairs ...string) map[string]string {
	out := map[string]string{}
	for i := 0; i+1 < len(pairs) && len(out) < maxArgs; i += 2 {
		add(out, pairs[i], pairs[i+1])
	}
	return nilIfEmpty(out)
}

// FromValue flattens a provider raw argument object into a bounded, redacted
// string map. Nested fields use dot notation.
func FromValue(value any) map[string]string {
	out := map[string]string{}
	flatten(out, "", value, 0)
	return nilIfEmpty(out)
}

// Clone returns a defensive copy of args.
func Clone(args map[string]string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	out := make(map[string]string, len(args))
	for key, value := range args {
		out[key] = value
	}
	return out
}

// IsSensitiveKey reports whether key names material that must be redacted.
func IsSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	for _, token := range sensitiveKeyTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

// HasRedactedValue reports whether args contains a redacted value marker.
func HasRedactedValue(args map[string]string) bool {
	for _, value := range args {
		if IsRedactedValue(value) {
			return true
		}
	}
	return false
}

// IsRedactedValue reports whether value is the ToolRef.Args redaction marker.
func IsRedactedValue(value string) bool {
	return strings.TrimSpace(value) == RedactedValue
}

func flatten(out map[string]string, prefix string, value any, depth int) {
	if len(out) >= maxArgs || depth > maxDepth {
		return
	}
	switch v := value.(type) {
	case nil:
		if prefix != "" {
			add(out, prefix, "null")
		}
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			flatten(out, joinKey(prefix, key), v[key], depth+1)
			if len(out) >= maxArgs {
				return
			}
		}
	case map[string]string:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			add(out, joinKey(prefix, key), v[key])
			if len(out) >= maxArgs {
				return
			}
		}
	case []any:
		for i, item := range v {
			flatten(out, joinKey(prefix, fmt.Sprintf("%d", i)), item, depth+1)
			if len(out) >= maxArgs {
				return
			}
		}
	case []string:
		for i, item := range v {
			add(out, joinKey(prefix, fmt.Sprintf("%d", i)), item)
			if len(out) >= maxArgs {
				return
			}
		}
	case string:
		add(out, prefix, v)
	case bool:
		add(out, prefix, fmt.Sprintf("%t", v))
	case float64:
		add(out, prefix, fmt.Sprintf("%g", v))
	case float32:
		add(out, prefix, fmt.Sprintf("%g", v))
	case int:
		add(out, prefix, fmt.Sprintf("%d", v))
	case int64:
		add(out, prefix, fmt.Sprintf("%d", v))
	case int32:
		add(out, prefix, fmt.Sprintf("%d", v))
	case uint:
		add(out, prefix, fmt.Sprintf("%d", v))
	case uint64:
		add(out, prefix, fmt.Sprintf("%d", v))
	case uint32:
		add(out, prefix, fmt.Sprintf("%d", v))
	default:
		if prefix != "" {
			add(out, prefix, fmt.Sprint(v))
		}
	}
}

func add(out map[string]string, key string, value string) {
	key = strings.TrimSpace(key)
	if key == "" || len(out) >= maxArgs {
		return
	}
	value = strings.TrimSpace(value)
	if IsSensitiveKey(key) || policy.ContainsSecretPattern(value) {
		out[key] = RedactedValue
		return
	}
	out[key] = truncate(value)
}

func joinKey(prefix string, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return prefix
	}
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func truncate(value string) string {
	runes := []rune(value)
	if len(runes) <= maxValueRunes {
		return value
	}
	return string(runes[:maxValueRunes])
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, ".", "_")
	return value
}

func nilIfEmpty(args map[string]string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	return args
}
