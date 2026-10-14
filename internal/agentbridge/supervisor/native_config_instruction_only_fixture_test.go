package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type nativeConfigInstructionOnlyRun struct {
	provider bridge.Provider
	result   agentbridge.Result
	command  process.Command
}

func nativeConfigInstructionOnlyProviders() []bridge.Provider {
	return []bridge.Provider{"openclaw", "cursor"}
}

func runNativeConfigInstructionOnlyTask(
	t *testing.T,
	provider bridge.Provider,
) nativeConfigInstructionOnlyRun {
	t.Helper()

	reporter, running := startNativeConfigInstructionOnlyActor(t, provider)
	waitForNativeConfigTaskClaim(t, reporter)
	waitForNativeConfigProcessSpawn(t, running)

	go completeNativeConfigProcess(running)
	result := waitForNativeConfigResult(t, reporter)

	return nativeConfigInstructionOnlyRun{
		provider: provider,
		result:   result,
		command:  running.Command(),
	}
}
