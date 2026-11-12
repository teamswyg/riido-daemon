package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

// The user's inherit=core policy intentionally omits arbitrary process env.
// Explicitly pass only this run's broker wiring through the native config,
// keeping capabilities out of model-visible instructions and config files.
func contextCLIConfig(req agentbridge.StartRequest) map[string]any {
	if !req.ContextCLIEnabled || !agentbridge.ContextCLIAvailable(req.Env) {
		return nil
	}
	values := make(map[string]any, 4)
	for _, key := range []string{"RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID", "RIIDO_WORKSPACE_ID"} {
		if value := req.Env[key]; value != "" {
			values["shell_environment_policy.set."+key] = value
		}
	}
	return values
}
