package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestDaemonToolStartGateUsesActivePolicyBundle(t *testing.T) {
	settings := daemonPolicySettings("policy-bundle.tool-start.v1", policy.ToolUseNetworkEgress)
	gate := daemonToolStartGate(settings)

	if decision := gate(shellTool("curl https://example.com")); decision.Block {
		t.Fatalf("allowed network surface should not block: %+v", decision)
	}
	if decision := gate(shellTool("terraform destroy")); !decision.Block {
		t.Fatalf("unallowed destructive command should block: %+v", decision)
	}
}

func TestDaemonToolApprovalGateUsesActivePolicyBundle(t *testing.T) {
	settings := daemonPolicySettings("policy-bundle.tool-approval.v1", policy.ToolUseNetworkEgress)
	gate := daemonToolApprovalGate(settings)

	if decision := gate(shellTool("curl https://example.com")); decision.Block {
		t.Fatalf("allowed network approval should not block: %+v", decision)
	}
	if decision := gate(shellTool("cat .env.local")); !decision.Block {
		t.Fatalf("unallowed secret exposure approval should block: %+v", decision)
	} else if decision.Code != "approval_timeout" {
		t.Fatalf("unallowed secret exposure approval code = %q", decision.Code)
	}
}

func shellTool(command string) agentbridge.ToolRef {
	return agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": command}}
}
