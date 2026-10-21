package agentbridge

import (
	"strings"
	"testing"
)

func TestTelemetryProgressCodeMeaningsReachSharedInstructionSurfaces(t *testing.T) {
	for name, instruction := range map[string]string{
		"injected": TelemetryContractInstruction(),
		"native":   strings.Join(TelemetryNativeConfigHardRules(), "\n"),
	} {
		t.Run(name, func(t *testing.T) {
			for _, meaning := range []string{
				"1001 thinking",
				"1101 tool collecting",
				"1102 collection completed count",
				"1103 tool running",
				"1104 tool completed",
			} {
				if !strings.Contains(instruction, meaning) {
					t.Errorf("missing progress code meaning %q", meaning)
				}
			}
		})
	}
}

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
