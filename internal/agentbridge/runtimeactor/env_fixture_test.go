package runtimeactor

import "strings"

func envListValue(env []string, wantKey string) (string, bool) {
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, wantKey) {
			return value, true
		}
	}
	return "", false
}
