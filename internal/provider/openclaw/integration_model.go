package openclaw

import (
	"os"
	"strings"
)

const (
	openClawFastSideEffectModel         = "ollama/llama3.2:latest"
	openClawLongContextIntegrationModel = "ollama/qwen3:8b"
	openClawIntegrationModelEnv         = "RIIDO_OPENCLAW_INTEGRATION_MODEL"
)

func IntegrationModelCandidates() []string {
	if model := strings.TrimSpace(os.Getenv(openClawIntegrationModelEnv)); model != "" {
		return []string{model}
	}
	return []string{
		openClawFastSideEffectModel,
		openClawLongContextIntegrationModel,
	}
}

func openClawIntegrationModels() []string {
	return IntegrationModelCandidates()
}
