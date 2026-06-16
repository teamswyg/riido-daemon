package toolpolicy

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestClassifyToolUseSurfaceMapsProviderNeutralLabels(t *testing.T) {
	for _, tc := range []struct {
		name string
		tool agentbridge.ToolRef
		want policy.ToolUseSurface
	}{
		{"codex shell approval", agentbridge.ToolRef{Kind: "shell"}, policy.ToolUseDestructiveCommand},
		{"claude bash approval", agentbridge.ToolRef{Name: "Bash", Kind: "Bash"}, policy.ToolUseDestructiveCommand},
		{"codex patch apply", agentbridge.ToolRef{Kind: "patch_apply"}, policy.ToolUseProtectedPathWrite},
		{"protected path write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"path": ".git/config"}}, policy.ToolUseProtectedPathWrite},
		{"network fetch", agentbridge.ToolRef{Name: "WebFetch"}, policy.ToolUseNetworkEgress},
		{"network shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}, policy.ToolUseNetworkEgress},
		{"secret token", agentbridge.ToolRef{Name: "Token"}, policy.ToolUseSecretExposure},
		{"secret arg key", agentbridge.ToolRef{Name: "Read", Args: map[string]string{"api_token": "[redacted]"}}, policy.ToolUseSecretExposure},
		{"secret redacted arg value", agentbridge.ToolRef{Name: "Read", Args: map[string]string{"note": "[redacted]"}}, policy.ToolUseSecretExposure},
		{"secret env read shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "cat .env.local"}}, policy.ToolUseSecretExposure},
		{"secret manager shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "aws secretsmanager get-secret-value --secret-id prod/api"}}, policy.ToolUseSecretExposure},
		{"destructive shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "rm -rf .git"}}, policy.ToolUseDestructiveCommand},
		{"protected env write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"file_path": ".env.production"}}, policy.ToolUseProtectedPathWrite},
		{"protected ssh write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"path": "~/.ssh/config"}}, policy.ToolUseProtectedPathWrite},
		{"protected env shell write", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "printf TOKEN=x > .env"}}, policy.ToolUseProtectedPathWrite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ClassifyToolUseSurface(tc.tool)
			if !ok {
				t.Fatalf("tool should classify: %+v", tc.tool)
			}
			if got != tc.want {
				t.Fatalf("surface = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClassifyToolUseSurfaceUsesArgsToAvoidBroadShellClassification(t *testing.T) {
	if got, ok := ClassifyToolUseSurface(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "go test ./..."}}); ok {
		t.Fatalf("safe shell command must stay unclassified for human path: %q", got)
	}
}

func TestClassifyToolUseSurfaceLeavesUnknownToolsForHumanApproval(t *testing.T) {
	if got, ok := ClassifyToolUseSurface(agentbridge.ToolRef{Kind: "read", Name: "Read"}); ok {
		t.Fatalf("read-only tool must not auto-classify as a risk surface: %q", got)
	}
}

func TestPolicyAutoApproverOnlyApprovesExplicitAllowedSurface(t *testing.T) {
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
	if approver(agentbridge.ToolRef{Kind: "read"}) {
		t.Fatal("unclassified tool must stay on human approval path")
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
}

func TestPolicyToolStartGateBlocksClassifiedRiskWithoutApprovalPath(t *testing.T) {
	gate := PolicyToolStartGate(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost)

	decision := gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "terraform destroy"}})
	if !decision.Block {
		t.Fatalf("started destructive tool must block: %+v", decision)
	}
	if decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision code = %q", decision.Code)
	}
	decision = gate(agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "cat .env.local"}})
	if !decision.Block {
		t.Fatalf("started secret exposure tool must block: %+v", decision)
	}
	if decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("secret exposure decision code = %q", decision.Code)
	}
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

func TestPolicyToolApprovalGateBlocksClassifiedRiskWithoutApprovalPath(t *testing.T) {
	gate := PolicyToolApprovalGate(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost)

	decision := gate(agentbridge.ToolRef{Kind: "patch_apply"})
	if !decision.Block {
		t.Fatalf("headless patch approval must block: %+v", decision)
	}
	if decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision code = %q", decision.Code)
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

func TestDecisionForToolReturnsRequireApprovalWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForTool(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionRequireApproval || decision.Code != "TOOL_USE_REQUIRES_APPROVAL" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestDecisionForStartedToolInterruptsWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForStartedTool(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionInterruptAndBlock || decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestDecisionForHeadlessApprovalInterruptsWhenBundleDoesNotAllow(t *testing.T) {
	decision, ok := DecisionForHeadlessApproval(testPolicyBundle(policy.AllowedSurfaceSet{}), policy.TrustTierHost, agentbridge.ToolRef{Kind: "patch_apply"})
	if !ok {
		t.Fatal("patch_apply should classify")
	}
	if decision.Action != policy.ToolUseActionInterruptAndBlock || decision.Code != "TOOL_USE_NOT_IN_POLICY_BUNDLE" {
		t.Fatalf("decision = %+v", decision)
	}
}

func testPolicyBundle(surfaces policy.AllowedSurfaceSet) policy.PolicyBundle {
	return policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        "policy-bundle.toolpolicy-test.v1",
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {
				AllowedSurfaces: surfaces,
			},
		},
	}
}
