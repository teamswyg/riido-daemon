package detectutil

import "strings"

// EnvListWithLaunchPATH clones env and appends PATH when env does not already
// contain a PATH entry.
func EnvListWithLaunchPATH(env []string, preferred string) []string {
	out := append([]string(nil), env...)
	if envListHasPATH(out) {
		return out
	}
	preferred = strings.TrimSpace(preferred)
	if preferred == "" {
		preferred = LaunchPATH()
	}
	if preferred != "" {
		out = append(out, pathEnvKey()+"="+preferred)
	}
	return out
}

// EnvListWithLaunchPATHFromMap clones env and appends the frozen PATH value
// from launchEnv when env does not already contain a PATH entry.
func EnvListWithLaunchPATHFromMap(env []string, launchEnv map[string]string) []string {
	out := append([]string(nil), env...)
	if envListHasPATH(out) {
		return out
	}
	key, value, ok := envMapPATHEntry(launchEnv)
	if ok {
		return append(out, key+"="+value)
	}
	if path := LaunchPATH(); path != "" {
		out = append(out, pathEnvKey()+"="+path)
	}
	return out
}
