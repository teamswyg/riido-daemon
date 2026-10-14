package toolargs

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func add(out map[string]string, key, value string) {
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
