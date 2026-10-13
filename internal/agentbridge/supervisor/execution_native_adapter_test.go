package supervisor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

type nativeCorrectionAdapter struct {
	requests chan agentbridge.StartRequest
	names    chan string
}

func (a *nativeCorrectionAdapter) Name() string { return codex.Name }
func (a *nativeCorrectionAdapter) Detect(ctx context.Context, env agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return codex.Detect(ctx, env)
}
func (a *nativeCorrectionAdapter) BuildStart(req agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	a.requests <- req
	return codex.BuildStart(req, codex.StartOptions{})
}
func (a *nativeCorrectionAdapter) NewParser() agentbridge.Parser { return codex.NewParser() }
func (a *nativeCorrectionAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return codex.Translate(raw)
}
func (a *nativeCorrectionAdapter) BlockedArgs() []string { return codex.BlockedArgs() }
func (a *nativeCorrectionAdapter) NewProtocolDriver(req agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	d, err := codex.NewProtocolDriver(req)
	return &nativeNameObserver{ProtocolDriver: d, names: a.names}, err
}

type nativeNameObserver struct {
	agentbridge.ProtocolDriver
	names chan string
}

func (d *nativeNameObserver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "notification:thread/name/updated" {
		if p, ok := raw.Payload["params"].(map[string]any); ok {
			if name, ok := p["threadName"].(string); ok {
				d.names <- name
			}
			if name, ok := p["name"].(string); ok {
				d.names <- name
			}
		}
	}
	return d.ProtocolDriver.OnRaw(ctx, raw, io)
}
