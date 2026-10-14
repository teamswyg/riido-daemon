package toolargs

import "fmt"

func flattenAnyMap(out map[string]string, prefix string, values map[string]any, depth int) {
	for _, key := range sortedKeys(values) {
		flatten(out, joinKey(prefix, key), values[key], depth+1)
		if len(out) >= maxArgs {
			return
		}
	}
}

func flattenStringMap(out map[string]string, prefix string, values map[string]string) {
	for _, key := range sortedKeys(values) {
		add(out, joinKey(prefix, key), values[key])
		if len(out) >= maxArgs {
			return
		}
	}
}

func flattenAnySlice(out map[string]string, prefix string, values []any, depth int) {
	for i, item := range values {
		flatten(out, joinKey(prefix, fmt.Sprintf("%d", i)), item, depth+1)
		if len(out) >= maxArgs {
			return
		}
	}
}

func flattenStringSlice(out map[string]string, prefix string, values []string) {
	for i, item := range values {
		add(out, joinKey(prefix, fmt.Sprintf("%d", i)), item)
		if len(out) >= maxArgs {
			return
		}
	}
}
