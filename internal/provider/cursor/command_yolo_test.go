package cursor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestBuildStartYoloIsExplicitOptIn(t *testing.T) {
	cmd, _ := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{})
	if strings.Contains(strings.Join(cmd.Args, " "), "--yolo") {
		t.Fatalf("--yolo must NOT be set by default: %v", cmd.Args)
	}
	if _, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{AllowYolo: true}); err == nil {
		t.Fatal("AllowYolo without policy allow must be rejected")
	}
	if _, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{
		AllowYolo:           true,
		TrustTier:           policy.TrustTierHost,
		UnsafeBypassAllowed: true,
	}); err == nil {
		t.Fatal("AllowYolo on Host trust tier must be rejected")
	}
	cmd, err := BuildStart(agentbridge.StartRequest{Prompt: "x"}, StartOptions{
		AllowYolo:           true,
		TrustTier:           policy.TrustTierEphemeralVM,
		UnsafeBypassAllowed: true,
	})
	if err != nil {
		t.Fatalf("isolated policy-approved AllowYolo should pass: %v", err)
	}
	if !strings.Contains(strings.Join(cmd.Args, " "), "--yolo") {
		t.Fatalf("--yolo must be set when AllowYolo: %v", cmd.Args)
	}
}
