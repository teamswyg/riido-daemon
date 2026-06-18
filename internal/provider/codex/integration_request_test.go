package codex

import (
	"os"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const (
	codexIntegrationArtifactName = "riido-codex-side-effect.txt"
	codexIntegrationArtifactBody = "RIIDO_CODEX_FILESYSTEM_SIDE_EFFECT_OK"
)

type codexIntegrationExpected struct {
	status       agentbridge.ResultStatus
	workdir      string
	artifactName string
	artifactBody string
}

func codexIntegrationRequest(
	t *testing.T,
) (agentbridge.StartRequest, codexIntegrationExpected) {
	t.Helper()
	workdir := t.TempDir()
	req := agentbridge.StartRequest{
		Prompt: codexIntegrationPrompt(),
		Cwd:    workdir,
		Env:    codexIntegrationEnv(),
	}
	return req, codexIntegrationExpected{
		status:       agentbridge.ResultCompleted,
		workdir:      workdir,
		artifactName: codexIntegrationArtifactName,
		artifactBody: codexIntegrationArtifactBody,
	}
}

func codexIntegrationPrompt() string {
	return `In the current working directory, create a file named ` +
		codexIntegrationArtifactName +
		` with exactly this content and no trailing commentary in the file: ` +
		codexIntegrationArtifactBody + `

After the file is written, respond with exactly "ok".`
}

func codexIntegrationEnv() map[string]string {
	env := map[string]string{}
	if value := os.Getenv("CODEX_HOME"); value != "" {
		env["CODEX_HOME"] = value
	}
	if value := os.Getenv("HOME"); value != "" {
		env["HOME"] = value
	}
	return env
}
