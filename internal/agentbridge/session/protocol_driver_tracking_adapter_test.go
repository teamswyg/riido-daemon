package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// trackingAdapter records whether Translate was called so protocol-driver
// tests can prove whether the legacy adapter path was bypassed.
type trackingAdapter struct {
	translateCalls int
}

func (a *trackingAdapter) Name() string { return "tracking" }

func (a *trackingAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a *trackingAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, nil
}

func (a *trackingAdapter) NewParser() agentbridge.Parser { return &echoParser{} }

func (a *trackingAdapter) Translate(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	a.translateCalls++
	return nil, nil, nil
}

func (a *trackingAdapter) BlockedArgs() []string { return nil }

func (a *trackingAdapter) TranslateCalls() int { return a.translateCalls }

// echoParser turns every stdout chunk into one RawEvent of Type "chunk".
type echoParser struct{}

func (p *echoParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}

func (p *echoParser) FeedStderr(_ []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *echoParser) Close() ([]agentbridge.RawEvent, error)              { return nil, nil }
