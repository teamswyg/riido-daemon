package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type cursorConfig struct {
	Model cursorConfigModel `json:"model"`
}

type cursorConfigModel struct {
	DisplayModelID string `json:"displayModelId"`
	DisplayName    string `json:"displayName"`
}

func cursorRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	defaultModel := cursorRuntimeDefaultModel(userHome)
	if models := generatedProviderRuntimeModels(cursorProviderName, defaultModel.ModelID); len(models) > 0 {
		return models
	}
	return cursorRuntimeModelsFallback(defaultModel)
}

func cursorRuntimeModelsFallback(defaultModel runtimeactor.RuntimeModel) []runtimeactor.RuntimeModel {
	if models := cursorRuntimeModelsFromCommand(defaultModel.ModelID); len(models) > 0 {
		return models
	}
	if defaultModel.ModelID != "" {
		return []runtimeactor.RuntimeModel{defaultModel}
	}
	return nil
}

func cursorRuntimeDefaultModel(userHome func() (string, error)) runtimeactor.RuntimeModel {
	if userHome == nil {
		return runtimeactor.RuntimeModel{}
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return runtimeactor.RuntimeModel{}
	}
	body, err := os.ReadFile(filepath.Join(home, ".cursor", "cli-config.json"))
	if err != nil {
		return runtimeactor.RuntimeModel{}
	}
	models := parseCursorRuntimeModels(body)
	if len(models) == 0 {
		return runtimeactor.RuntimeModel{}
	}
	return models[0]
}

func parseCursorRuntimeModels(body []byte) []runtimeactor.RuntimeModel {
	var cfg cursorConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil
	}
	modelID := normalizeCursorRuntimeModelID(cfg.Model.DisplayModelID)
	label := strings.TrimSpace(cfg.Model.DisplayName)
	if label == "" {
		label = modelID
	}
	model, ok := runtimeModelRecord(modelID, label, true)
	if !ok {
		return nil
	}
	return []runtimeactor.RuntimeModel{model}
}
