package openclaw

import "strings"

func applyTaskScopedConfig(config map[string]any, workdir, model string) {
	defaults := configMap(configMap(config, "agents"), "defaults")
	defaults["workspace"] = workdir
	if strings.TrimSpace(model) != "" {
		applyTaskScopedModel(defaults, config, model)
	}
}

func applyTaskScopedModel(defaults, config map[string]any, model string) {
	configMap(defaults, "model")["primary"] = model
	for _, entry := range configList(configMap(config, "agents"), "list") {
		if agent, ok := entry.(map[string]any); ok && stringField(agent, "id") == "main" {
			agent["model"] = model
		}
	}
}

func configMap(parent map[string]any, key string) map[string]any {
	child, _ := parent[key].(map[string]any)
	if child == nil {
		child = map[string]any{}
		parent[key] = child
	}
	return child
}

func configList(parent map[string]any, key string) []any {
	list, _ := parent[key].([]any)
	return list
}
