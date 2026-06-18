package detectutil

import "strings"

// ResolveExecutable returns the absolute path to the binary.
//
// The override behaves as a pin: a valid override is the only accepted
// candidate, and an invalid override fails closed instead of falling back to
// PATH.
func ResolveExecutable(name, envOverride string) (string, bool) {
	candidates := ResolveExecutableCandidates(name, envOverride)
	if len(candidates) == 0 {
		return "", false
	}
	return candidates[0], true
}

// ResolveExecutableCandidates returns executable candidates in PATH order.
func ResolveExecutableCandidates(name, envOverride string) []string {
	override := strings.TrimSpace(envOverride)
	if override != "" {
		return overrideExecutableCandidate(override)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	return pathExecutableCandidates(name)
}

func overrideExecutableCandidate(override string) []string {
	if !isRegularFile(override) {
		return nil
	}
	return []string{override}
}
