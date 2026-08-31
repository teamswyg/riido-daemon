package claude

import (
	"context"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

type claudeAuthStatus struct {
	LoggedIn bool `json:"loggedIn"`
}

func claudeAuthProbe(ctx context.Context, executable string) detectutil.AuthProbeStatus {
	probe := detectutil.VersionProbeStrict(ctx, executable, "auth", "status")
	if !probe.OK {
		return detectutil.AuthProbeUnknown
	}
	var status claudeAuthStatus
	if json.Unmarshal([]byte(probe.Output), &status) != nil {
		return detectutil.AuthProbeUnknown
	}
	if status.LoggedIn && probe.ExitCode == 0 {
		return detectutil.AuthProbeAuthenticated
	}
	return detectutil.AuthProbeUnauthenticated
}
