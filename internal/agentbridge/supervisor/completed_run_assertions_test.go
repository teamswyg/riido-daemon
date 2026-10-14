package supervisor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func assertSupervisorCompletedRun(
	t *testing.T,
	res agentbridge.Result,
	running *process.FakeRunning,
) {
	t.Helper()

	if res.Status != agentbridge.ResultCompleted || res.Output != "done" {
		t.Fatalf("result: %+v", res)
	}
	if res.Workdir == "" {
		t.Fatalf("expected isolated workdir in result: %+v", res)
	}
	if running.Command().Dir != res.Workdir {
		t.Fatalf("spawn dir %q != result workdir %q", running.Command().Dir, res.Workdir)
	}
	if !hasEnvPrefix(running.Command().Env, "TEST_NATIVE_CONFIG_VERSION=") {
		t.Fatalf("native config version was not passed to adapter metadata: %+v", running.Command())
	}

	assertCompletedNativeConfigInjected(t, res.Workdir)
	assertCompletedArchiveManifest(t, res.Workdir)
	assertCompletedRunEvents(t, res.Workdir)
}
