package contextbroker

import (
	"os"
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// LaunchEnv issues a capability for binding and writes the CLI wiring into
// env, the provider launch environment for that run. env must be a map the
// caller owns. On failure env is left without a run capability, so the CLI
// falls back to user mode and reports USER_MODE_UNAVAILABLE.
func (b *Broker) LaunchEnv(binding Binding, env map[string]string) error {
	token, err := b.Registry.Issue(binding)
	if err != nil {
		return err
	}
	env[contextcli.EnvRunToken] = token
	if b.Socket != "" {
		env[contextcli.EnvDaemonSocket] = b.Socket
	}
	setIfPresent(env, contextcli.EnvTaskID, binding.TaskID)
	setIfPresent(env, contextcli.EnvWorkspaceID, binding.WorkspaceID)
	prependPATH(env, b.BinDir)
	return nil
}

func setIfPresent(env map[string]string, key, value string) {
	if value != "" {
		env[key] = value
	}
}

// prependPATH puts dir first on PATH (moving it if already present), keeping the existing key spelling
// (Windows uses Path).
func prependPATH(env map[string]string, dir string) {
	if dir == "" {
		return
	}
	key := "PATH"
	for existing := range env {
		if strings.EqualFold(existing, "PATH") {
			key = existing
			break
		}
	}
	// Drop later occurrences so an older riido earlier on PATH cannot win.
	entries := slices.DeleteFunc(strings.Split(env[key], string(os.PathListSeparator)), func(entry string) bool {
		return entry == dir || entry == ""
	})
	env[key] = strings.Join(append([]string{dir}, entries...), string(os.PathListSeparator))
}
