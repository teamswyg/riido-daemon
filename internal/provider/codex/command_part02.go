package codex

import (
	"fmt"
	"maps"
	"sort"
)

// buildEnv merges caller env with adapter-reserved env. Reserved keys
// always win — caller values for those keys are silently dropped. We
// don't surface the drop as a Warning event because env collisions are
// expected (the daemon may pass through user $PATH etc. that contains
// no secrets, and adding a warning per env collision would be noisy).
// If we later need observability, switch to returning a separate
// DroppedEnvKeys slice.
func buildEnv(caller map[string]string, _ StartOptions) []string {
	reserved := map[string]string{}

	merged := make(map[string]string, len(caller)+len(reserved))
	for k, v := range caller {
		if _, isReserved := reserved[k]; isReserved {
			continue
		}
		merged[k] = v
	}
	maps.Copy(merged, reserved)

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, k := range keys {
		env = append(env, fmt.Sprintf("%s=%s", k, merged[k]))
	}
	return env
}
