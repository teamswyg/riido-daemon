package main

import "testing"

func TestOpenClawRuntimeModelsUseProviderConfigDefault(t *testing.T) {
	body := []byte(`{
		"agents": {"defaults": {"model": {"primary": "ollama/llama3.2:latest"}}},
		"models": {"providers": {"ollama": {"models": [
			{"id":"qwen3:8b","name":"Qwen 3 8B"},
			{"id":"llama3.2:latest","name":"Llama 3.2"}
		]}}}
	}`)
	models := parseOpenClawRuntimeModels(body)
	if len(models) != 2 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].ModelID != "ollama/llama3.2:latest" ||
		models[0].Label != "Llama 3.2" || !models[0].IsDefault {
		t.Fatalf("resolved default lost: %+v", models)
	}
}

func TestOpenClawRuntimeModelsDoNotInventDefault(t *testing.T) {
	models := parseOpenClawRuntimeModels([]byte(`{"agents":{"defaults":{"model":{}}}}`))
	if len(models) != 0 {
		t.Fatalf("models = %+v", models)
	}
}
