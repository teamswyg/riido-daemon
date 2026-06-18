package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func submitSpawnCommand(
	spawn agentbridge.StartCommand,
	startReq agentbridge.StartRequest,
	launchEnv map[string]string,
) process.Command {
	spawnCommand := toProcessCommand(spawn)
	if spawnCommand.Dir == "" {
		spawnCommand.Dir = startReq.Cwd
	}
	spawnCommand.Env = detectutil.EnvListWithLaunchPATHFromMap(spawnCommand.Env, launchEnv)
	return spawnCommand
}
