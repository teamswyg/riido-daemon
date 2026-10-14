package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// TestIntegration spawns the real Claude Code CLI and runs a trivial
// prompt. Skipped unless AGENTBRIDGE_INTEGRATION=1 is set AND the
// `claude` binary is on $PATH. This matches spec §10 Phase 6 / §6.X
// (Dev/Prod Parity).
func TestIntegration(t *testing.T) {
	ctx := claudeIntegrationContext(t)
	workdir := t.TempDir()
	req := claudeIntegrationRequest(workdir)
	req.Prompt = strings.Split(req.Prompt, "After the file is written,")[0]
	req.Prompt, req.SystemPrompt, _, _ = agentbridge.ApplyRuntimeInstructionContract(Name, req.Prompt, "", "After completing the file write, include RIIDO_EXECUTION_ACK_CLAUDE in your final reply.")

	sess := startClaudeIntegrationSession(t, ctx, req)
	drainClaudeIntegrationEvents(sess)

	res := <-sess.Result()
	requireClaudeIntegrationCompleted(t, res)
	assertClaudeIntegrationArtifact(t, workdir)
	if !strings.Contains(res.Output, "RIIDO_EXECUTION_ACK_CLAUDE") {
		t.Fatalf("native instruction not followed: %q", res.Output)
	}
}
