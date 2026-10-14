package session

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func (r *sessionRunner) deferExit(status process.ExitStatus, ok bool) {
	if !ok {
		r.exitedCh = nil
		return
	}
	st := status
	r.deferredExit = &st
	r.exitedCh = nil
}

func (r *sessionRunner) flushDeferredExitWhenDrained() bool {
	if r.stdoutCh != nil || r.stderrCh != nil || r.deferredExit == nil {
		return false
	}
	r.flushParserClose()
	if r.state.Terminal {
		return true
	}
	r.applyDriverProcessExit()
	if r.state.Terminal {
		return true
	}
	r.applyProcessExitFallback()
	return true
}

func (r *sessionRunner) flushParserClose() {
	if raws, err := r.parser.Close(); err == nil {
		r.feed(raws)
	}
}

func (r *sessionRunner) applyProcessExitFallback() {
	errMsg := ""
	if r.deferredExit.Err != nil {
		errMsg = r.deferredExit.Err.Error()
	}
	r.applyEvents([]agentbridge.Event{{Kind: agentbridge.EventProcessExit, ExitCode: r.deferredExit.Code, Err: errMsg}}, nil)
	if !r.state.Terminal {
		r.applyEvents([]agentbridge.Event{{
			Kind:   agentbridge.EventResult,
			Result: agentbridge.Result{Status: agentbridge.ResultAborted, Error: "process exited without provider result"},
		}}, nil)
	}
}
