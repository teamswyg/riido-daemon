package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type recordingAdapter struct {
	name        string
	startCmd    agentbridge.StartCommand
	blocked     []string
	translateFn func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error)
	inputFn     func(agentbridge.Command) ([]byte, error)
	parser      *recordingParser
}

func (a *recordingAdapter) Name() string { return a.name }

func (a *recordingAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a *recordingAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return a.startCmd, nil
}

func (a *recordingAdapter) NewParser() agentbridge.Parser { return a.parser }

func (a *recordingAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if a.translateFn != nil {
		return a.translateFn(raw)
	}
	return nil, nil, nil
}

func (a *recordingAdapter) BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	if a.inputFn != nil {
		return a.inputFn(cmd)
	}
	return nil, nil
}

func (a *recordingAdapter) BlockedArgs() []string { return a.blocked }
