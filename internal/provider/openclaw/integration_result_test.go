package openclaw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func checkOpenClawIntegrationResult(obs openClawIntegrationObservation) error {
	res := obs.result
	if res.Status != agentbridge.ResultCompleted {
		return fmt.Errorf(
			"openclaw integration did not complete: status=%s error=%q output=%q events=%q",
			res.Status,
			res.Error,
			res.Output,
			openClawFailureEvidence(obs.events),
		)
	}
	if strings.TrimSpace(res.Output) == "" {
		return fmt.Errorf("openclaw integration output is empty")
	}
	return nil
}

func checkOpenClawIntegrationArtifact(
	expected openClawIntegrationExpected,
	obs openClawIntegrationObservation,
) error {
	artifact, err := os.ReadFile(filepath.Join(expected.workdir, expected.artifactName))
	if err != nil {
		return fmt.Errorf(
			"openclaw integration completed without writing expected artifact %q in %q: %w output=%q events=%q",
			expected.artifactName,
			expected.workdir,
			err,
			obs.result.Output,
			openClawFailureEvidence(obs.events),
		)
	}
	if strings.TrimSpace(string(artifact)) != expected.artifactBody {
		return fmt.Errorf(
			"openclaw artifact content = %q, want %q output=%q events=%q",
			string(artifact),
			expected.artifactBody,
			obs.result.Output,
			openClawFailureEvidence(obs.events),
		)
	}
	return nil
}
