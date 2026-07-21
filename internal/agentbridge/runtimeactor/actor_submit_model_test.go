package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestSubmitModelUsesSupportedCodexDefault(t *testing.T) {
	models := []RuntimeModel{
		{ModelID: "gpt-5.5"},
		{ModelID: "gpt-5.6-sol", IsDefault: true},
	}
	if got := submitModel(bridge.Provider("codex"), "", "", models); got != "gpt-5.6-sol" {
		t.Fatalf("submit model = %q", got)
	}
}

func TestSubmitModelPreservesExplicitCodexSelection(t *testing.T) {
	models := []RuntimeModel{{ModelID: "gpt-5.6-sol", IsDefault: true}}
	if got := submitModel(bridge.Provider("codex"), "gpt-5.5", "", models); got != "gpt-5.5" {
		t.Fatalf("submit model = %q", got)
	}
}

func TestSubmitModelDoesNotRewriteResumedCodexSession(t *testing.T) {
	models := []RuntimeModel{{ModelID: "gpt-5.6-sol", IsDefault: true}}
	if got := submitModel(bridge.Provider("codex"), "", "thread-1", models); got != "" {
		t.Fatalf("resume model = %q", got)
	}
}

func TestSubmitModelDoesNotChangeOtherProviders(t *testing.T) {
	models := []RuntimeModel{{ModelID: "sonnet", IsDefault: true}}
	if got := submitModel(bridge.Provider("claude"), "", "", models); got != "" {
		t.Fatalf("claude model = %q", got)
	}
}
