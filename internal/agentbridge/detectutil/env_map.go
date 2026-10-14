package detectutil

import "maps"

// EnvMapWithLaunchPATH clones env and adds PATH when env does not already
// contain an explicit PATH key.
func EnvMapWithLaunchPATH(env map[string]string) map[string]string {
	out := make(map[string]string, len(env)+1)
	maps.Copy(out, env)
	if envMapHasPATH(out) {
		return out
	}
	if path := LaunchPATH(); path != "" {
		out[pathEnvKey()] = path
	}
	return out
}

// EnvMapPATHValue returns the PATH-like value from env, if one is present.
func EnvMapPATHValue(env map[string]string) string {
	_, value, _ := envMapPATHEntry(env)
	return value
}

func envMapHasPATH(env map[string]string) bool {
	_, _, ok := envMapPATHEntry(env)
	return ok
}
