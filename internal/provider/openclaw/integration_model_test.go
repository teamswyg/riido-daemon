package openclaw

import "testing"

func TestOpenClawIntegrationModelCanBeOverridden(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, " ollama/fast-local ")
	got := openClawIntegrationModels()
	if len(got) != 1 || got[0] != "ollama/fast-local" {
		t.Fatalf("models=%v", got)
	}
}

func TestOpenClawIntegrationPrefersFastLocalModel(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, "")
	got := openClawIntegrationModels()
	if len(got) < 2 || got[0] != openClawFastIntegrationModel {
		t.Fatalf("models=%v", got)
	}
}
