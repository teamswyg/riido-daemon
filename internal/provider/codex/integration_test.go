package codex

import "testing"

func TestIntegration(t *testing.T) {
	ctx := requireCodexIntegration(t)
	req, expected := codexIntegrationRequest(t)

	res := runCodexIntegrationSession(t, ctx, req)
	if res.Status != expected.status {
		t.Fatalf("codex integration did not complete: %+v", res)
	}
	assertCodexIntegrationArtifact(t, expected)
}
