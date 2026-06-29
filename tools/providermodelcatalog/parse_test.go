package main

import "testing"

func TestParseCursorModels(t *testing.T) {
	rows := parseCursorModels("Available models\n\nauto - Auto (current)\ngpt-5.2 - GPT-5.2\n")
	if len(rows) != 2 || rows[0].ModelID != "cursor-auto" || rows[1].Label != "GPT-5.2" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseClaudeModels(t *testing.T) {
	body := "--model <model> use 'fable', 'opus', or 'sonnet' and 'claude-fable-5'.\n  -n, --name"
	rows := parseClaudeModels(body)
	if len(rows) != 4 || rows[2].ModelID != "sonnet" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseCodexModels(t *testing.T) {
	rows := parseCodexModels([]byte(`{"models":[
		{"slug":"gpt-5.5","display_name":"GPT-5.5"},
		{"slug":"gpt-5.4","display_name":"GPT-5.4"}
	]}`))
	if len(rows) != 2 || rows[0].ModelID != "gpt-5.4" || rows[1].Label != "GPT-5.5" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseOpenClawModels(t *testing.T) {
	rows := parseOpenClawModels([]byte(`{"models":{"providers":{"ollama":{"models":[
		{"id":"llama3.2:latest","name":"Llama 3.2"},
		{"id":"ollama/qwen3:8b","name":"Qwen 3 8B"}
	]}}}}`))
	if len(rows) != 2 || rows[0].ModelID != "ollama/llama3.2:latest" ||
		rows[1].ModelID != "ollama/qwen3:8b" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestValidateCatalogRequiresEveryProviderToHaveMultipleModels(t *testing.T) {
	catalog := catalog{Providers: map[string][]model{
		"codex":    {{ModelID: "a"}, {ModelID: "b"}},
		"cursor":   {{ModelID: "a"}, {ModelID: "b"}},
		"openclaw": {{ModelID: "a"}, {ModelID: "b"}},
		"claude":   {{ModelID: "a"}},
	}}
	if err := validateCatalog(catalog); err == nil {
		t.Fatal("validateCatalog unexpectedly passed with a single claude model")
	}
}
