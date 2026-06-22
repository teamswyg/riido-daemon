package openclaw

import (
	"maps"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func buildStartEnv(req agentbridge.StartRequest) ([]string, []string, error) {
	env := make(map[string]string, len(req.Env)+1)
	maps.Copy(env, req.Env)

	tempFiles, err := maybeWriteTaskScopedConfig(req, env)
	if err != nil {
		return nil, nil, err
	}
	return envList(env), tempFiles, nil
}
