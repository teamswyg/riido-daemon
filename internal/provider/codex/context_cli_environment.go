package codex

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// Process inheritance does not survive Codex's core/none shell policy. Supply
// the current run binding on the private thread RPC, never in argv or prompts.
// Dotted config overrides merge with the user's unrelated explicit values.
func contextCLIShellConfig(req agentbridge.StartRequest) map[string]any {
	if !req.ContextCLIEnabled || !agentbridge.ContextCLIAvailable(req.Env) {
		return nil
	}
	values := make(map[string]string, 4)
	for _, key := range []string{contextcli.EnvRunToken, contextcli.EnvDaemonSocket, contextcli.EnvTaskID, contextcli.EnvWorkspaceID} {
		// An empty optional workspace clears any previous thread override.
		values[key] = req.Env[key]
	}
	return map[string]any{"shell_environment_policy.set": values}
}
