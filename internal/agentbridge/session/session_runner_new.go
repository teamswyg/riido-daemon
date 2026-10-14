package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func newSessionRunner(ctx context.Context, cfg Config, sess *Session, proc process.RunningProcess) *sessionRunner {
	state := agentbridge.NewState()
	state.LastSemanticActivity = cfg.Now()
	runner := &sessionRunner{
		ctx:       ctx,
		cfg:       cfg,
		sess:      sess,
		proc:      proc,
		parser:    cfg.Adapter.NewParser(),
		state:     state,
		telemetry: agentbridge.NewTelemetryParser(),
		io:        newProtocolIO(proc),
		startedAt: cfg.Now(),
		stdoutCh:  proc.Stdout(),
		stderrCh:  proc.Stderr(),
		exitedCh:  proc.Exited(),
	}
	runner.startTimers()
	return runner
}
