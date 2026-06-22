package openclaw

import (
	"os"
	"strings"
)

const (
	openClawFastIntegrationModel   = "ollama/llama3.2:latest"
	openClawStableIntegrationModel = "ollama/qwen3:8b"
	openClawLegacyIntegrationModel = "ollama/qwen2.5-coder:1.5b"
	openClawIntegrationModelEnv    = "RIIDO_OPENCLAW_INTEGRATION_MODEL"
)

func IntegrationModelCandidates() []string {
	if model := strings.TrimSpace(os.Getenv(openClawIntegrationModelEnv)); model != "" {
		return []string{model}
	}
	return []string{
		openClawFastIntegrationModel,
		openClawStableIntegrationModel,
		openClawLegacyIntegrationModel,
	}
}

func openClawIntegrationModels() []string {
	return IntegrationModelCandidates()
}
