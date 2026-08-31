package openclaw

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func openClawIntegrationRequest(
	t *testing.T,
	detect agentbridge.DetectResult,
	model string,
) (agentbridge.StartRequest, openClawIntegrationExpected) {
	t.Helper()
	workdir := t.TempDir()
	preseedOpenClawIntegrationWorkspace(t, workdir)
	req := agentbridge.StartRequest{
		Prompt:     openClawIntegrationPrompt(),
		Cwd:        workdir,
		Executable: detect.Executable,
		Model:      model,
		CustomArgs: []string{"--thinking", "off"},
	}
	return req, openClawIntegrationExpected{
		workdir:      workdir,
		artifactName: openClawIntegrationArtifactName,
		artifactBody: openClawIntegrationArtifactBody,
	}
}

func openClawIntegrationPrompt() string {
	return `Use the write tool exactly once.
Path: ` + openClawIntegrationArtifactName + `
Content:
` + openClawIntegrationArtifactBody + `

Do not add quotes, punctuation, spaces, or a trailing newline.
After writing the file, respond with exactly "ok".`
}
