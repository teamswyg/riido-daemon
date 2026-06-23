package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCodexRuntimeModelsReadConfiguredDefaultModel(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("model = \"gpt-5.5\"\nmodel_reasoning_effort = \"xhigh\"\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), config, 0o600); err != nil {
		t.Fatal(err)
	}
	cache := []byte(`{"models":[
		{"slug":"gpt-5.4","display_name":"GPT-5.4"},
		{"slug":"gpt-5.5","display_name":"GPT-5.5"}
	]}`)
	if err := os.WriteFile(filepath.Join(configDir, "models_cache.json"), cache, 0o600); err != nil {
		t.Fatal(err)
	}
	models := codexRuntimeModels(func() (string, error) { return home, nil })
	if len(models) != 2 || models[1].ModelID != "gpt-5.5" ||
		models[1].Label != "GPT-5.5" || !models[1].IsDefault {
		t.Fatalf("models = %+v", models)
	}
}

func TestCodexRuntimeModelsFallBackToConfiguredModel(t *testing.T) {
	home := t.TempDir()
	configDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	config := []byte("model = \"gpt-5.5\"\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), config, 0o600); err != nil {
		t.Fatal(err)
	}
	models := codexRuntimeModels(func() (string, error) { return home, nil })
	if len(models) != 1 || models[0].ModelID != "gpt-5.5" || !models[0].IsDefault {
		t.Fatalf("models = %+v", models)
	}
}

func TestCodexRuntimeModelsMissingConfigDoesNotInventModel(t *testing.T) {
	models := codexRuntimeModels(func() (string, error) { return t.TempDir(), nil })
	if len(models) != 0 {
		t.Fatalf("models = %+v", models)
	}
}
