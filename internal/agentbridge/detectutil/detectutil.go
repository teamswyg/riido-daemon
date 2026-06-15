// Package detectutil owns the small C4 helper surface concrete provider
// adapters use to implement Detect: PATH lookup with env overrides and a
// short-running version probe.
//
// It does not own provider capability classification, provider-specific
// parsing, process spawning for runs, or scheduling decisions. We isolate
// it here so the agentbridge port doesn't grow an `os/exec` dependency
// and each provider package stays focused on its own command-line shape.
package detectutil

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const versionProbeTimeout = 10 * time.Second

// loginShellPATHTimeout bounds the one-shot login-shell PATH probe so a slow
// or misbehaving shell profile can never hang Detect.
const loginShellPATHTimeout = 3 * time.Second

// Markers wrap $PATH in the login-shell probe so noise printed by profile
// scripts (banners, version managers) is excluded from the captured value.
const (
	loginPATHMarkerStart = "__RIIDO_PATH_START__"
	loginPATHMarkerEnd   = "__RIIDO_PATH_END__"
)

// ResolveExecutable returns the absolute path to the binary.
//
// The override behaves as a PIN, not a hint:
//
//  1. If envOverride is non-empty and points to a regular file →
//     use it.
//  2. If envOverride is non-empty but the path does NOT exist or is a
//     directory → ("", false). We do NOT silently fall back to PATH;
//     doing so would let a misconfigured RIIDO_*_PATH appear to "work"
//     against a different binary than the operator intended.
//  3. If envOverride is empty → exec.LookPath(name).
//
// Returns (path, true) on success, ("", false) if not found.
func ResolveExecutable(name, envOverride string) (string, bool) {
	candidates := ResolveExecutableCandidates(name, envOverride)
	if len(candidates) == 0 {
		return "", false
	}
	return candidates[0], true
}

// ResolveExecutableCandidates returns executable candidates in PATH order.
//
// It preserves the same env override pin semantics as ResolveExecutable:
// an override returns at most that one file, and an invalid override
// returns no candidates. With no override, the first element matches the
// normal exec.LookPath result, followed by any later same-name PATH hits.
func ResolveExecutableCandidates(name, envOverride string) []string {
	override := strings.TrimSpace(envOverride)
	if override != "" {
		if !isRegularFile(override) {
			return nil
		}
		return []string{override}
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}

	var candidates []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		key := filepath.Clean(p)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, p)
	}

	p, err := exec.LookPath(name)
	if err == nil {
		add(p)
	}

	if strings.ContainsAny(name, `/\`) {
		return candidates
	}

	for _, dir := range augmentedSearchDirs() {
		if dir == "" {
			dir = "."
		}
		for _, candidateName := range executableNames(name) {
			path := filepath.Join(dir, candidateName)
			if isExecutableFile(path) {
				add(path)
			}
		}
	}

	return candidates
}

// LaunchPATH returns the PATH value provider child processes should inherit
// when the caller has not supplied an explicit PATH. It uses the same augmented
// search directories as executable detection so a GUI-launched daemon can both
// find provider CLIs and let those CLIs find child tools such as git/node.
func LaunchPATH() string {
	return strings.Join(augmentedSearchDirs(), string(os.PathListSeparator))
}

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

// EnvListWithLaunchPATH clones env and appends PATH when env does not already
// contain a PATH entry. The preferred value should normally come from
// EnvMapWithLaunchPATH so adapter BuildStart and final process spawn share the
// same frozen launch path.
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

func envMapHasPATH(env map[string]string) bool {
	_, _, ok := envMapPATHEntry(env)
	return ok
}

func envMapPATHEntry(env map[string]string) (string, string, bool) {
	for key, value := range env {
		if strings.EqualFold(key, pathEnvKey()) {
			return key, value, true
		}
	}
	return "", "", false
}
