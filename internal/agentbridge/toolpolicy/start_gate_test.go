package toolpolicy

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestPolicyToolStartGateBlocksClassifiedRiskWithoutApprovalPath(t *testing.T) {
	gate := PolicyToolStartGate(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost)

	assertStartGateBlocks(t, gate, agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "terraform destroy"}})
	assertStartGateBlocks(t, gate, agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "cat .env.local"}})
}

func TestPolicyToolStartGateAllowsExplicitSurfaceAndUnclassifiedTools(t *testing.T) {
	gate := PolicyToolStartGate(testPolicyBundle(policy.AllowedSurfaceSet{
		ToolUse: []policy.ToolUseSurface{policy.ToolUseNetworkEgress},
	}), policy.TrustTierHost)

	if decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}); decision.Block {
		t.Fatalf("allowed network surface should not block: %+v", decision)
	}
	if decision := gate(agentbridge.ToolRef{Kind: "read", Name: "Read"}); decision.Block {
		t.Fatalf("unclassified read tool should not block: %+v", decision)
	}
}

func assertStartGateBlocks(t *testing.T, gate agentbridge.ToolStartGate, tool agentbridge.ToolRef) {
	t.Helper()
	decision := gate(tool)
	if !decision.Block {
		t.Fatalf("started tool must block: %+v", decision)
	}
	if decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision code = %q", decision.Code)
	}
}
