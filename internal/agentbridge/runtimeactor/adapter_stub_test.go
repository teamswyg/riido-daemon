package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type stubAdapter struct {
	name       string
	detected   agentbridge.DetectResult
	startReqCh chan agentbridge.StartRequest
}

func (a *stubAdapter) Name() string { return a.name }

func (a *stubAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return a.detected, nil
}

func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	if a.startReqCh != nil {
		select {
		case a.startReqCh <- req:
		default:
		}
	}
	exe := req.Executable
	if exe == "" {
		exe = a.name
	}
	return agentbridge.StartCommand{Executable: exe}, nil
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
