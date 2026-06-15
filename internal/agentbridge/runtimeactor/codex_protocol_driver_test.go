package runtimeactor

import (
	"context"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// codexLikeAdapter is a minimal Codex-shaped fake adapter that
// implements both agentbridge.Adapter AND agentbridge.ProtocolDriverProvider.
// It produces driver instances that walk an initialize / thread / turn
// handshake. We don't depend on the real codex package here to keep
// runtimeactor core's import graph clean — the RuntimeActor only knows
// about the optional agentbridge.ProtocolDriverProvider interface.
type codexLikeAdapter struct{}

func (codexLikeAdapter) Name() string { return "codex-like" }
func (codexLikeAdapter) Detect(_ context.Context, _ agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (codexLikeAdapter) BuildStart(_ agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{Executable: "codex-like"}, nil
}
func (codexLikeAdapter) NewParser() agentbridge.Parser { return &lineJSONParser{} }
func (codexLikeAdapter) Translate(_ agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	return nil, nil, nil // unused when ProtocolDriver is installed
}
func (codexLikeAdapter) BlockedArgs() []string { return nil }

// NewProtocolDriver is the optional hook RuntimeActor probes via
// type assertion against agentbridge.ProtocolDriverProvider.
func (codexLikeAdapter) NewProtocolDriver(_ agentbridge.StartRequest) (agentbridge.ProtocolDriver, error) {
	return &codexLikeDriver{pending: map[int64]string{}}, nil
}

// codexLikeDriver walks the same shape Codex would but with no JSON
// schema dependency — initialize → thread → turn → completed.
type codexLikeDriver struct {
	nextID  int64
	pending map[int64]string
	thread  string
}

func (d *codexLikeDriver) OnStart(ctx context.Context, io agentbridge.ProtocolIO) error {
	return d.send(ctx, io, "initialize", nil)
}

func (d *codexLikeDriver) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "response" {
		id, _ := payloadInt64(raw.Payload, "id")
		method := d.pending[id]
		delete(d.pending, id)
		switch method {
		case "initialize":
			return nil, nil, d.send(ctx, io, "thread/start", nil)
		case "thread/start":
			r, _ := raw.Payload["result"].(map[string]any)
			d.thread, _ = r["thread_id"].(string)
			return nil, nil, d.send(ctx, io, "turn/start", map[string]any{"thread_id": d.thread})
		case "turn/start":
			return []agentbridge.Event{{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}}, nil, nil
		}
		return nil, nil, nil
	}
	if raw.Type == "notification:turn_completed" {
		return []agentbridge.Event{{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultCompleted,
				Output: stringFromPayload(raw.Payload, "output"),
			},
		}}, nil, nil
	}
	return nil, nil, nil
}

func (d *codexLikeDriver) OnProcessExit(_ context.Context, _ agentbridge.ProcessExitStatus, _ agentbridge.ProtocolIO) ([]agentbridge.Event, error) {
	return nil, nil
}

func (d *codexLikeDriver) OnClose(_ context.Context, _ agentbridge.ProtocolIO) error { return nil }

func (d *codexLikeDriver) send(ctx context.Context, io agentbridge.ProtocolIO, method string, params map[string]any) error {
	d.nextID++
	d.pending[d.nextID] = method
	frame := map[string]any{"jsonrpc": "2.0", "id": d.nextID, "method": method}
	if params != nil {
		frame["params"] = params
	}
	b, _ := json.Marshal(frame)
	return io.WriteStdin(ctx, append(b, '\n'))
}

// lineJSONParser turns each '\n'-terminated JSON line into a RawEvent
// classified as response / notification:<method> / server_request:<method>.
type lineJSONParser struct{ buf []byte }

func (p *lineJSONParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.buf = append(p.buf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := -1
		for i, b := range p.buf {
			if b == '\n' {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		line := p.buf[:idx]
		p.buf = p.buf[idx+1:]
		if len(line) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			continue
		}
		raw := agentbridge.RawEvent{Source: agentbridge.RawSourceStdout, Payload: m, Bytes: line}
		method, _ := m["method"].(string)
		_, hasID := m["id"]
		switch {
		case method != "" && hasID:
			raw.Type = "server_request:" + method
		case method != "":
			raw.Type = "notification:" + method
		default:
			raw.Type = "response"
		}
		out = append(out, raw)
	}
	return out, nil
}
func (p *lineJSONParser) FeedStderr(_ []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *lineJSONParser) Close() ([]agentbridge.RawEvent, error)              { return nil, nil }

func payloadInt64(p map[string]any, key string) (int64, bool) {
	switch v := p[key].(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

func stringFromPayload(p map[string]any, key string) string {
	if p == nil {
		return ""
	}
	if params, ok := p["params"].(map[string]any); ok {
		if s, ok := params[key].(string); ok {
			return s
		}
	}
	s, _ := p[key].(string)
	return s
}
