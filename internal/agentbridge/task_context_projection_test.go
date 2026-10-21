package agentbridge

import (
	"strings"
	"testing"
)

func TestContextProjectionSemanticsReachEveryProvider(t *testing.T) {
	for _, provider := range []string{"claude", "codex", "cursor", "openclaw"} {
		prompt, system, _, _ := ApplyRuntimeInstructionContract(provider, "SYNTHETIC_EXPLICIT_DIRECTIVE", "", "")
		for _, rule := range []string{"context lookup failed", "explicitly empty", "unconfirmed freshness", "unknown chronology", "latest Yjs"} {
			if !strings.Contains(prompt+"\n"+system, rule) {
				t.Errorf("%s missing %q", provider, rule)
			}
		}
		if (provider == "codex" || provider == "claude") && prompt != "SYNTHETIC_EXPLICIT_DIRECTIVE" {
			t.Fatal("native user directive modified")
		}
	}
}
