package main

import (
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func runtimeModelRecord(modelID, label string, isDefault bool) (runtimeactor.RuntimeModel, bool) {
	modelID = strings.TrimSpace(modelID)
	label = strings.TrimSpace(label)
	if modelID == "" {
		return runtimeactor.RuntimeModel{}, false
	}
	if label == "" {
		label = modelID
	}
	return runtimeactor.RuntimeModel{ModelID: modelID, Label: label, IsDefault: isDefault}, true
}

func normalizeRuntimeModels(models []runtimeactor.RuntimeModel, defaultID string) []runtimeactor.RuntimeModel {
	out := make([]runtimeactor.RuntimeModel, 0, len(models))
	seen := make(map[string]bool, len(models))
	defaultID = strings.TrimSpace(defaultID)
	for _, model := range models {
		rec, ok := runtimeModelRecord(model.ModelID, model.Label, false)
		if !ok || seen[rec.ModelID] {
			continue
		}
		seen[rec.ModelID] = true
		out = append(out, rec)
	}
	slices.SortFunc(out, func(a, b runtimeactor.RuntimeModel) int {
		return strings.Compare(a.ModelID, b.ModelID)
	})
	return markRuntimeModelDefault(out, defaultID)
}

func markRuntimeModelDefault(models []runtimeactor.RuntimeModel, defaultID string) []runtimeactor.RuntimeModel {
	if len(models) == 0 {
		return nil
	}
	index := 0
	for i := range models {
		models[i].IsDefault = false
		if defaultID != "" && models[i].ModelID == defaultID {
			index = i
		}
	}
	models[index].IsDefault = true
	return models
}
