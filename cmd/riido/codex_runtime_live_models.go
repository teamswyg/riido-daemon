package main

import (
	"context"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

func liveCodexRuntimeModels() []runtimeactor.RuntimeModel {
	catalog, err := codex.ModelCatalog(context.Background(), providerDetectEnvironment(agentbridge.DetectEnv{}))
	if err != nil {
		return nil
	}
	cached := codexRuntimeModels(os.UserHomeDir)
	preferred := ""
	for _, model := range cached {
		if model.IsDefault {
			preferred = model.ModelID
		}
	}
	found := false
	for _, model := range catalog {
		if model.ModelID == preferred {
			found = true
		}
	}
	models := make([]runtimeactor.RuntimeModel, 0, len(catalog))
	for _, model := range catalog {
		isDefault := model.IsDefault
		if found {
			isDefault = model.ModelID == preferred
		}
		models = append(models, runtimeactor.RuntimeModel{ModelID: model.ModelID, Label: model.Label, IsDefault: isDefault})
	}
	return models
}
