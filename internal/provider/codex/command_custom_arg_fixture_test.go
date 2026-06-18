package codex

import "strings"

func containsAnyCommandArg(args []string, tokens ...string) bool {
	joined := strings.Join(args, " ")
	for _, token := range tokens {
		if strings.Contains(joined, token) {
			return true
		}
	}
	return false
}
