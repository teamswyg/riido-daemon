package main

import "testing"

func TestCursorRuntimeModelsUseProviderConfigCurrentModel(t *testing.T) {
	body := []byte(`{"model":{"displayModelId":"auto","displayName":"Auto"}}`)
	models := parseCursorRuntimeModels(body)
	if len(models) != 1 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].ModelID != "cursor-auto" || models[0].Label != "Auto" || !models[0].IsDefault {
		t.Fatalf("current auto model not normalized: %+v", models[0])
	}
}

func TestCursorRuntimeModelsKeepProviderModelID(t *testing.T) {
	body := []byte(`{"model":{"displayModelId":"gpt-5.2","displayName":"GPT-5.2"}}`)
	models := parseCursorRuntimeModels(body)
	if len(models) != 1 || models[0].ModelID != "gpt-5.2" || !models[0].IsDefault {
		t.Fatalf("models = %+v", models)
	}
}

func TestParseCursorRuntimeModelListKeepsMultipleModels(t *testing.T) {
	body := []byte(`Available models

auto - Auto (current)
gpt-5.3-codex - Codex 5.3
gpt-5.2 - GPT-5.2

Tip: use --model <id>`)
	models := parseCursorRuntimeModelList(body, "cursor-auto")
	if len(models) != 3 {
		t.Fatalf("models = %+v", models)
	}
	if models[0].ModelID != "cursor-auto" || !models[0].IsDefault {
		t.Fatalf("default auto model not preserved: %+v", models)
	}
}
