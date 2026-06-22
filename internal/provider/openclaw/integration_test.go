package openclaw

import "testing"

func TestIntegration(t *testing.T) {
	ctx, detect := requireOpenClawIntegration(t)
	req, expected := openClawIntegrationRequest(t, detect)

	obs := runOpenClawIntegrationSession(t, ctx, req, expected.sessionID)
	assertOpenClawIntegrationResult(t, obs)
	assertOpenClawIntegrationArtifact(t, expected)
}
