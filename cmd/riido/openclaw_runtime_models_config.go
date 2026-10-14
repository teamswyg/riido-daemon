package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func openClawConfiguredModels(cfg openClawConfig) []runtimeactor.RuntimeModel {
	models := make([]runtimeactor.RuntimeModel, 0)
	for provider, entry := range cfg.Models.Providers {
		models = append(models, openClawProviderModels(provider, entry)...)
	}
	return models
}

func openClawProviderModels(provider string, entry openClawConfigProvider) []runtimeactor.RuntimeModel {
	models := make([]runtimeactor.RuntimeModel, 0, len(entry.Models))
	for _, candidate := range entry.Models {
		model, ok := runtimeModelRecord(
			openClawQualifiedModelID(provider, candidate.ID),
			candidate.Name,
			false,
		)
		if ok {
			models = append(models, model)
		}
	}
	return models
}

func openClawQualifiedModelID(provider, modelID string) string {
	provider = strings.TrimSpace(provider)
	modelID = strings.TrimSpace(modelID)
	if provider == "" || strings.Contains(modelID, "/") {
		return modelID
	}
	return provider + "/" + modelID
}
