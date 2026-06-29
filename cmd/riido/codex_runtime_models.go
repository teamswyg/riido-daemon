package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func codexRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	home := runtimeModelHome(userHome)
	if home == "" {
		return nil
	}
	modelID := codexConfiguredModelIDFromHome(home)
	if models := codexRuntimeModelsFromCache(home, modelID); len(models) > 0 {
		return models
	}
	model, ok := runtimeModelRecord(modelID, modelID, true)
	if !ok {
		return nil
	}
	return []runtimeactor.RuntimeModel{model}
}

func codexConfiguredModelIDFromHome(home string) string {
	if strings.TrimSpace(home) == "" {
		return ""
	}
	body, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return ""
	}
	return codexModelValueFromConfig(string(body))
}

func codexModelValueFromConfig(body string) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == "model" {
			return parseCodexModelValue(rawValue)
		}
	}
	return ""
}
