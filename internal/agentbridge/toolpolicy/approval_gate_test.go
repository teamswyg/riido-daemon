package toolpolicy

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestPolicyToolApprovalGateBlocksClassifiedRiskWithoutApprovalPath(t *testing.T) {
	gate := PolicyToolApprovalGate(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost)

	decision := gate(agentbridge.ToolRef{Kind: "patch_apply"})
	if !decision.Block {
		t.Fatalf("headless patch approval must block: %+v", decision)
	}
	if decision.Code != "approval_timeout" {
		t.Fatalf("decision code = %q", decision.Code)
	}
	if !strings.Contains(decision.Reason, "no approval path") {
		t.Fatalf("decision reason = %q", decision.Reason)
	}
}

func TestPolicyToolApprovalGateAllowsExplicitSurfaceAndUnclassifiedTools(t *testing.T) {
	gate := PolicyToolApprovalGate(testPolicyBundle(policy.AllowedSurfaceSet{
		ToolUse: []policy.ToolUseSurface{policy.ToolUseProtectedPathWrite},
	}), policy.TrustTierHost)

	if decision := gate(agentbridge.ToolRef{Kind: "patch_apply"}); decision.Block {
		t.Fatalf("allowed patch surface should not block: %+v", decision)
	}
	if decision := gate(agentbridge.ToolRef{Kind: "read", Name: "Read"}); decision.Block {
		t.Fatalf("unclassified read tool should not block: %+v", decision)
	}
}
