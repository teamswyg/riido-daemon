package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type codexModelCache struct {
	Models []codexModelCacheRecord `json:"models"`
}

type codexModelCacheRecord struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
}

func codexRuntimeModelsFromCache(home, defaultID string) []runtimeactor.RuntimeModel {
	body, err := os.ReadFile(filepath.Join(home, ".codex", "models_cache.json"))
	if err != nil {
		return nil
	}
	return parseCodexRuntimeModelsCache(body, defaultID)
}

func parseCodexRuntimeModelsCache(body []byte, defaultID string) []runtimeactor.RuntimeModel {
	var cache codexModelCache
	if err := json.Unmarshal(body, &cache); err != nil {
		return nil
	}
	models := make([]runtimeactor.RuntimeModel, 0, len(cache.Models))
	for _, candidate := range cache.Models {
		model, ok := runtimeModelRecord(candidate.Slug, candidate.DisplayName, false)
		if ok {
			models = append(models, model)
		}
	}
	return normalizeRuntimeModels(models, defaultID)
}
