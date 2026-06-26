package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestBuildStartRequiresExplicitPermissionMode(t *testing.T) {
	_, err := BuildStart(agentbridge.StartRequest{}, StartOptions{})
	if err == nil {
		t.Fatal("expected error when PermissionMode is empty (no implicit bypass)")
	}
}

func TestBuildStartBypassRequiresPolicyAllow(t *testing.T) {
	if _, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode: PermissionModeBypassDangerous,
	}); err == nil {
		t.Fatal("bypassPermissions without trust tier and policy allow must be rejected")
	}
	if _, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode:      PermissionModeBypassDangerous,
		TrustTier:           policy.TrustTierHost,
		UnsafeBypassAllowed: true,
	}); err == nil {
		t.Fatal("bypassPermissions on Host trust tier must be rejected")
	}
	cmd, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode:      PermissionModeBypassDangerous,
		TrustTier:           policy.TrustTierIsolatedContainer,
		UnsafeBypassAllowed: true,
	})
	if err != nil {
		t.Fatalf("isolated policy-approved bypassPermissions should pass: %v", err)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--permission-mode bypassPermissions") {
		t.Fatalf("bypassPermissions mode missing from args: %v", cmd.Args)
	}
}

func TestBuildStartAllowsExplicitBetaFullAccessHarness(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode:        PermissionModeBypassDangerous,
		BetaFullAccessAllowed: true,
	})
	if err != nil {
		t.Fatalf("beta full-access harness should pass: %v", err)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--permission-mode bypassPermissions") {
		t.Fatalf("bypassPermissions mode missing from args: %v", cmd.Args)
	}
}
