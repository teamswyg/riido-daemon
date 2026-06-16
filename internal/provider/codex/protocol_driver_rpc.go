package codex

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (d *protocolDriver) sendNotification(ctx context.Context, io agentbridge.ProtocolIO, method codexMethod, params map[string]any) error {
	frame := map[string]any{
		"jsonrpc": "2.0",
		"method":  string(method),
	}
	if params != nil {
		frame["params"] = params
	}
	return d.writeFrame(ctx, io, frame)
}

func (d *protocolDriver) writeFrame(ctx context.Context, io agentbridge.ProtocolIO, frame map[string]any) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("codex driver: marshal %v: %w", frame["method"], err)
	}
	b = append(b, '\n')
	return io.WriteStdin(ctx, b)
}

// rpcID extracts the JSON-RPC "id" field as int64. JSON numbers come
// back as float64; we accept int / int64 too for callers that build
// payloads directly without a JSON roundtrip (tests, internal helpers).
func rpcID(p map[string]any) (int64, bool) {
	switch v := p["id"].(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

// mapField returns p[key] as a map, or nil.
func mapField(p map[string]any, key string) map[string]any {
	if p == nil {
		return nil
	}
	m, _ := p[key].(map[string]any)
	return m
}

func threadIDFromResult(result map[string]any) string {
	if id := stringField(result, "thread_id"); id != "" {
		return id
	}
	thread := mapField(result, "thread")
	if id := stringField(thread, "id"); id != "" {
		return id
	}
	return stringField(thread, "sessionId")
}

func codexRPCErrorMessage(payload map[string]any) string {
	if msg := errMessage(payload); msg != "" {
		return msg
	}
	return "codex rpc error"
}

func codexNotificationErrorMessage(p map[string]any) string {
	if msg := stringField(p, "message"); msg != "" {
		return msg
	}
	if msg := stringField(p, "detail"); msg != "" {
		return msg
	}
	if errText := stringField(p, "error"); errText != "" {
		return errText
	}
	errMap := mapField(p, "error")
	if msg := stringField(errMap, "message"); msg != "" {
		return msg
	}
	return "codex runtime error"
}
