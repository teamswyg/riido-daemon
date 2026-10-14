package openclaw

import (
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	ctx, detect := requireOpenClawIntegration(t)
	var lastErr error
	for _, model := range openClawIntegrationModels() {
		req, expected := openClawIntegrationRequest(t, detect, model)
		obs := runOpenClawIntegrationSession(t, ctx, req)
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
		if !strings.Contains(obs.result.Output, "RIIDO_EXECUTION_ACK_OPENCLAW") {
			t.Fatalf("runtime instruction not followed: %q", obs.result.Output)
		}
		t.Logf("openclaw integration model %s passed", model)
		return
	}
	t.Fatalf("openclaw integration failed for all models: %v", lastErr)
}
