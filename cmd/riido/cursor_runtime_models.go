package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
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
	if userHome == nil {
		return nil
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return nil
	}
	body, err := os.ReadFile(filepath.Join(home, ".cursor", "cli-config.json"))
	if err != nil {
		return nil
	}
	return parseCursorRuntimeModels(body)
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

func normalizeCursorRuntimeModelID(modelID string) string {
	if strings.TrimSpace(modelID) == "auto" {
		return providercatalog.DefaultCursorModelID
	}
	return strings.TrimSpace(modelID)
}
