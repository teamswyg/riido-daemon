package bridge

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func newSpawnProcess(
	cmd agentbridge.StartCommand,
	defaultDir string,
	launchEnv map[string]string,
) process.Command {
	spawnProcess := toProcessCommand(cmd)
	if spawnProcess.Dir == "" {
		spawnProcess.Dir = defaultDir
	}
	spawnProcess.Env = detectutil.EnvListWithLaunchPATHFromMap(spawnProcess.Env, launchEnv)
	return spawnProcess
}

func toProcessCommand(cmd agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: cmd.Executable,
		Args:       cmd.Args,
		Env:        cmd.Env,
		Dir:        cmd.Dir,
	}
}
