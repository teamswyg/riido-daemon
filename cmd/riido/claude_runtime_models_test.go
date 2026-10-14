package main

import "testing"

func TestParseClaudeRuntimeModelHelpUsesCliAliases(t *testing.T) {
	body := []byte(`  --model <model>                       Model for the current session.
                                        Provide an alias for the latest model
                                        (e.g. 'fable', 'opus', or 'sonnet')
                                        or a model's full name
                                        (e.g. 'claude-fable-5').
  -n, --name <name>                     Set a display name`)
	models := parseClaudeRuntimeModelHelp(body)
	if len(models) != 4 {
		t.Fatalf("models = %+v", models)
	}
	if countRuntimeModelDefaults(models) != 1 {
		t.Fatalf("default model count invalid: %+v", models)
	}
}
