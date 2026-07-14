package claude

import (
	"context"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

type claudeAuthStatus struct {
	LoggedIn bool `json:"loggedIn"`
}

func claudeAuthenticated(ctx context.Context, executable string) bool {
	probe := detectutil.VersionProbeStrict(ctx, executable, "auth", "status")
	if !probe.OK || probe.ExitCode != 0 {
		return false
	}
	var status claudeAuthStatus
	return json.Unmarshal([]byte(probe.Output), &status) == nil && status.LoggedIn
}
