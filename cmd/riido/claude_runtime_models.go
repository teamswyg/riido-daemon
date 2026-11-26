package main

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	claudeprovider "github.com/teamswyg/riido-daemon/internal/provider/claude"
)

func claudeRuntimeModels() []runtimeactor.RuntimeModel {
	if models := generatedProviderRuntimeModels(claudeprovider.Name, ""); len(models) > 0 {
		return models
	}
	body := runtimeModelCommandOutput(claudeprovider.DefaultExecutable, "--help")
	return parseClaudeRuntimeModelHelp(body)
}

func parseClaudeRuntimeModelHelp(body []byte) []runtimeactor.RuntimeModel {
	ids := claudeprovider.ModelIDsFromHelp(string(body))
	models := make([]runtimeactor.RuntimeModel, 0, len(ids))
	for _, id := range ids {
		model, _ := runtimeModelRecord(id, claudeprovider.ModelLabel(id), false)
		models = append(models, model)
	}
	return normalizeRuntimeModels(models, "")
}
