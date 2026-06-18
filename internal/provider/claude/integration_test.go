package claude

import "testing"

// TestIntegration spawns the real Claude Code CLI and runs a trivial
// prompt. Skipped unless AGENTBRIDGE_INTEGRATION=1 is set AND the
// `claude` binary is on $PATH. This matches spec §10 Phase 6 / §6.X
// (Dev/Prod Parity).
func TestIntegration(t *testing.T) {
	ctx := claudeIntegrationContext(t)
	workdir := t.TempDir()
	req := claudeIntegrationRequest(workdir)

	sess := startClaudeIntegrationSession(t, ctx, req)
	drainClaudeIntegrationEvents(sess)

	res := <-sess.Result()
	requireClaudeIntegrationCompleted(t, res)
	assertClaudeIntegrationArtifact(t, workdir)
}
