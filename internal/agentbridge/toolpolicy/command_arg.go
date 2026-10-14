package toolpolicy

import "strings"

func commandArg(args map[string]string) (string, bool) {
	for _, key := range []string{"command", "cmd", "script", "input.command"} {
		if value, ok := args[key]; ok {
			return value, true
		}
	}
	for key, value := range args {
		normalized := normalizeToolToken(key)
		if normalized == "command" || normalized == "cmd" || strings.HasSuffix(normalized, "_command") {
			return value, true
		}
	}
	return "", false
}
