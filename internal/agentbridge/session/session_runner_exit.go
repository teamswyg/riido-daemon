package session

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func (r *sessionRunner) startProtocol() {
	if r.cfg.ProtocolDriver == nil {
		return
	}
	if err := r.cfg.ProtocolDriver.OnStart(r.ctx, r.io); err != nil {
		r.applyEvents([]agentbridge.Event{
			{Kind: agentbridge.EventError, Err: err.Error()},
			{Kind: agentbridge.EventResult, Result: agentbridge.Result{Status: agentbridge.ResultFailed, Error: err.Error()}},
		}, nil)
	}
}

func (r *sessionRunner) closeProtocol() {
	if r.cfg.ProtocolDriver != nil {
		_ = r.cfg.ProtocolDriver.OnClose(r.ctx, r.io)
	}
}

func (r *sessionRunner) cancel(cause error) {
	reason := "cancelled"
	if cause != nil {
		reason = cause.Error()
	}
	r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventCancellation, Err: reason})
}

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
	if raws, err := r.parser.Close(); err == nil {
		r.feed(raws)
	}
	if r.state.Terminal {
		return true
	}
	r.applyDriverProcessExit()
	if r.state.Terminal {
		return true
	}
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
	return true
}

func (r *sessionRunner) applyDriverProcessExit() {
	if r.cfg.ProtocolDriver == nil || r.deferredExit == nil {
		return
	}
	exitStatus := agentbridge.ProcessExitStatus{Code: r.deferredExit.Code}
	if r.deferredExit.Err != nil {
		exitStatus.Err = r.deferredExit.Err.Error()
	}
	driverEvents, err := r.cfg.ProtocolDriver.OnProcessExit(r.ctx, exitStatus, r.io)
	if err != nil {
		r.emit(agentbridge.Event{Kind: agentbridge.EventError, Err: err.Error()})
	}
	if len(driverEvents) > 0 {
		r.applyEvents(driverEvents, nil)
	}
}
