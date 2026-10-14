package openclaw

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	openClawIntegrationArtifactName = "riido-openclaw-side-effect.txt"
	openClawIntegrationArtifactBody = "RIIDO_OPENCLAW_FILESYSTEM_SIDE_EFFECT_OK"
)

type openClawIntegrationExpected struct {
	workdir      string
	artifactName string
	artifactBody string
}

func preseedOpenClawIntegrationWorkspace(t *testing.T, workdir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(workdir, "AGENTS.md"), []byte(integrationAgentInstructions()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func integrationAgentInstructions() string {
	return "Riido integration workspace. Follow the user task exactly.\n"
}
