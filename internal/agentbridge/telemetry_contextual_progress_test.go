package agentbridge

import (
	"strings"
	"testing"
)

func TestContextualProgressRulesReachEveryInstructionSurface(t *testing.T) {
	surfaces := map[string]string{
		"injected": InjectTelemetryContract("inspect reply permissions"),
		"native":   strings.Join(TelemetryNativeConfigHardRules(), "\n"),
	}
	for _, provider := range []string{"claude", "openclaw", "codex", "cursor"} {
		prompt, system, _, _ := ApplyRuntimeInstructionContract(provider, "inspect reply permissions", "", "")
		surfaces[provider] = prompt + "\n" + system
	}
	for name, instruction := range surfaces {
		t.Run(name, func(t *testing.T) {
			for _, required := range []string{
				"actual current or completed action for this task",
				"user's language",
				"one short sentence",
				"approximately 120 characters",
				"meaningful work changes",
				"Do not repeat",
				"Never invent tool results",
				"raw commands, code, internal function or tool names, debug logs, prompt dumps, secrets, or long answer drafts",
				"description", "summary", "state-neutral",
			} {
				if !strings.Contains(instruction, required) {
					t.Errorf("missing contextual progress rule %q", required)
				}
			}
			for _, forbidden := range []string{"periodically emit", "free-form user-facing copy"} {
				if strings.Contains(instruction, forbidden) {
					t.Errorf("conflicting generic progress rule %q remains", forbidden)
				}
			}
		})
	}
}
