package main

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type daemonTestAdapter struct {
	name  string
	delay time.Duration
}

func (a daemonTestAdapter) Name() string { return a.name }

func (a daemonTestAdapter) Detect(ctx context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	if a.delay > 0 {
		select {
		case <-time.After(a.delay):
		case <-ctx.Done():
			return agentbridge.DetectResult{}, ctx.Err()
		}
	}
	return agentbridge.DetectResult{
		Available:         true,
		Executable:        a.name,
		Version:           "test",
		SupportsStreaming: true,
	}, nil
}

func (a daemonTestAdapter) BuildStart(agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{Executable: a.name}, nil
}

func (a daemonTestAdapter) NewParser() agentbridge.Parser { return daemonTestParser{} }

func (a daemonTestAdapter) Translate(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return nil, nil, nil
}

func (a daemonTestAdapter) BlockedArgs() []string { return nil }

type daemonTestParser struct{}

func (daemonTestParser) FeedStdout([]byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (daemonTestParser) FeedStderr([]byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (daemonTestParser) Close() ([]agentbridge.RawEvent, error)            { return nil, nil }
