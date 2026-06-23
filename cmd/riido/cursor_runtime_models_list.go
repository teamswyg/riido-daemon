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
	label = strings.TrimSpace(strings.TrimSuffix(label, "(current)"))
	return runtimeModelRecord(modelID, label, false)
}
