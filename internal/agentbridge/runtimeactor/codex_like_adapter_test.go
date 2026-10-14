package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// codexLikeAdapter is a minimal Codex-shaped fake adapter that implements both
// agentbridge.Adapter and agentbridge.ProtocolDriverProvider.
type codexLikeAdapter struct{}

func (codexLikeAdapter) Name() string { return "codex-like" }

func (codexLikeAdapter) Detect(
	_ context.Context,
	_ agentbridge.DetectEnv,
) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (codexLikeAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{Executable: "codex-like"}, nil
}

func (codexLikeAdapter) NewParser() agentbridge.Parser { return &lineJSONParser{} }

func (codexLikeAdapter) Translate(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return nil, nil, nil
}

func (codexLikeAdapter) BlockedArgs() []string { return nil }

func (codexLikeAdapter) NewProtocolDriver(_ agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return &codexLikeDriver{pending: map[int64]string{}}, nil
}

// Compile-time guarantee: the test-only adapter satisfies the optional
// interface RuntimeActor will probe for.
var _ agentbridge.ProtocolDriverProvider = codexLikeAdapter{}
