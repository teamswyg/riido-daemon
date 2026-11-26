package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func TestParseClaudeRuntimeModelHelpUsesCliAliases(t *testing.T) {
	body := []byte(`  --model <model>                       Model for the current session.
                                        Provide an alias for the latest model
                                        (e.g. 'fable', 'opus', or 'sonnet')
                                        or a model's full name
                                        (e.g. 'claude-fable-5').
  -n, --name <name>                     Set a display name`)
	models := parseClaudeRuntimeModelHelp(body)
	assertClaudeModels(t, models)
	if countRuntimeModelDefaults(models) != 0 {
		t.Fatalf("default model count invalid: %+v", models)
	}
}

func TestClaudeRuntimeModelsIncludeBuiltInAliases(t *testing.T) {
	assertClaudeModels(t, claudeRuntimeModels())
}

func TestParseClaudeRuntimeModelHelpRequiresModelOption(t *testing.T) {
	if models := parseClaudeRuntimeModelHelp([]byte("Claude Code help unavailable")); len(models) != 0 {
		t.Fatalf("models = %+v", models)
	}
}

func assertClaudeModels(t *testing.T, models []runtimeactor.RuntimeModel) {
	t.Helper()
	want := []struct{ id, label string }{
		{"claude-fable-5", "Claude Fable 5"},
		{"fable", "Fable 5"},
		{"haiku", "Haiku 4.5"},
		{"opus", "Opus 4.8"},
		{"sonnet", "Sonnet 5"},
	}
	if len(models) != len(want) {
		t.Fatalf("models = %+v, want %+v", models, want)
	}
	for i, model := range models {
		if model.ModelID != want[i].id || model.Label != want[i].label {
			t.Errorf("model %d = %q (%q), want %q (%q)", i, model.ModelID, model.Label, want[i].id, want[i].label)
		}
	}
}
