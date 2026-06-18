package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (r *sessionRunner) blockStartedToolIfNeeded(ev agentbridge.Event) bool {
	if ev.Kind != agentbridge.EventToolCallStarted {
		return false
	}
	decision := decideStartedTool(r.cfg.ToolStartGate, ev.Tool)
	if !decision.Block {
		return false
	}
	r.blockToolUse("tool use blocked by policy", toolBlockReason(decision))
	return true
}

func (r *sessionRunner) blockPendingApprovalIfNeeded(ev agentbridge.Event, cmds []agentbridge.Command) bool {
	if ev.Kind != agentbridge.EventToolApprovalNeeded || hasApproveToolCommand(cmds) {
		return false
	}
	if r.resolvePendingApprovalIfAvailable(ev.Tool) {
		return r.state.Terminal
	}
	decision := decideApprovalTool(r.cfg.ToolApprovalGate, ev.Tool)
	if !decision.Block {
		return false
	}
	r.blockToolUse("tool approval unavailable in headless run", toolBlockReason(decision))
	return true
}

func (r *sessionRunner) blockToolUse(warning, reason string) {
	cancel := []agentbridge.Command{{Kind: agentbridge.CommandCancelProvider, Reason: reason}}
	for _, cmdEvent := range executeCommands(r.ctx, r.proc, r.cfg.Adapter, cancel, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
	r.emit(agentbridge.Event{Kind: agentbridge.EventWarning, Text: warning, Err: reason})
	blocked := agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  reason,
		},
	}
	r.emit(blocked)
	var cmds []agentbridge.Command
	r.state, cmds = agentbridge.Reduce(r.state, blocked, r.cfg.AutoApprove)
	for _, cmdEvent := range executeCommands(r.ctx, r.proc, r.cfg.Adapter, cmds, r.cfg.ProcessKillTimeout) {
		r.emit(cmdEvent)
	}
}
