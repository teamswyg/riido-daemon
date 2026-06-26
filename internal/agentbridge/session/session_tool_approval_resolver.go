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
	resolved := make(chan toolApprovalResolverResult, 1)
	go func() {
		resolution, err := r.cfg.ToolApprovalResolver.ResolveToolApproval(ctx, r.cfg.TaskID, tool)
		resolved <- toolApprovalResolverResult{resolution: resolution, err: err}
	}()
	select {
	case result := <-resolved:
		if result.err != nil {
			r.emit(agentbridge.Event{Kind: agentbridge.EventWarning, Text: "tool approval resolver failed", Err: result.err.Error()})
			return false
		}
		r.executeResolvedToolApproval(tool, result.resolution)
		return true
	case <-r.ctx.Done():
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventCancellation, Err: r.ctx.Err().Error()})
		return true
	case req := <-r.sess.cancel:
		r.cancel(req)
		return true
	case <-r.hardC:
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "hard timeout"})
		return true
	case <-r.idleC:
		r.emitAndTerminate(agentbridge.Event{Kind: agentbridge.EventTimeout, Err: "semantic idle timeout"})
		return true
	}
}

func (r *sessionRunner) executeResolvedToolApproval(tool agentbridge.ToolRef, resolution agentbridge.ToolApprovalResolution) {
	cmd := agentbridge.Command{
		Kind:              agentbridge.CommandRejectTool,
		ToolID:            tool.ID,
		ProviderRequestID: tool.ProviderRequestID,
		Reason:            resolution.Reason,
	}
	if resolution.Approved {
		cmd.Kind = agentbridge.CommandApproveTool
	}
	for _, cmdEvent := range executeCommands(r.ctx, r.proc, r.cfg.Adapter, []agentbridge.Command{cmd}, r.cfg.ProcessKillTimeout) {
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
