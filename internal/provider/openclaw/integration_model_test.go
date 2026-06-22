package openclaw

import (
	"reflect"
	"testing"
)

func TestIntegrationModelCandidatesPreferFastLocalModel(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, "")

	got := IntegrationModelCandidates()
	want := []string{
		"ollama/llama3.2:latest",
		"ollama/qwen3:8b",
		"ollama/qwen2.5-coder:1.5b",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IntegrationModelCandidates() = %v, want %v", got, want)
	}
}

func TestIntegrationModelCandidatesAllowsOperatorPin(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, "ollama/custom-fast")

	got := IntegrationModelCandidates()
	want := []string{"ollama/custom-fast"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IntegrationModelCandidates() = %v, want %v", got, want)
	}
}
