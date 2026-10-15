package agentbridge

import (
	"strings"
	"testing"
)

func TestTaskContextContractUsesEveryNativeProviderInstructionSurface(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "cursor", "openclaw"} {
		t.Run(provider, func(t *testing.T) {
			prompt, system, _, _ := ApplyRuntimeInstructionContract(provider, "Implement the assigned task.", "", "Team instruction")
			combined := prompt + "\n" + system
			if strings.Count(combined, "Riido task context contract:") != 1 {
				t.Fatal("context contract lost or duplicated")
			}
			for _, rule := range []string{"latest user instruction", "reference evidence", "unavailable or truncated", "assigned task"} {
				if !strings.Contains(combined, rule) {
					t.Errorf("missing %q", rule)
				}
			}
			if provider == "claude" || provider == "codex" {
				if prompt != "Implement the assigned task." || !strings.Contains(system, "Riido task context contract:") {
					t.Fatal("native instruction surface misplaced")
				}
			}
			if provider == "cursor" && (!strings.Contains(prompt, "Implement the assigned task.") || strings.Index(prompt, "Implement the assigned task.") > strings.Index(prompt, "Riido task context contract:")) {
				t.Fatal("Cursor inline fallback lost task-first context")
			}
		})
	}
}
