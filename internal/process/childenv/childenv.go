// Package childenv builds the environment for processes the daemon starts
// on behalf of providers, validation checks, and repository preparation.
//
// Daemon-only credentials never reach those processes. The daemon reads
// RIIDO_DEVICE_SECRET at startup to authenticate to control-plane; agent
// runtimes, provider probes, and task commands run without it.
package childenv

import (
	"os"
	"slices"
	"strings"
)

// daemonOnlySecretKeys are removed from every child environment. Matching is
// case-insensitive because Windows environment names are.
var daemonOnlySecretKeys = []string{"RIIDO_DEVICE_SECRET", "RIIDO_SAAS_TOKEN", "RIIDO_TOKEN"}

// IsDaemonSecret reports whether key names a daemon-only credential.
func IsDaemonSecret(key string) bool {
	return slices.ContainsFunc(daemonOnlySecretKeys, func(secret string) bool {
		return strings.EqualFold(key, secret)
	})
}

// Scrub returns env without daemon-only credentials.
func Scrub(env []string) []string {
	out := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if IsDaemonSecret(key) {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// Environ returns the daemon environment without daemon-only credentials.
// Use it instead of leaving exec.Cmd.Env nil, which inherits everything.
func Environ() []string {
	return Scrub(os.Environ())
}
