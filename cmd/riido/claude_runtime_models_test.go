package main

import (
	"slices"
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
	assertClaudeModelIDs(t, models, []string{"claude-fable-5", "fable", "haiku", "opus", "sonnet"})
	if countRuntimeModelDefaults(models) != 0 {
		t.Fatalf("default model count invalid: %+v", models)
	}
}

func TestClaudeRuntimeModelsIncludeBuiltInAliases(t *testing.T) {
	assertClaudeModelIDs(t, claudeRuntimeModels(), []string{"claude-fable-5", "fable", "haiku", "opus", "sonnet"})
}

func TestParseClaudeRuntimeModelHelpRequiresModelOption(t *testing.T) {
	if models := parseClaudeRuntimeModelHelp([]byte("Claude Code help unavailable")); len(models) != 0 {
		t.Fatalf("models = %+v", models)
	}
}

func assertClaudeModelIDs(t *testing.T, models []runtimeactor.RuntimeModel, want []string) {
	t.Helper()
	var ids []string
	for _, model := range models {
		ids = append(ids, model.ModelID)
	}
	if !slices.Equal(ids, want) {
		t.Fatalf("model IDs = %v, want %v", ids, want)
	}
}
