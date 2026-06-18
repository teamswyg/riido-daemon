package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) consumeStdout(chunk []byte, ok bool) {
	if !ok {
		r.stdoutCh = nil
		return
	}
	raws, err := r.parser.FeedStdout(chunk)
	if err != nil {
		r.emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
		return
	}
	r.feed(raws)
}

func (r *sessionRunner) consumeStderr(chunk []byte, ok bool) {
	if !ok {
		r.stderrCh = nil
		return
	}
	raws, err := r.parser.FeedStderr(chunk)
	if err != nil {
		return
	}
	r.feed(raws)
}
