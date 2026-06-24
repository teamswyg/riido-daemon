package toolpolicy

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestPolicyAutoApproverApprovesAllowedAndUnclassifiedTools(t *testing.T) {
	bundle := testPolicyBundle(policy.AllowedSurfaceSet{
		ToolUse: []policy.ToolUseSurface{policy.ToolUseDestructiveCommand},
	})
	approver := PolicyAutoApprover(bundle, policy.TrustTierHost)

	if !approver(agentbridge.ToolRef{Kind: "shell", ProviderRequestID: "req-1"}) {
		t.Fatal("shell should auto-approve when destructive-command surface is allowed")
	}
	if approver(agentbridge.ToolRef{Kind: "patch_apply"}) {
		t.Fatal("patch_apply must stay on human approval path without protected-path-write allow")
	}
	if !approver(agentbridge.ToolRef{Kind: "read"}) {
		t.Fatal("unclassified tool should auto-approve on a known trust tier")
	}
	if !approver(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "go run main.go"}}) {
		t.Fatal("unclassified safe shell command should auto-approve on a known trust tier")
	}
}

func TestPolicyAutoApproverDoesNotApproveUnknownTier(t *testing.T) {
	bundle := testPolicyBundle(policy.AllowedSurfaceSet{
		ToolUse: []policy.ToolUseSurface{policy.ToolUseDestructiveCommand},
	})
	approver := PolicyAutoApprover(bundle, policy.TrustTierUnknown)

	if approver(agentbridge.ToolRef{Kind: "shell"}) {
		t.Fatal("Unknown trust tier must not auto-approve tool use")
	}
	if approver(agentbridge.ToolRef{Kind: "read"}) {
		t.Fatal("Unknown trust tier must not auto-approve unclassified tool use")
	}
}
