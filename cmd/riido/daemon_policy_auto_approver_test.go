package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestDaemonToolAutoApproverUsesActivePolicyBundle(t *testing.T) {
	settings := daemonPolicySettings(
		"policy-bundle.tool-auto.v1",
		policy.ToolUseDestructiveCommand,
	)
	approver := daemonToolAutoApprover(settings)

	if !approver(agentbridge.ToolRef{Kind: "shell"}) {
		t.Fatal("daemon policy auto approver should approve explicitly allowed shell surface")
	}
	if approver(agentbridge.ToolRef{Kind: "patch_apply"}) {
		t.Fatal("daemon policy auto approver must not approve unallowed patch surface")
	}
}
