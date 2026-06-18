package openclaw

import "testing"

func TestIntegration(t *testing.T) {
	ctx, detect := requireOpenClawIntegration(t)
	req, expected := openClawIntegrationRequest(t, detect)

	res := runOpenClawIntegrationSession(t, ctx, req, expected.sessionID)
	assertOpenClawIntegrationResult(t, res)
	assertOpenClawIntegrationArtifact(t, expected)
}
