package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func cleanupTempFiles(paths []string) []agentbridge.Event {
	var out []agentbridge.Event
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			out = append(out, agentbridge.Event{
				Kind: agentbridge.EventWarning,
				Text: "adapter temp file cleanup failed",
				Err:  fmt.Sprintf("%s: %v", path, err),
			})
		}
	}
	return out
}

func executeCommands(
	proc process.RunningProcess,
	adapter agentbridge.Adapter,
	cmds []agentbridge.Command,
	killTimeout time.Duration,
) []agentbridge.Event {
	var out []agentbridge.Event
	for _, c := range cmds {
		switch c.Kind {
		case agentbridge.CommandCancelProvider:
			if err := killProcess(proc, killTimeout); err != nil {
				out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider kill failed", Err: err.Error()})
			}
		case agentbridge.CommandWriteProviderInput:
			if len(c.Input) > 0 {
				if err := proc.WriteStdin(c.Input); err != nil {
					out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider input write failed", Err: err.Error()})
				}
			}
		case agentbridge.CommandApproveTool, agentbridge.CommandRejectTool:
			builder, ok := adapter.(agentbridge.ProviderInputBuilder)
			if !ok {
				out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command has no input builder"})
				continue
			}
			input, err := builder.BuildProviderInput(c)
			if err != nil {
				out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command build failed", Err: err.Error()})
				continue
			}
			if len(input) > 0 {
				if err := proc.WriteStdin(input); err != nil {
					out = append(out, agentbridge.Event{Kind: agentbridge.EventWarning, Text: "provider approval command write failed", Err: err.Error()})
				}
			}
		case agentbridge.CommandFlushEvents,
			agentbridge.CommandPersistSession,
			agentbridge.CommandStartProvider:
			// Other commands are no-ops at this layer; the supervisor /
			// runtime actor (to be added) will route them. For the session
			// actor we just need to ensure CancelProvider terminates the
			// child process, which Reduce already emits on Cancel/Timeout.
		}
	}
	return out
}

func killProcess(proc process.RunningProcess, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultProcessKillTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return proc.Kill(ctx)
}

func decideStartedTool(gate agentbridge.ToolStartGate, tool agentbridge.ToolRef) agentbridge.ToolStartDecision {
	if gate == nil {
		return agentbridge.ToolStartDecision{}
	}
	return gate(tool)
}

func toolBlockReason(decision agentbridge.ToolStartDecision) string {
	if decision.Code == "" {
		return decision.Reason
	}
	if decision.Reason == "" {
		return decision.Code
	}
	return decision.Code + ": " + decision.Reason
}

func drain(ch <-chan []byte) {
	if ch == nil {
		return
	}
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
