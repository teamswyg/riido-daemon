package processexec

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/process/childenv"
)

// mergeEnv returns the child environment: the daemon environment without
// daemon-only secrets, with caller overrides applied. It always returns an
// explicit list so a nil result can never fall back to full inheritance.
func mergeEnv(overrides []string) []string {
	env := childenv.Environ()
	indexByKey := make(map[string]int, len(env)+len(overrides))
	for i, entry := range env {
		key, _, ok := strings.Cut(entry, "=")
		if ok {
			indexByKey[key] = i
		}
	}
	for _, entry := range childenv.Scrub(overrides) {
		key, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if i, exists := indexByKey[key]; exists {
			env[i] = entry
			continue
		}
		indexByKey[key] = len(env)
		env = append(env, entry)
	}
	return env
}
