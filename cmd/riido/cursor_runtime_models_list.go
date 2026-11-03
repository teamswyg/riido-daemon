package main

import (
	"bufio"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func parseCursorRuntimeModelList(body []byte, defaultID string) []runtimeactor.RuntimeModel {
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	models := make([]runtimeactor.RuntimeModel, 0)
	for scanner.Scan() {
		model, ok := parseCursorRuntimeModelLine(scanner.Text())
		if ok {
			models = append(models, model)
			if defaultID == "" && model.IsDefault {
				defaultID = model.ModelID
			}
		}
	}
	return normalizeRuntimeModels(models, normalizeCursorRuntimeModelID(defaultID))
}

func parseCursorRuntimeModelLine(line string) (runtimeactor.RuntimeModel, bool) {
	modelID, label, ok := strings.Cut(strings.TrimSpace(line), " - ")
	if !ok {
		return runtimeactor.RuntimeModel{}, false
	}
	modelID = normalizeCursorRuntimeModelID(modelID)
	isDefault := strings.Contains(label, "(current") || strings.HasSuffix(strings.TrimSpace(label), "(default)")
	for _, suffix := range []string{"(current, default)", "(current)", "(default)"} {
		label = strings.TrimSpace(strings.TrimSuffix(label, suffix))
	}
	return runtimeModelRecord(modelID, label, isDefault)
}
