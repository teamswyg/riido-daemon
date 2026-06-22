package openclaw

import (
	"os"
	"strings"
	"testing"
)

const (
	defaultOpenClawIntegrationModel  = "ollama/llama3.2:latest"
	fallbackOpenClawIntegrationModel = "ollama/qwen3:8b"
	openClawIntegrationModelEnv      = "RIIDO_OPENCLAW_INTEGRATION_MODEL"
)

func openClawIntegrationModels() []string {
	if model := strings.TrimSpace(os.Getenv(openClawIntegrationModelEnv)); model != "" {
		return []string{model}
	}
	return []string{defaultOpenClawIntegrationModel, fallbackOpenClawIntegrationModel}
}

func TestOpenClawIntegrationModelCanBeOverridden(t *testing.T) {
	t.Setenv(openClawIntegrationModelEnv, " ollama/fast-local ")
	got := openClawIntegrationModels()
	if len(got) != 1 || got[0] != "ollama/fast-local" {
		t.Fatalf("models=%v", got)
	}
}
