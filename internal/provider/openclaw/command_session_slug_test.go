package openclaw

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartDerivesProviderSafeSessionIDFromRiidoComponentID(t *testing.T) {
	cmd, err := BuildStart(agentbridge.StartRequest{
		TaskID: "-4ckNAErFPZoB721KhZgt",
		Prompt: "do the thing",
	}, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sessionID := findCommandArgValue(cmd.Args, "--session-id")
	if sessionID == "" {
		t.Fatalf("session id not found: %v", cmd.Args)
	}
	if strings.HasPrefix(sessionID, "-") {
		t.Fatalf("session id must not start with hyphen: %q", sessionID)
	}
	if !strings.HasPrefix(sessionID, "riido-4ckNAErFPZoB721KhZgt-") {
		t.Fatalf("session id did not preserve task id slug: %q", sessionID)
	}
	if len(sessionID) > 80 || !isOpenClawSessionID(sessionID) {
		t.Fatalf("session id is not provider-safe: %q", sessionID)
	}
	if got := sessionIDFromTaskID("-4ckNAErFPZoB721KhZgt"); got != sessionID {
		t.Fatalf("session id must be deterministic: %q != %q", got, sessionID)
	}
}

func findCommandArgValue(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
