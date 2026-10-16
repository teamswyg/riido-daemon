package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const unsupportedCodexModelFixture = "The 'fixture-model' model is not supported when using Codex with a ChatGPT account."

func TestCodexModelRejectionIsBlockedWithoutInferringAuthOrTransport(t *testing.T) {
	for _, row := range []struct {
		name, message string
		params        map[string]any
		want          agentbridge.ResultStatus
	}{
		{name: "native message", message: unsupportedCodexModelFixture, want: agentbridge.ResultBlocked},
		{name: "HTTP JSON message", message: "400 Bad Request: {\"error\":{\"message\":\"" + unsupportedCodexModelFixture + "\"}}", want: agentbridge.ResultBlocked},
		{name: "structured unsupported", params: map[string]any{"code": "unsupported_model"}, want: agentbridge.ResultBlocked},
		{name: "nested model not found", params: map[string]any{"error": map[string]any{"code": "model_not_found"}}, want: agentbridge.ResultBlocked},
		{name: "quoted diagnostic in unrelated prose", message: "Earlier quote: " + unsupportedCodexModelFixture, want: agentbridge.ResultFailed},
		{name: "authentication", message: "401 Unauthorized; please log in", want: agentbridge.ResultFailed},
		{name: "generic invalid request", message: "400 Bad Request", want: agentbridge.ResultFailed},
		{name: "transport", message: "stream disconnected before completion", want: agentbridge.ResultFailed},
		{name: "unknown model message", message: "model connection failed", want: agentbridge.ResultFailed},
	} {
		t.Run(row.name, func(t *testing.T) {
			if got := codexFailureStatus(row.params, row.message); got != row.want {
				t.Fatalf("status=%s want=%s", got, row.want)
			}
		})
	}
}

func TestCodexUnsupportedModelTerminalIsBlockedButNativeRetryRemainsWarning(t *testing.T) {
	d, io := startedProtocolDriver(t)
	events, _, err := d.OnRaw(context.Background(), makeNotification("error", map[string]any{"message": unsupportedCodexModelFixture, "willRetry": true}), io)
	if err != nil || len(events) != 1 || events[0].Kind != agentbridge.EventWarning {
		t.Fatalf("native retry changed: %+v %v", events, err)
	}
	events, _, err = d.OnRaw(context.Background(), makeNotification("turn/completed", map[string]any{"turn": map[string]any{"status": "failed", "error": map[string]any{"message": unsupportedCodexModelFixture}}}), io)
	if err != nil || len(events) != 1 || events[0].Result.Status != agentbridge.ResultBlocked {
		t.Fatalf("model rejection=%+v %v", events, err)
	}
}

func TestCodexModelRejectionDoesNotRewriteSuccessfulNativeRetry(t *testing.T) {
	d, io := startedProtocolDriver(t)
	_, _, _ = d.OnRaw(context.Background(), makeNotification("error", map[string]any{"message": unsupportedCodexModelFixture}), io)
	events, _, err := d.OnRaw(context.Background(), makeNotification("turn/completed", map[string]any{"turn": map[string]any{"status": "completed"}, "output": "accepted native result"}), io)
	if err != nil || len(events) != 1 || events[0].Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("successful native status overwritten: %+v %v", events, err)
	}
}

func TestCodexUnsupportedModelRPCResponseIsBlocked(t *testing.T) {
	for _, errorBody := range []map[string]any{
		{"code": "model_not_found", "message": "Selected model is unavailable"},
		{"message": unsupportedCodexModelFixture},
	} {
		d, io := startedProtocolDriver(t)
		events, _, err := d.OnRaw(context.Background(), agentbridge.RawEvent{
			Source: agentbridge.RawSourceStdout, Type: "error",
			Payload: map[string]any{"id": int64(1), "error": errorBody},
		}, io)
		if err != nil || len(events) != 2 || events[1].Result.Status != agentbridge.ResultBlocked {
			t.Fatalf("RPC model rejection=%+v %v", events, err)
		}
	}
}
