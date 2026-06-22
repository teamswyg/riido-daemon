package openclaw

import (
	"os"
	"strings"
)

const (
	openClawFastIntegrationModel     = "ollama/qwen2.5-coder:1.5b"
	openClawDefaultIntegrationModel  = "ollama/llama3.2:latest"
	openClawFallbackIntegrationModel = "ollama/qwen3:8b"
	openClawIntegrationModelEnv      = "RIIDO_OPENCLAW_INTEGRATION_MODEL"
)

func IntegrationModelCandidates() []string {
	if model := strings.TrimSpace(os.Getenv(openClawIntegrationModelEnv)); model != "" {
		return []string{model}
	}
	return []string{
		openClawFastIntegrationModel,
		openClawDefaultIntegrationModel,
		openClawFallbackIntegrationModel,
	}
}

func openClawIntegrationModels() []string {
	return IntegrationModelCandidates()
}
