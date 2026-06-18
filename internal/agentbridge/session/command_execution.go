package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func executeCommands(
	ctx context.Context,
	proc process.RunningProcess,
	adapter agentbridge.Adapter,
	cmds []agentbridge.Command,
	killTimeout time.Duration,
) []agentbridge.Event {
	executor := commandExecutor{
		proc:        proc,
		adapter:     adapter,
		killTimeout: killTimeout,
	}

	var out []agentbridge.Event
	for _, cmd := range cmds {
		out = append(out, executor.execute(ctx, cmd)...)
	}
	return out
}

type commandExecutor struct {
	proc        process.RunningProcess
	adapter     agentbridge.Adapter
	killTimeout time.Duration
}

func (e commandExecutor) execute(ctx context.Context, cmd agentbridge.Command) []agentbridge.Event {
	switch cmd.Kind {
	case agentbridge.CommandCancelProvider:
		return e.cancelProvider(ctx)
	case agentbridge.CommandWriteProviderInput:
		return e.writeProviderInput(cmd.Input)
	case agentbridge.CommandApproveTool, agentbridge.CommandRejectTool:
		return e.writeApprovalCommand(cmd)
	case agentbridge.CommandFlushEvents,
		agentbridge.CommandPersistSession,
		agentbridge.CommandStartProvider:
		return nil
	default:
		return nil
	}
}

func (e commandExecutor) cancelProvider(ctx context.Context) []agentbridge.Event {
	if err := killProcess(ctx, e.proc, e.killTimeout); err != nil {
		return []agentbridge.Event{{
			Kind: agentbridge.EventWarning,
			Text: "provider kill failed",
			Err:  err.Error(),
		}}
	}
	return nil
}
