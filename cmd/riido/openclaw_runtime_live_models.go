package main

import (
	"context"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

func liveOpenClawRuntimeModels() []runtimeactor.RuntimeModel {
	catalog, err := openclaw.ModelCatalog(context.Background(), providerDetectEnvironment(agentbridge.DetectEnv{}))
	if err != nil {
		return openClawRuntimeModels(os.UserHomeDir)
	}
	var models []runtimeactor.RuntimeModel
	defaultID := ""
	for _, row := range catalog {
		if row.IsDefault {
			defaultID = row.ModelID
		}
		models = append(models, runtimeactor.RuntimeModel{ModelID: row.ModelID, Label: row.Label, IsDefault: row.IsDefault})
	}
	return normalizeRuntimeModels(models, defaultID)
}
