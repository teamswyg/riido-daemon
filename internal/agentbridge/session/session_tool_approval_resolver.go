package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type toolApprovalResolverResult struct {
	resolution agentbridge.ToolApprovalResolution
	err        error
}

func (r *sessionRunner) resolvePendingApprovalIfAvailable(tool agentbridge.ToolRef) bool {
	if r.cfg.ToolApprovalResolver == nil {
		return false
	}
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	if r.idleTimer != nil {
		r.idleTimer.Stop()
	}
	resolved := make(chan toolApprovalResolverResult, 1)
	go func() {
		resolution, err := r.cfg.ToolApprovalResolver.ResolveToolApproval(ctx, r.cfg.TaskID, tool)
		resolved <- toolApprovalResolverResult{resolution: resolution, err: err}
	}()
	select {
	case result := <-resolved:
		r.resetIdle()
		if result.err != nil {
			r.emit(agentbridge.Event{Kind: agentbridge.EventWarning, Text: "tool approval resolver failed", Err: result.err.Error()})
			return false
		}
		r.executeResolvedToolApproval(tool, result.resolution)
		return true
	case <-r.ctx.Done():
		r.emitAndTerminate(contextTerminationEvent(r.ctx))
		return true
	case req := <-r.sess.cancel:
		r.cancel(req)
		return true
	case <-r.hardC:
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		return true
	}
}

func (r *sessionRunner) executeResolvedToolApproval(tool agentbridge.ToolRef, resolution agentbridge.ToolApprovalResolution) {
	if r.state.Terminal {
		return
	}
	select {
	case <-r.ctx.Done():
		r.emitAndTerminate(contextTerminationEvent(r.ctx))
		return
	case req := <-r.sess.cancel:
		r.cancel(req)
		return
	case <-r.hardC:
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		return
	default:
	}
	cmd := agentbridge.Command{
		Kind:              agentbridge.CommandRejectTool,
		ToolID:            tool.ID,
		ToolKind:          tool.Kind,
		ProviderRequestID: tool.ProviderRequestID,
		Reason:            resolution.Reason,
	}
	if resolution.Approved {
		cmd.Kind = agentbridge.CommandApproveTool
	}
	for _, cmdEvent := range r.executeCommands(r.ctx, []agentbridge.Command{cmd}) {
		r.emit(cmdEvent)
	}
	if !resolution.Approved {
		r.blockToolUse("tool approval rejected by resolver", toolApprovalRejectionReason(resolution))
	}
}

func toolApprovalRejectionReason(resolution agentbridge.ToolApprovalResolution) string {
	if resolution.Reason != "" {
		return resolution.Reason
	}
	return "tool approval denied"
}
