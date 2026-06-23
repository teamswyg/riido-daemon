package main

import "testing"

func TestOpenClawRuntimeModelsUseProviderConfigDefault(t *testing.T) {
	body := []byte(`{
		"agents": {"defaults": {"model": {"primary": "ollama/llama3.2:latest"}}}
	}`)
	models := parseOpenClawRuntimeModels(body)
	if len(models) != 1 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].ModelID != "ollama/llama3.2:latest" ||
		models[0].Label != "ollama/llama3.2:latest" ||
		!models[0].IsDefault {
		t.Fatalf("resolved default lost: %+v", models[0])
	}
}

func TestOpenClawRuntimeModelsDoNotInventDefault(t *testing.T) {
	models := parseOpenClawRuntimeModels([]byte(`{"agents":{"defaults":{"model":{}}}}`))
	if len(models) != 0 {
		t.Fatalf("models = %+v", models)
	}
}
