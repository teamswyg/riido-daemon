package claude

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func buildStartEnv(req agentbridge.StartRequest) []string {
	env := make([]string, 0, len(req.Env))
	for k, v := range req.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}
