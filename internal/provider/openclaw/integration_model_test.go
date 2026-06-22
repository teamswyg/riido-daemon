package openclaw

import "testing"

func TestOpenClawIntegrationModelCanBeOverridden(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, " ollama/fast-local ")
	got := openClawIntegrationModels()
	if len(got) != 1 || got[0] != "ollama/fast-local" {
		t.Fatalf("models=%v", got)
	}
}

func TestOpenClawIntegrationPrefersStableLocalModel(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, "")
	got := openClawIntegrationModels()
	if len(got) != 3 || got[0] != openClawStableIntegrationModel {
		t.Fatalf("models=%v", got)
	}
	if got[1] != openClawFastIntegrationModel {
		t.Fatalf("models=%v", got)
	}
}
