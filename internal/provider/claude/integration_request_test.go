package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	claudeIntegrationArtifactName = "riido-claude-side-effect.txt"
	claudeIntegrationArtifactBody = "RIIDO_CLAUDE_FILESYSTEM_SIDE_EFFECT_OK"
)

func claudeIntegrationRequest(workdir string) agentbridge.StartRequest {
	return agentbridge.StartRequest{
		Prompt: `In the current working directory, create a file named ` +
			claudeIntegrationArtifactName +
			` with exactly this content and no trailing commentary in the file: ` +
			claudeIntegrationArtifactBody +
			`

After the file is written, respond with exactly "ok".`,
		Cwd: workdir,
	}
}

func assertClaudeIntegrationArtifact(t *testing.T, workdir string) {
	t.Helper()

	path := filepath.Join(workdir, claudeIntegrationArtifactName)
	artifact, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"claude integration completed without writing expected artifact %q in %q: %v",
			claudeIntegrationArtifactName,
			workdir,
			err,
		)
	}
	if strings.TrimSpace(string(artifact)) != claudeIntegrationArtifactBody {
		t.Fatalf("claude artifact content = %q, want %q", string(artifact), claudeIntegrationArtifactBody)
	}
}
