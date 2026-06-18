package bridge

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type stubAdapter struct {
	name         string
	detected     agentbridge.DetectResult
	startCommand agentbridge.StartCommand
	seenStart    agentbridge.StartRequest
}

func (a *stubAdapter) Name() string { return a.name }

func (a *stubAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return a.detected, nil
}

func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	a.seenStart = req
	if a.startCommand.Executable != "" {
		return a.startCommand, nil
	}
	cmd := a.startCommand
	cmd.Executable = req.Executable
	if cmd.Executable == "" {
		cmd.Executable = a.name
	}
	return cmd, nil
}

func (a *stubAdapter) NewParser() agentbridge.Parser { return &stubParser{} }

func (a *stubAdapter) Translate(
	raw agentbridge.RawEvent,
) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "chunk" {
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: string(raw.Bytes),
			},
		}}, nil, nil
	}
	return nil, nil, nil
}

func (a *stubAdapter) BlockedArgs() []string { return nil }
