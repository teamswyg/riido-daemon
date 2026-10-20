package agentbridge

import (
	"strings"
	"testing"
)

func TestHostContextContractBoundsDiscoveryAndHistoryOnEveryProvider(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "cursor", "openclaw"} {
		t.Run(provider, func(t *testing.T) {
			prompt, system, _, _ := ApplyRuntimeInstructionContract(provider, "Summarize my recent work.", "", "Team instruction")
			combined := prompt + "\n" + system
			for _, rule := range []string{
				"history/thread/session", "missing Git repository does not establish", "already-authorized native CLI",
				"authorized for the current requester", "Do not read credential files", "at most three relevant threads",
				"includeOutputs=false", "turnLimit=3", "maxOutputCharsPerItem=1500", "same authorized principal",
				"Do not reread the latest page", "Use the returned cursor", "after new messages or mutations",
			} {
				if !strings.Contains(combined, rule) {
					t.Errorf("missing host context rule %q", rule)
				}
			}
			if strings.Count(combined, "Riido host context discovery contract:") != 1 {
				t.Fatal("host context contract lost or duplicated")
			}
			if (provider == "codex" || provider == "claude") && prompt != "Summarize my recent work." {
				t.Fatal("native providers must preserve user prompt")
			}
		})
	}
}
