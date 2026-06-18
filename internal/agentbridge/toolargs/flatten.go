package toolargs

import (
	"fmt"
	"sort"
)

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
		flattenAnyMap(out, prefix, v, depth)
	case map[string]string:
		flattenStringMap(out, prefix, v)
	case []any:
		flattenAnySlice(out, prefix, v, depth)
	case []string:
		flattenStringSlice(out, prefix, v)
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

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
