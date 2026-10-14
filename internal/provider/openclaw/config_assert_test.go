package openclaw

import (
	"encoding/json"
	"os"
	"testing"
)

func assertTaskScopedConfig(t *testing.T, path, workspace, model string) {
	t.Helper()
	config := readOpenClawConfig(t, path)
	defaults := configMap(configMap(config, "agents"), "defaults")
	if defaults["workspace"] != workspace {
		t.Fatalf("workspace=%v, want %s", defaults["workspace"], workspace)
	}
	if mainModel(config) != model {
		t.Fatalf("main model=%q, want %s", mainModel(config), model)
	}
}

func readOpenClawConfig(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	return config
}

func mainModel(config map[string]any) string {
	for _, entry := range configList(configMap(config, "agents"), "list") {
		if agent, ok := entry.(map[string]any); ok && stringField(agent, "id") == "main" {
			return stringField(agent, "model")
		}
	}
	return ""
}
