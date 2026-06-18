package openclaw

import (
	"strconv"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	openClawIntegrationArtifactName = "riido-openclaw-side-effect.txt"
	openClawIntegrationArtifactBody = "RIIDO_OPENCLAW_FILESYSTEM_SIDE_EFFECT_OK"
)

type openClawIntegrationExpected struct {
	sessionID    string
	workdir      string
	artifactName string
	artifactBody string
}

func openClawIntegrationRequest(
	t *testing.T,
	detect agentbridge.DetectResult,
) (agentbridge.StartRequest, openClawIntegrationExpected) {
	t.Helper()
	sessionID := "integration-openclaw-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	workdir := t.TempDir()
	req := agentbridge.StartRequest{
		Prompt:     openClawIntegrationPrompt(),
		Cwd:        workdir,
		Executable: detect.Executable,
	}
	return req, openClawIntegrationExpected{
		sessionID:    sessionID,
		workdir:      workdir,
		artifactName: openClawIntegrationArtifactName,
		artifactBody: openClawIntegrationArtifactBody,
	}
}

func openClawIntegrationPrompt() string {
	return `In the current working directory, create a file named ` +
		openClawIntegrationArtifactName +
		` with exactly this content and no trailing commentary in the file: ` +
		openClawIntegrationArtifactBody + `

After the file is written, respond with exactly "ok".`
}
