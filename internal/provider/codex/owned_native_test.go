package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

// Native protocol smoke: real Codex process, isolated home and local model
// responses. No user chats, external tools, provider credentials or billing.
func TestIntegrationOwnedNativeFinalAnswer(t *testing.T) {
	req := nativeOwnedRequest(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	assertNativeOwnedFinalAnswer(t, runCodexIntegrationSession(t, ctx, req))
}

func TestIntegrationOwnedNativeRecoversMissingCompletion(t *testing.T) {
	req := nativeOwnedRequest(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	spawn, err := BuildStart(req, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	protocol, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	driver := &missingOwnedCompletion{protocolDriver: protocol.(*protocolDriver)}
	sess, err := session.Start(ctx, session.Config{
		TaskID: req.TaskID, RuntimeID: "codex", Adapter: integrationAdapter{}, Process: processexec.New(),
		Spawn: processCommandFromStart(spawn), Request: req, ProtocolDriver: driver,
		SemanticIdle: 2 * time.Second, HardTimeout: 12 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	go drainCodexIntegrationEvents(sess)
	result := <-sess.Result()
	if !driver.droppedCompletion {
		t.Fatal("fixture did not suppress the native completion notification")
	}
	assertNativeOwnedFinalAnswer(t, result)
}

// Suppress the transport notification while leaving the real app-server's
// history and recovery RPC untouched.
type missingOwnedCompletion struct {
	*protocolDriver
	droppedCompletion bool
}

func (d *missingOwnedCompletion) OnRaw(ctx context.Context, raw agentbridge.RawEvent, io agentbridge.ProtocolIO) ([]agentbridge.Event, []agentbridge.Command, error) {
	if raw.Type == "notification:turn/completed" {
		d.droppedCompletion = true
		return nil, nil, nil
	}
	return d.protocolDriver.OnRaw(ctx, raw, io)
}

func assertNativeOwnedFinalAnswer(t *testing.T, result agentbridge.Result) {
	t.Helper()
	if result.Status != agentbridge.ResultCompleted || strings.TrimSpace(result.Output) != "Parent final review" {
		t.Fatalf("native result=%+v", result)
	}
}

func nativeOwnedRequest(t *testing.T) agentbridge.StartRequest {
	t.Helper()
	exe := os.Getenv("RIIDO_CODEX_NATIVE_TEST_EXECUTABLE")
	if exe == "" {
		t.Skip("set RIIDO_CODEX_NATIVE_TEST_EXECUTABLE to a Codex binary")
	}
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, nativePhasedResponse())
	}))
	t.Cleanup(model.Close)
	home := t.TempDir()
	config := fmt.Sprintf("model_provider = 'fixture'\nmodel = 'gpt-5'\n[model_providers.fixture]\nname = 'fixture'\nbase_url = '%s/v1'\nwire_api = 'responses'\nrequires_openai_auth = false\n", model.URL)
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	return agentbridge.StartRequest{Executable: exe, TaskID: "owned-native", Cwd: home, Prompt: "Review fixture", Env: map[string]string{"CODEX_HOME": home}}
}

func nativePhasedResponse() string {
	var out strings.Builder
	emit := func(event string, data map[string]any) {
		data["type"] = event
		wire, _ := json.Marshal(data)
		fmt.Fprintf(&out, "event: %s\ndata: %s\n\n", event, wire)
	}
	emit("response.created", map[string]any{"response": map[string]any{"id": "fixture", "object": "response", "status": "in_progress", "output": []any{}}})
	var items []any
	for i, phase := range []string{"commentary", "final_answer"} {
		text := []string{"Reviewing before posting", "Parent final review"}[i]
		id := fmt.Sprintf("message-%d", i)
		item := map[string]any{"id": id, "type": "message", "role": "assistant", "phase": phase, "status": "in_progress", "content": []any{}}
		emit("response.output_item.added", map[string]any{"output_index": i, "item": item})
		emit("response.output_text.delta", map[string]any{"item_id": id, "output_index": i, "content_index": 0, "delta": text})
		item["status"] = "completed"
		item["content"] = []any{map[string]any{"type": "output_text", "text": text, "annotations": []any{}}}
		emit("response.output_item.done", map[string]any{"output_index": i, "item": item})
		items = append(items, item)
	}
	emit("response.completed", map[string]any{"response": map[string]any{"id": "fixture", "object": "response", "status": "completed", "output": items, "usage": map[string]any{"input_tokens": 1, "output_tokens": 1, "total_tokens": 2}}})
	return out.String()
}
