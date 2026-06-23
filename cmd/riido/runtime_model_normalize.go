package main

import (
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
