package runtimeactor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
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

// TestRuntimeActorRunsCodexWithProtocolDriver is the M-9 integration
// success criterion: a TaskRequest goes through RuntimeActor.Submit →
// SessionActor → ProtocolDriver, walks the full handshake against a
// fake process, and reaches ResultCompleted WITHOUT any inline test
// driver writing to stdin.
func TestRuntimeActorRunsCodexWithProtocolDriver(t *testing.T) {
	a, p := startActor(t, Config{
		Adapters:      []agentbridge.Adapter{codexLikeAdapter{}},
		MaxConcurrent: 1,
	})

	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-codex", Provider: "codex-like", Prompt: "do the thing",
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	running := waitForRunning(t, p, 0, time.Second)

	// Drain events so the session loop doesn't block.
	go func() {
		for range h.Events() {
		}
	}()

	// "Codex server" side. It must observe stdin writes and emit the
	// matching JSON-RPC responses on stdout.
	driveCodexServer(t, running)

	// The session must reach ResultCompleted purely through the driver
	// path — no inline test-driver-side stdin writes here.
	select {
	case res := <-h.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("status: %s", res.Status)
		}
		if res.Output != "all done" {
			t.Fatalf("output: %q", res.Output)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session never reached ResultCompleted via ProtocolDriver")
	}

	// Slot must release.
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		s, _ := a.Status(context.Background())
		if s.RunningSessions == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("RunningSessions never returned to 0")
}

// driveCodexServer simulates the codex app-server: it reads stdin
// frames and emits the matching JSON-RPC responses on stdout.
func driveCodexServer(t *testing.T, r *process.FakeRunning) {
	t.Helper()
	go func() {
		seenInitialize := false
		seenThread := false
		seenTurn := false
		for {
			select {
			case frame, ok := <-r.StdinRecv():
				if !ok {
					return
				}
				s := string(frame)
				switch {
				case strings.Contains(s, `"method":"initialize"`):
					seenInitialize = true
					r.EmitStdout(jsonRPCResponse(1, map[string]any{"server": "codex-like"}))
				case strings.Contains(s, `"method":"thread/start"`):
					if !seenInitialize {
						return
					}
					seenThread = true
					r.EmitStdout(jsonRPCResponse(2, map[string]any{"thread_id": "th-1"}))
				case strings.Contains(s, `"method":"turn/start"`):
					if !seenThread {
						return
					}
					seenTurn = true
					r.EmitStdout(jsonRPCResponse(3, map[string]any{}))
					// Then emit turn_completed notification.
					r.EmitStdout(jsonRPCNotification("turn_completed", map[string]any{"output": "all done"}))
				default:
					_ = seenTurn // unused in failure paths
				}
			case <-time.After(3 * time.Second):
				return
			}
		}
	}()
}

func jsonRPCResponse(id int64, result map[string]any) []byte {
	m := map[string]any{"jsonrpc": "2.0", "id": id, "result": result}
	b, _ := json.Marshal(m)
	return append(b, '\n')
}

func jsonRPCNotification(method string, params map[string]any) []byte {
	m := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	b, _ := json.Marshal(m)
	return append(b, '\n')
}

// Compile-time guarantee: the test-only adapter satisfies the optional
// interface RuntimeActor will probe for.
var _ agentbridge.ProtocolDriverProvider = codexLikeAdapter{}
