package main

import (
	"embed"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

//go:embed provider_model_catalog.generated.json
var providerModelCatalogFS embed.FS

type generatedProviderModelCatalog struct {
	Providers map[string][]generatedProviderModel `json:"providers"`
}

type generatedProviderModel struct {
	ModelID string `json:"model_id"`
	Label   string `json:"label"`
}

func generatedProviderRuntimeModels(provider, defaultID string) []runtimeactor.RuntimeModel {
	body, err := providerModelCatalogFS.ReadFile("provider_model_catalog.generated.json")
	if err != nil {
		return nil
	}
	var catalog generatedProviderModelCatalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return nil
	}
	models := make([]runtimeactor.RuntimeModel, 0, len(catalog.Providers[provider]))
	for _, candidate := range catalog.Providers[provider] {
		model, ok := runtimeModelRecord(candidate.ModelID, candidate.Label, false)
		if ok {
			models = append(models, model)
		}
	}
	return normalizeRuntimeModels(models, defaultID)
}
