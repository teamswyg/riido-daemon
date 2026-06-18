package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
