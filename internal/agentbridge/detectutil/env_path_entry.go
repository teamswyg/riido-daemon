package detectutil

import "strings"

func envMapPATHEntry(env map[string]string) (string, string, bool) {
	for key, value := range env {
		if strings.EqualFold(key, pathEnvKey()) {
			return key, value, true
		}
	}
	return "", "", false
}

func envListHasPATH(env []string) bool {
	for _, entry := range env {
		key, _, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(key, pathEnvKey()) {
			return true
		}
	}
	return false
}
