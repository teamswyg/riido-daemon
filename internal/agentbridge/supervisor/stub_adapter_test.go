package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type stubAdapter struct {
	name string
}

func (a *stubAdapter) Name() string { return a.name }

func (a *stubAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true, Version: "1.0", Executable: a.name}, nil
}

func (a *stubAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	cmd := agentbridge.StartCommand{Executable: a.name}
	if version := req.Metadata[MetadataNativeConfigVersion]; version != "" {
		cmd.Env = append(cmd.Env, "TEST_NATIVE_CONFIG_VERSION="+version)
	}
	if nativeConfigHome := req.Metadata[MetadataNativeConfigHome]; nativeConfigHome != "" {
		cmd.Env = append(cmd.Env, "TEST_NATIVE_CONFIG_HOME="+nativeConfigHome)
	}
	return cmd, nil
}

func (a *stubAdapter) NewParser() agentbridge.Parser { return &stubParser{} }

func (a *stubAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "event" {
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)}}, nil, nil
	}
	if raw.Type == "chunk" {
		return []agentbridge.Event{
			{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			{Kind: agentbridge.EventResult, Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: string(raw.Bytes),
			}},
		}, nil, nil
	}
	return nil, nil, nil
}

func (a *stubAdapter) BlockedArgs() []string { return nil }
