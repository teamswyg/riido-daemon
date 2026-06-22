package openclaw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func assertOpenClawIntegrationResult(t *testing.T, obs openClawIntegrationObservation) {
	t.Helper()
	res := obs.result
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf(
			"openclaw integration did not complete: status=%s error=%q output=%q events=%q",
			res.Status,
			res.Error,
			res.Output,
			openClawFailureEvidence(obs.events),
		)
	}
	if strings.TrimSpace(res.Output) == "" {
		t.Fatalf("openclaw integration output is empty")
	}
}

func assertOpenClawIntegrationArtifact(t *testing.T, expected openClawIntegrationExpected) {
	t.Helper()
	artifact, err := os.ReadFile(filepath.Join(expected.workdir, expected.artifactName))
	if err != nil {
		t.Fatalf(
			"openclaw integration completed without writing expected artifact %q in %q: %v",
			expected.artifactName,
			expected.workdir,
			err,
		)
	}
	if strings.TrimSpace(string(artifact)) != expected.artifactBody {
		t.Fatalf("openclaw artifact content = %q, want %q", string(artifact), expected.artifactBody)
	}
}
