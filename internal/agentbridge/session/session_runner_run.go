package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func run(ctx context.Context, cfg Config, sess *Session, proc process.RunningProcess) {
	defer close(sess.done)
	defer close(sess.events)
	defer close(sess.result)

	runner := newSessionRunner(ctx, cfg, sess, proc)
	defer runner.stopTimers()
	runner.startProtocol()
	defer runner.closeProtocol()

	runner.loop()
	runner.finish()
}
