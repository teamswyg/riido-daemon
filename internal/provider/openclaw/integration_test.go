package openclaw

import "testing"

func TestIntegration(t *testing.T) {
	ctx, detect := requireOpenClawIntegration(t)
	var lastErr error
	for _, model := range openClawIntegrationModels() {
		req, expected := openClawIntegrationRequest(t, detect, model)
		obs := runOpenClawIntegrationSession(t, ctx, req, expected.sessionID)
		if err := checkOpenClawIntegrationResult(obs); err != nil {
			lastErr = err
			t.Logf("openclaw integration model %s failed: %v", model, err)
			continue
		}
		if err := checkOpenClawIntegrationArtifact(expected, obs); err != nil {
			lastErr = err
			t.Logf("openclaw integration model %s failed: %v", model, err)
			continue
		}
		t.Logf("openclaw integration model %s passed", model)
		return
	}
	t.Fatalf("openclaw integration failed for all models: %v", lastErr)
}
